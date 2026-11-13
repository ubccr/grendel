// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/pkg/model"
)

// SecureBootRequest reads the secure boot state, and with Keys the contents of the databases too, over the one connection
type SecureBootRequest struct {
	Keys bool
	// Databases keeps only these databases, every non default one when empty
	Databases []string
	// Defaults also lists the read only factory default databases, such as dbDefault
	Defaults bool
}

// SecureBootCert is a PEM certificate with the fingerprint the caller worked out for it
type SecureBootCert struct {
	PEM         string
	Fingerprint string
}

// AddSecureBootKeyRequest enrolls certificates or a signature. The caller validates it: Database is a UEFI name Grendel writes to, every certificate parses, and the signature is lower case hex of the right length
type AddSecureBootKeyRequest struct {
	Database     string
	Certificates []SecureBootCert
	// Signature is hex encoded signature data, and SignatureType its UEFI type such as EFI_CERT_SHA256_GUID
	Signature     string
	SignatureType string
	Owner         string
}

// DeleteSecureBootKeyRequest names a key by Fingerprint, a certificate's SHA-256 or a signature's hash in lower case hex, which is the same on every node, or by ID, which is not
type DeleteSecureBootKeyRequest struct {
	Database    string
	Fingerprint string
	ID          string
}

// ResetSecureBootKeysRequest resets every database when Database is empty, or only that one. DeletePK only applies to every database
type ResetSecureBootKeysRequest struct {
	Database string
	Type     schemas.ResetKeysType
}

// SecureBoot reads the UEFI secure boot state of every target, and their keys when asked
func (s *Service) SecureBoot(ctx context.Context, targets []Target, req SecureBootRequest) []NodeResult[model.BmcSecureBoot] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*model.BmcSecureBoot, error) { return s.secureBootOne(ctx, t, req) })
}

// SetSecureBoot turns UEFI secure boot on or off on every target, from the next boot
func (s *Service) SetSecureBoot(ctx context.Context, targets []Target, enable bool) []NodeResult[Ack] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*Ack, error) { return s.setSecureBootOne(ctx, t, enable) })
}

// AddSecureBootKey enrolls certificates or a signature in a secure boot database on every target
func (s *Service) AddSecureBootKey(ctx context.Context, targets []Target, req AddSecureBootKeyRequest) []NodeResult[Ack] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*Ack, error) { return s.addSecureBootKeyOne(ctx, t, req) })
}

// DeleteSecureBootKey removes a certificate or signature from a secure boot database on every target
func (s *Service) DeleteSecureBootKey(ctx context.Context, targets []Target, req DeleteSecureBootKeyRequest) []NodeResult[Ack] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*Ack, error) { return s.deleteSecureBootKeyOne(ctx, t, req) })
}

// ResetSecureBootKeys resets or empties the secure boot databases of every target
func (s *Service) ResetSecureBootKeys(ctx context.Context, targets []Target, req ResetSecureBootKeysRequest) []NodeResult[Ack] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*Ack, error) { return s.resetSecureBootKeysOne(ctx, t, req) })
}

// secureBootDatabases are the databases Grendel writes to, keyed by lower case name. The UEFI dbr and dbt and every *Default copy are left read only
var secureBootDatabases = map[string]string{"pk": "PK", "kek": "KEK", "db": "db", "dbx": "dbx"}

// WritableDatabase returns the UEFI name of a database Grendel can change, matched without regard to case
func WritableDatabase(name string) (string, bool) {
	db, ok := secureBootDatabases[strings.ToLower(strings.TrimSpace(name))]
	return db, ok
}

// Fingerprint is the SHA-256 of a DER certificate as lower case hex, how Grendel names a certificate across nodes
func Fingerprint(der []byte) string {
	sum := sha256.Sum256(der)
	return hex.EncodeToString(sum[:])
}

// NormalizeFingerprint lower cases a hex fingerprint and drops the colons or spaces it is often printed with
func NormalizeFingerprint(fp string) string {
	return strings.ToLower(strings.NewReplacer(":", "", " ", "", "-", "").Replace(strings.TrimSpace(fp)))
}

// ParseCertificate reads the first PEM certificate in s
func ParseCertificate(s string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(s))
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errors.New("not a PEM encoded certificate")
	}
	return x509.ParseCertificate(block.Bytes)
}

// secureBootConn is a connection with the system's SecureBoot resource and where a change to it is staged
type secureBootConn struct {
	*conn
	sb       *schemas.SecureBoot
	settings schemas.Settings
}

