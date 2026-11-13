// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stmcginnis/gofish/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sbRoot = "/redfish/v1/Systems/System.Embedded.1/SecureBoot"

func sbColl(db, kind string) string { return sbRoot + "/SecureBootDatabases/" + db + "/" + kind }

// testCert makes a self signed certificate, returning it as the caller hands it to the redfish layer
func testCert(t *testing.T, cn string) SecureBootCert {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		NotAfter:     time.Date(2036, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	return SecureBootCert{PEM: string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})), Fingerprint: Fingerprint(der)}
}

// secureBootFixture serves an iDRAC style SecureBoot resource with PK, KEK, db, dbx and dbDefault databases, each with empty writable collections. enabled is SecureBootEnable, and staged, when not empty, is the SecureBootEnable held in a separate settings object. Actions and patches are accepted and recorded in f.writes
func secureBootFixture(t *testing.T, enabled bool, staged string) *fixture {
	t.Helper()

	f := dellFixture()
	f.tree["/redfish/v1/Systems/System.Embedded.1"] = strings.Replace(systemDell, `"Boot": {`, `"SecureBoot": {"@odata.id": "`+sbRoot+`"}, "Boot": {`, 1)

	settings := ""
	if staged != "" {
		settings = `, "@Redfish.Settings": {"SettingsObject": {"@odata.id": "` + sbRoot + `/Settings"}}`
		f.tree[sbRoot+"/Settings"] = `{"@odata.id": "` + sbRoot + `/Settings", "SecureBootEnable": ` + staged + `}`
	}
	f.tree[sbRoot] = fmt.Sprintf(`{
  "@odata.id": "%s", "Id": "SecureBoot", "SecureBootEnable": %t, "SecureBootMode": "DeployedMode", "SecureBootCurrentBoot": "Disabled",
  "SecureBootDatabases": {"@odata.id": "%s/SecureBootDatabases"},
  "Actions": {"#SecureBoot.ResetKeys": {"target": "%s/Actions/SecureBoot.ResetKeys"}}%s
}`, sbRoot, enabled, sbRoot, sbRoot, settings)

	var members []string
	for _, id := range []string{"PK", "KEK", "db", "dbx", "dbDefault"} {
		uri := sbRoot + "/SecureBootDatabases/" + id
		members = append(members, `{"@odata.id": "`+uri+`"}`)
		f.tree[uri] = fmt.Sprintf(`{"@odata.id": "%s", "Id": "%s", "DatabaseId": "%s",
  "Certificates": {"@odata.id": "%s/Certificates"}, "Signatures": {"@odata.id": "%s/Signatures"},
  "Actions": {"#SecureBootDatabase.ResetKeys": {"target": "%s/Actions/SecureBootDatabase.ResetKeys"}}}`, uri, id, id, uri, uri, uri)
		f.collections[uri+"/Certificates"] = map[string]string{}
		f.collections[uri+"/Signatures"] = map[string]string{}
	}
	f.tree[sbRoot+"/SecureBootDatabases"] = `{"@odata.id": "` + sbRoot + `/SecureBootDatabases", "Members": [` + strings.Join(members, ",") + `]}`

	f.handler = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method != http.MethodPatch && !strings.Contains(r.URL.Path, "/Actions/") {
			return false
		}
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return f
}

// enroll puts a certificate in the fixture under the given id, as if it was there before the test
func enroll(f *fixture, db, id string, cert SecureBootCert) {
	f.member(sbColl(db, "Certificates"), id, map[string]any{"CertificateString": cert.PEM, "CertificateType": "PEM"})
}

func deletes(f *fixture) []string {
	var out []string
	for _, w := range f.writes {
		if w.Method == http.MethodDelete {
			out = append(out, w.Path)
		}
	}
	return out
}

func sbTarget(t *testing.T, f *fixture) []Target {
	return []Target{{Name: "cpn-a01-01", Addr: f.serve(t).URL}}
}

func TestSecureBootStatus(t *testing.T) {
	f := secureBootFixture(t, false, "")

	results := testService(nil).SecureBoot(context.Background(), sbTarget(t, f), SecureBootRequest{})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)

	sb := results[0].Data
	assert.False(t, sb.Enabled)
	assert.Nil(t, sb.PendingEnabled)
	assert.Equal(t, "DeployedMode", sb.Mode)
	assert.Equal(t, "Disabled", sb.CurrentBoot)
	assert.Equal(t, []string{"PK", "KEK", "db", "dbx", "dbDefault"}, sb.Databases)
	assert.Nil(t, sb.Keys)
	assert.Zero(t, f.hits[sbRoot+"/SecureBootDatabases/db"], "the names come from the collection, without fetching each database")
}