func (s *Service) connectSecureBoot(ctx context.Context, t Target) (_ *secureBootConn, err error) {
	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	sys, err := firstSystem(c)
	if err != nil {
		return nil, err
	}
	sb, err := sys.SecureBoot()
	if err != nil {
		return nil, err
	}
	if sb == nil {
		return nil, errors.New("bmc exposes no secure boot resource")
	}
	settings, err := redfishSettings(c, sb.ODataID)
	if err != nil {
		return nil, err
	}
	return &secureBootConn{conn: c, sb: sb, settings: settings}, nil
}

// staged reports whether changes go to a separate settings object rather than the resource itself
func (b *secureBootConn) staged() bool {
	return b.settings.SettingsObject != "" && b.settings.SettingsObject != b.sb.ODataID
}

// pendingEnable reads a staged SecureBootEnable that differs from the current one, nil when there is none
func (b *secureBootConn) pendingEnable() (*bool, error) {
	if !b.staged() {
		return nil, nil
	}

	var raw struct {
		SecureBootEnable *bool
	}
	if err := b.getJSON(b.settings.SettingsObject, &raw); err != nil {
		return nil, fmt.Errorf("reading the secure boot settings object: %w", err)
	}
	if raw.SecureBootEnable == nil || *raw.SecureBootEnable == b.sb.SecureBootEnable {
		return nil, nil
	}
	return raw.SecureBootEnable, nil
}

func (b *secureBootConn) getJSON(uri string, out any) error {
	res, err := b.client.Get(uri)
	defer schemas.DeferredCleanupHTTPResponse(res)
	if err != nil {
		return err
	}
	return json.NewDecoder(res.Body).Decode(out)
}

// dbLink is one database as the collection lists it. Name is the last part of its uri, which the spec makes its Id and so its DatabaseId
type dbLink struct {
	name string
	uri  string
}

// databaseLinks reads the database collection without fetching each database, which gofish's SecureBootDatabases would do
func (b *secureBootConn) databaseLinks() ([]dbLink, error) {
	var root struct {
		SecureBootDatabases struct {
			ID string `json:"@odata.id"`
		}
	}
	if err := json.Unmarshal(b.sb.RawData, &root); err != nil {
		return nil, err
	}
	if root.SecureBootDatabases.ID == "" {
		return nil, errors.New("bmc exposes no secure boot databases")
	}

	var coll struct {
		Members []struct {
			ID string `json:"@odata.id"`
		}
	}
	if err := b.getJSON(root.SecureBootDatabases.ID, &coll); err != nil {
		return nil, fmt.Errorf("reading the secure boot databases: %w", err)
	}

	links := make([]dbLink, 0, len(coll.Members))
	for _, m := range coll.Members {
		name, err := url.PathUnescape(path.Base(strings.TrimRight(m.ID, "/")))
		if err != nil {
			name = path.Base(m.ID)
		}
		links = append(links, dbLink{name: name, uri: m.ID})
	}
	return links, nil
}

// database fetches one database by its UEFI name, matched without regard to case
func (b *secureBootConn) database(name string) (*schemas.SecureBootDatabase, error) {
	links, err := b.databaseLinks()
	if err != nil {
		return nil, err
	}
	for _, l := range links {
		if strings.EqualFold(l.name, name) {
			return schemas.GetSecureBootDatabase(b.client, l.uri)
		}
	}
	return nil, fmt.Errorf("bmc has no %s secure boot database", name)
}

func (s *Service) secureBootOne(ctx context.Context, t Target, req SecureBootRequest) (*model.BmcSecureBoot, error) {
	b, err := s.connectSecureBoot(ctx, t)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	pending, err := b.pendingEnable()
	if err != nil {
		return nil, err
	}

	out := &model.BmcSecureBoot{
		Enabled:        b.sb.SecureBootEnable,
		PendingEnabled: pending,
		Mode:           string(b.sb.SecureBootMode),
		CurrentBoot:    string(b.sb.SecureBootCurrentBoot),
		Databases:      []string{},
	}

	links, err := b.databaseLinks()
	if err != nil {
		// the state still says whether secure boot is on, so only a request for the keys fails without the databases
		if req.Keys {
			return nil, err
		}
		return out, nil
	}
	for _, l := range links {
		out.Databases = append(out.Databases, l.name)
	}
	if !req.Keys {
		return out, nil
	}

	out.Keys = []model.BmcSecureBootDatabase{}
	for _, l := range links {
		if len(req.Databases) > 0 {
			if !slices.ContainsFunc(req.Databases, func(n string) bool { return strings.EqualFold(n, l.name) }) {
				continue
			}
		} else if !req.Defaults && strings.HasSuffix(l.name, "Default") {
			continue
		}

		db, err := schemas.GetSecureBootDatabase(b.client, l.uri)
		if err != nil {
			return nil, fmt.Errorf("reading the %s database: %w", l.name, err)
		}
		listed, err := listDatabase(l.name, db)
		if err != nil {
			return nil, fmt.Errorf("reading the %s database: %w", l.name, err)
		}
		out.Keys = append(out.Keys, listed)
	}
	return out, nil
}

func (s *Service) setSecureBootOne(ctx context.Context, t Target, enable bool) (*Ack, error) {
	b, err := s.connectSecureBoot(ctx, t)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	word := "disabled"
	if enable {
		word = "enabled"
	}

	pending, err := b.pendingEnable()
	if err != nil {
		return nil, err
	}
	next := b.sb.SecureBootEnable
	if pending != nil {
		next = *pending
	}
	if next == enable {
		return doneWith("already %s, nothing sent", word)
	}

	if b.staged() {
		res, err := b.client.Patch(b.settings.SettingsObject, map[string]any{"SecureBootEnable": enable})
		schemas.DeferredCleanupHTTPResponse(res)
		if err != nil {
			return nil, err
		}
	} else {
		b.sb.SecureBootEnable = enable
		if err := b.sb.Update(); err != nil {
			return nil, err
		}
	}

	return doneWith("secure boot %s from the next boot", word)
}

func listDatabase(name string, db *schemas.SecureBootDatabase) (model.BmcSecureBootDatabase, error) {
	out := model.BmcSecureBootDatabase{
		ID:           name,
		Certificates: []model.BmcSecureBootCert{},
		Signatures:   []model.BmcSecureBootSignature{},
	}

	certs, err := db.Certificates()
	if err != nil {
		return out, err
	}
	for _, c := range certs {
		out.Certificates = append(out.Certificates, certEntry(c))
	}

	sigs, err := db.Signatures()
	if err != nil {
		return out, err
	}
	for _, sig := range sigs {
		out.Signatures = append(out.Signatures, model.BmcSecureBootSignature{
			ID:    sig.ID,
			Type:  sig.SignatureType,
			Hash:  strings.ToLower(sig.SignatureString),
			Owner: sig.UefiSignatureOwner,
		})
	}
	return out, nil
}

// certEntry prefers what the certificate itself says, since bmcs fill the parsed fields unevenly
func certEntry(c *schemas.Certificate) model.BmcSecureBootCert {
	parsed, _ := ParseCertificate(c.CertificateString)
	out := model.BmcSecureBootCert{
		ID:          c.ID,
		Subject:     c.Subject.CommonName,
		Issuer:      c.Issuer.CommonName,
		NotBefore:   c.ValidNotBefore,
		NotAfter:    c.ValidNotAfter,
		Fingerprint: certFingerprint(c, parsed),
		Owner:       c.UefiSignatureOwner,
	}
	if parsed != nil {
		out.Subject = parsed.Subject.CommonName
		out.Issuer = parsed.Issuer.CommonName
		out.NotBefore = parsed.NotBefore.UTC().Format(time.RFC3339)
		out.NotAfter = parsed.NotAfter.UTC().Format(time.RFC3339)
	}
	if out.Subject == "" && c.Subject.Organization != "" {
		out.Subject = c.Subject.Organization
	}
	return out
}

// certFingerprint is the certificate's SHA-256, worked out from the certificate when the bmc includes it, otherwise the bmc's own fingerprint when that is a SHA-256. parsed is the certificate when it was already read, nil when it was not or could not be
func certFingerprint(c *schemas.Certificate, parsed *x509.Certificate) string {
	if parsed != nil {
		return Fingerprint(parsed.Raw)
	}
	alg := strings.ToUpper(strings.NewReplacer("-", "", "_", "").Replace(c.FingerprintHashAlgorithm))
	if strings.HasSuffix(alg, "SHA256") && c.Fingerprint != "" {
		return NormalizeFingerprint(c.Fingerprint)
	}
	return ""
}

func fingerprintOf(c *schemas.Certificate) string {
	parsed, _ := ParseCertificate(c.CertificateString)
	return certFingerprint(c, parsed)
}

func (s *Service) addSecureBootKeyOne(ctx context.Context, t Target, req AddSecureBootKeyRequest) (*Ack, error) {
	b, err := s.connectSecureBoot(ctx, t)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	db, err := b.database(req.Database)
	if err != nil {
		return nil, err
	}

	if len(req.Certificates) == 0 {
		return addSignature(db, req)
	}

	existing, err := db.Certificates()
	if err != nil {
		return nil, err
	}
	enrolled := map[string]string{}
	for _, c := range existing {
		if fp := fingerprintOf(c); fp != "" {
			enrolled[fp] = c.ID
		}
	}

	var added []string
	skipped := 0
	for i, cert := range req.Certificates {
		if _, ok := enrolled[cert.Fingerprint]; ok {
			skipped++
			continue
		}
		created, err := db.AddCertificate(cert.PEM, schemas.PEMCertificateType, req.Owner)
		if err != nil {
			if len(added) == 0 {
				return nil, err
			}
			return nil, fmt.Errorf("enrolled %d in %s, then failed on certificate %d: %s", len(added), req.Database, i+1, errorMessage(err))
		}
		id := "?"
		if created != nil && created.ID != "" {
			id = created.ID
		}
		added = append(added, id)
	}

	if len(added) == 0 {
		return doneWith("already enrolled in %s, nothing sent", req.Database)
	}
	note := fmt.Sprintf("enrolled in %s as %s", req.Database, strings.Join(added, ", "))
	if slices.Contains(added, "?") {
		note += ", list the keys to confirm"
	}
	if skipped > 0 {
		note += fmt.Sprintf(", %d already enrolled", skipped)
	}
	return doneWith("%s", note)
}