func TestSecureBootStatusPending(t *testing.T) {
	f := secureBootFixture(t, false, "true")

	results := testService(nil).SecureBoot(context.Background(), sbTarget(t, f), SecureBootRequest{})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	require.NotNil(t, results[0].Data.PendingEnabled)
	assert.True(t, *results[0].Data.PendingEnabled)
}

func TestSecureBootMissing(t *testing.T) {
	f := dellFixture()

	results := testService(nil).SecureBoot(context.Background(), sbTarget(t, f), SecureBootRequest{})
	assert.False(t, results[0].Success)
	assert.Equal(t, "bmc exposes no secure boot resource", results[0].Message)
}

func TestSetSecureBoot(t *testing.T) {
	f := secureBootFixture(t, false, "")

	results := testService(nil).SetSecureBoot(context.Background(), sbTarget(t, f), true)
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "secure boot enabled from the next boot", results[0].Message)
	assert.Equal(t, []map[string]any{{"SecureBootEnable": true}}, f.written(http.MethodPatch, sbRoot))
}

func TestSetSecureBootStaged(t *testing.T) {
	f := secureBootFixture(t, true, "true")

	results := testService(nil).SetSecureBoot(context.Background(), sbTarget(t, f), false)
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, []map[string]any{{"SecureBootEnable": false}}, f.written(http.MethodPatch, sbRoot+"/Settings"), "a bmc with a settings object takes the change there")
	assert.Empty(t, f.written(http.MethodPatch, sbRoot))
}

func TestSetSecureBootAlreadySet(t *testing.T) {
	// disabled now, but enabling is already staged
	f := secureBootFixture(t, false, "true")

	results := testService(nil).SetSecureBoot(context.Background(), sbTarget(t, f), true)
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "already enabled, nothing sent", results[0].Message)
	assert.Empty(t, f.writes)
}

func TestSecureBootKeys(t *testing.T) {
	f := secureBootFixture(t, true, "")
	cert := testCert(t, "Grendel Test DB")
	enroll(f, "db", "7", cert)
	enroll(f, "dbDefault", "1", cert)

	svc := testService(nil)
	results := svc.SecureBoot(context.Background(), sbTarget(t, f), SecureBootRequest{Keys: true})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)

	var ids []string
	for _, db := range results[0].Data.Keys {
		ids = append(ids, db.ID)
	}
	assert.Equal(t, []string{"PK", "KEK", "db", "dbx"}, ids, "defaults are left out unless asked for")
	assert.Zero(t, f.hits[sbRoot+"/SecureBootDatabases/dbDefault"], "a database left out is not fetched")

	db := results[0].Data.Keys[2]
	require.Len(t, db.Certificates, 1)
	assert.Equal(t, "7", db.Certificates[0].ID)
	assert.Equal(t, "Grendel Test DB", db.Certificates[0].Subject)
	assert.Equal(t, "2036-01-01T00:00:00Z", db.Certificates[0].NotAfter)
	assert.Equal(t, cert.Fingerprint, db.Certificates[0].Fingerprint)

	results = svc.SecureBoot(context.Background(), sbTarget(t, f), SecureBootRequest{Keys: true, Databases: []string{"DBDEFAULT"}})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	require.Len(t, results[0].Data.Keys, 1)
	assert.Equal(t, "dbDefault", results[0].Data.Keys[0].ID)
}

func TestAddSecureBootCerts(t *testing.T) {
	f := secureBootFixture(t, true, "")
	a, b := testCert(t, "a"), testCert(t, "b")
	enroll(f, "db", "7", a)
	targets := sbTarget(t, f)
	coll := sbColl("db", "Certificates")

	results := testService(nil).AddSecureBootKey(context.Background(), targets, AddSecureBootKeyRequest{Database: "db", Certificates: []SecureBootCert{a, b}, Owner: "26dc4851-195f-4ae1-9a19-fbf883bbb35e"})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "enrolled in db as 1, 1 already enrolled", results[0].Message, "the fixture numbers its own posts from 1")
	posts := f.written(http.MethodPost, coll)
	require.Len(t, posts, 1, "only the certificate not already there is sent")
	assert.Equal(t, b.PEM, posts[0]["CertificateString"])
	assert.Equal(t, "26dc4851-195f-4ae1-9a19-fbf883bbb35e", posts[0]["UefiSignatureOwner"])

	results = testService(nil).AddSecureBootKey(context.Background(), targets, AddSecureBootKeyRequest{Database: "db", Certificates: []SecureBootCert{a, b}})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "already enrolled in db, nothing sent", results[0].Message)
	assert.Len(t, f.written(http.MethodPost, coll), 1)
}