func addSignature(db *schemas.SecureBootDatabase, req AddSecureBootKeyRequest) (*Ack, error) {
	existing, err := db.Signatures()
	if err != nil {
		return nil, err
	}
	for _, sig := range existing {
		if strings.EqualFold(sig.SignatureString, req.Signature) && strings.EqualFold(sig.SignatureType, req.SignatureType) {
			return doneWith("already enrolled in %s as %s, nothing sent", req.Database, sig.ID)
		}
	}

	created, err := db.AddSignature(req.Signature, schemas.UEFISignatureTypeRegistry, req.SignatureType, req.Owner)
	if err != nil {
		return nil, err
	}
	if created != nil && created.ID != "" {
		return doneWith("enrolled in %s as %s", req.Database, created.ID)
	}
	return doneWith("enrolled in %s, list the keys to confirm", req.Database)
}

func (s *Service) deleteSecureBootKeyOne(ctx context.Context, t Target, req DeleteSecureBootKeyRequest) (*Ack, error) {
	b, err := s.connectSecureBoot(ctx, t)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	db, err := b.database(req.Database)
	if err != nil {
		return nil, err
	}

	var uris, ids []string
	certs, err := db.Certificates()
	if err != nil {
		return nil, err
	}
	for _, c := range certs {
		if (req.ID != "" && c.ID == req.ID) || (req.Fingerprint != "" && fingerprintOf(c) == req.Fingerprint) {
			uris, ids = append(uris, c.ODataID), append(ids, c.ID)
		}
	}

	sigs, err := db.Signatures()
	if err != nil {
		return nil, err
	}
	for _, sig := range sigs {
		if (req.ID != "" && sig.ID == req.ID) || (req.Fingerprint != "" && strings.EqualFold(sig.SignatureString, req.Fingerprint)) {
			uris, ids = append(uris, sig.ODataID), append(ids, sig.ID)
		}
	}

	if len(uris) == 0 {
		return doneWith("not enrolled in %s, nothing sent", req.Database)
	}

	for i, uri := range uris {
		res, err := b.client.Delete(uri)
		schemas.DeferredCleanupHTTPResponse(res)
		if err != nil {
			if i == 0 {
				return nil, err
			}
			return nil, fmt.Errorf("removed %s from %s, then failed on %s: %s", strings.Join(ids[:i], ", "), req.Database, ids[i], errorMessage(err))
		}
	}
	return doneWith("removed %s from %s", strings.Join(ids, ", "), req.Database)
}

// hasResetAction reports whether the SecureBoot resource declares ResetKeys. Without it gofish would post to an empty target, which reaches the service root and fails with a message that says nothing about why
func (b *secureBootConn) hasResetAction() bool {
	var raw struct {
		Actions map[string]struct {
			Target string `json:"target"`
		}
	}
	if err := json.Unmarshal(b.sb.RawData, &raw); err != nil {
		return false
	}
	return raw.Actions["#SecureBoot.ResetKeys"].Target != ""
}

func (s *Service) resetSecureBootKeysOne(ctx context.Context, t Target, req ResetSecureBootKeysRequest) (*Ack, error) {
	b, err := s.connectSecureBoot(ctx, t)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	var (
		task  *schemas.TaskMonitorInfo
		scope = "every database"
	)
	if req.Database == "" {
		if !b.hasResetAction() {
			return nil, errors.New("bmc does not support resetting secure boot keys")
		}
		task, err = b.sb.ResetKeys(req.Type)
	} else {
		db, dbErr := b.database(req.Database)
		if dbErr != nil {
			return nil, dbErr
		}
		scope = req.Database
		task, err = db.ResetKeys(schemas.SecureBootDatabaseResetKeysType(req.Type))
	}
	if err != nil {
		return nil, err
	}

	if task != nil && task.TaskMonitor != "" {
		return doneWith("%s on %s, task %s", req.Type, scope, path.Base(task.TaskMonitor))
	}
	return doneWith("%s on %s", req.Type, scope)
}