func TestAddSecureBootSignature(t *testing.T) {
	f := secureBootFixture(t, true, "")
	hash := strings.Repeat("ab", 32)

	results := testService(nil).AddSecureBootKey(context.Background(), sbTarget(t, f), AddSecureBootKeyRequest{Database: "dbx", Signature: hash, SignatureType: "EFI_CERT_SHA256_GUID"})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, []map[string]any{{"SignatureString": hash, "SignatureType": "EFI_CERT_SHA256_GUID", "SignatureTypeRegistry": "UEFI"}}, f.written(http.MethodPost, sbColl("dbx", "Signatures")))
}

func TestDeleteSecureBootKeyByFingerprint(t *testing.T) {
	cert, other := testCert(t, "Grendel Test DB"), testCert(t, "Other")

	// the same certificate under a different id on each node
	f1 := secureBootFixture(t, true, "")
	enroll(f1, "db", "3", cert)
	enroll(f1, "db", "4", other)
	f2 := secureBootFixture(t, true, "")
	enroll(f2, "db", "9", cert)

	targets := []Target{{Name: "cpn-a01-01", Addr: f1.serve(t).URL}, {Name: "cpn-a01-02", Addr: f2.serve(t).URL}}
	results := testService(nil).DeleteSecureBootKey(context.Background(), targets, DeleteSecureBootKeyRequest{Database: "db", Fingerprint: cert.Fingerprint})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	require.True(t, results[1].Success, "message was: %s", results[1].Message)
	assert.Equal(t, "removed 3 from db", results[0].Message)
	assert.Equal(t, "removed 9 from db", results[1].Message)
	assert.Equal(t, []string{sbColl("db", "Certificates") + "/3"}, deletes(f1))
	assert.Equal(t, []string{sbColl("db", "Certificates") + "/9"}, deletes(f2))
}

func TestDeleteSecureBootKeyNotEnrolled(t *testing.T) {
	f := secureBootFixture(t, true, "")

	results := testService(nil).DeleteSecureBootKey(context.Background(), sbTarget(t, f), DeleteSecureBootKeyRequest{Database: "db", ID: "42"})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "not enrolled in db, nothing sent", results[0].Message)
	assert.Empty(t, deletes(f))
}

func TestResetSecureBootKeys(t *testing.T) {
	f := secureBootFixture(t, true, "")
	targets := sbTarget(t, f)
	svc := testService(nil)

	results := svc.ResetSecureBootKeys(context.Background(), targets, ResetSecureBootKeysRequest{Type: schemas.ResetAllKeysToDefaultResetKeysType})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "ResetAllKeysToDefault on every database", results[0].Message)
	assert.Equal(t, []map[string]any{{"ResetKeysType": "ResetAllKeysToDefault"}}, f.written(http.MethodPost, sbRoot+"/Actions/SecureBoot.ResetKeys"))

	results = svc.ResetSecureBootKeys(context.Background(), targets, ResetSecureBootKeysRequest{Database: "dbx", Type: schemas.DeleteAllKeysResetKeysType})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "DeleteAllKeys on dbx", results[0].Message)
	assert.Equal(t, []map[string]any{{"ResetKeysType": "DeleteAllKeys"}}, f.written(http.MethodPost, sbRoot+"/SecureBootDatabases/dbx/Actions/SecureBootDatabase.ResetKeys"))
}

func TestResetSecureBootKeysWithoutTheAction(t *testing.T) {
	f := secureBootFixture(t, true, "")
	f.tree[sbRoot] = strings.Replace(f.tree[sbRoot], `"#SecureBoot.ResetKeys"`, `"#SecureBoot.Other"`, 1)

	results := testService(nil).ResetSecureBootKeys(context.Background(), sbTarget(t, f), ResetSecureBootKeysRequest{Type: schemas.DeletePKResetKeysType})
	assert.False(t, results[0].Success)
	assert.Equal(t, "bmc does not support resetting secure boot keys", results[0].Message)
	assert.Empty(t, f.writes, "nothing is posted to an action the bmc does not declare")
}
