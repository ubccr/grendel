// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-faster/jx"
	"github.com/go-fuego/fuego"
	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/pkg/client"
	"github.com/ubccr/grendel/pkg/model"
)

func testCertPEM(t *testing.T) (string, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test"}, NotBefore: time.Now(), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})), redfish.Fingerprint(der)
}

func TestSecureBootKey(t *testing.T) {
	cert, fp := testCertPEM(t)
	cert2, fp2 := testCertPEM(t)
	hash := strings.Repeat("AB", 32)

	req, op, err := secureBootKey(BmcSecureBootKeyBody{Database: "DB", Certificates: []string{cert, cert2}, Owner: "{26DC4851-195F-4AE1-9A19-FBF883BBB35E}"})
	if err != nil {
		t.Fatal(err)
	}
	if req.Database != "db" || len(req.Certificates) != 2 || req.Certificates[0].Fingerprint != fp || op != "bmc secureboot add db certificate "+fp+","+fp2 {
		t.Errorf("certificate: req %+v, operation %q", req, op)
	}
	if req.Owner != "26dc4851-195f-4ae1-9a19-fbf883bbb35e" {
		t.Errorf("owner = %s, want the canonical form", req.Owner)
	}

	req, op, err = secureBootKey(BmcSecureBootKeyBody{Database: "dbx", Signature: hash})
	if err != nil {
		t.Fatal(err)
	}
	if req.Signature != strings.ToLower(hash) || req.SignatureType != "EFI_CERT_SHA256_GUID" || !strings.Contains(op, "signature") {
		t.Errorf("signature: req %+v, operation %q", req, op)
	}

	refused := map[string]BmcSecureBootKeyBody{
		"no key":             {Database: "db"},
		"both":               {Database: "db", Certificates: []string{cert}, Signature: hash},
		"bad pem":            {Database: "db", Certificates: []string{cert, "-----BEGIN CERTIFICATE-----\nnope\n-----END CERTIFICATE-----"}},
		"default database":   {Database: "dbDefault", Certificates: []string{cert}},
		"unknown database":   {Database: "dbt", Certificates: []string{cert}},
		"bad owner":          {Database: "db", Certificates: []string{cert}, Owner: "me"},
		"not hex":            {Database: "dbx", Signature: strings.Repeat("zz", 32)},
		"wrong length":       {Database: "dbx", Signature: "abcd"},
		"unknown sig type":   {Database: "dbx", Signature: hash, SignatureType: "EFI_CERT_MD5_GUID"},
		"sha384 given 32 b.": {Database: "dbx", Signature: hash, SignatureType: "EFI_CERT_SHA384_GUID"},
	}
	for name, body := range refused {
		if _, _, err := secureBootKey(body); err == nil {
			t.Errorf("%s: expected a 400", name)
		}
	}
}

func TestSecureBootDelete(t *testing.T) {
	fp := strings.Repeat("ab", 32)
	colons := strings.TrimSuffix(strings.Repeat("AB:", 32), ":")

	req, _, err := secureBootDelete(BmcSecureBootDeleteBody{Database: "db", Fingerprint: colons})
	if err != nil {
		t.Fatal(err)
	}
	if req.Fingerprint != fp {
		t.Errorf("fingerprint = %s, want it normalized", req.Fingerprint)
	}

	refused := map[string]BmcSecureBootDeleteBody{
		"nothing":     {Database: "db"},
		"both":        {Database: "db", Fingerprint: fp, ID: "3"},
		"short":       {Database: "db", Fingerprint: fp[:16]},
		"not hex":     {Database: "db", Fingerprint: strings.Repeat("zz", 32)},
		"default":     {Database: "KEKDefault", ID: "1"},
		"no database": {ID: "1"},
	}
	for name, body := range refused {
		if _, _, err := secureBootDelete(body); err == nil {
			t.Errorf("%s: expected a 400", name)
		}
	}
}

func TestSecureBootReset(t *testing.T) {
	req, op, err := secureBootReset(BmcSecureBootResetBody{Type: "delete-pk"})
	if err != nil {
		t.Fatal(err)
	}
	if req.Type != schemas.DeletePKResetKeysType || req.Database != "" || op != "bmc secureboot reset delete-pk" {
		t.Errorf("req %+v, operation %q", req, op)
	}

	req, _, err = secureBootReset(BmcSecureBootResetBody{Type: "to-default", Database: "kek"})
	if err != nil {
		t.Fatal(err)
	}
	if req.Database != "KEK" {
		t.Errorf("database = %s", req.Database)
	}

	refused := map[string]BmcSecureBootResetBody{
		"no type":                {},
		"unknown type":           {Type: "default"},
		"delete-pk one database": {Type: "delete-pk", Database: "db"},
		"default database":       {Type: "to-default", Database: "dbDefault"},
	}
	for name, body := range refused {
		if _, _, err := secureBootReset(body); err == nil {
			t.Errorf("%s: expected a 400", name)
		}
	}
}

func TestGeneratedClientDecodesSecureBoot(t *testing.T) {
	on := true
	body, err := json.Marshal(model.BmcResponse[model.BmcSecureBoot]{RunID: "run", Results: []model.BmcResult[model.BmcSecureBoot]{
		{Node: "cpn-1", Success: true, Data: &model.BmcSecureBoot{Enabled: false, PendingEnabled: &on, Mode: "DeployedMode", Databases: []string{"PK", "db"}}},
		{Node: "cpn-2", Success: true, Data: &model.BmcSecureBoot{Databases: []string{}}},
		{Node: "cpn-3", Message: "boom"},
	}})
	if err != nil {
		t.Fatal(err)
	}

	var sb client.BmcResponseModelBmcSecureBoot
	if err := sb.Decode(jx.DecodeBytes(body)); err != nil {
		t.Fatalf("decoding the status %s: %v", body, err)
	}
	if v, ok := sb.Results.Value[0].Value.Data.Value.PendingEnabled.Get(); !ok || !v {
		t.Error("pending_enabled lost")
	}
	if _, ok := sb.Results.Value[1].Value.Data.Value.PendingEnabled.Get(); ok {
		t.Error("a null pending_enabled decoded as set")
	}

	body, err = json.Marshal(model.BmcResponse[model.BmcSecureBoot]{RunID: "run", Results: []model.BmcResult[model.BmcSecureBoot]{
		{Node: "cpn-1", Success: true, Data: &model.BmcSecureBoot{Databases: []string{"db"}, Keys: []model.BmcSecureBootDatabase{
			{ID: "db", Certificates: []model.BmcSecureBootCert{{ID: "1", Subject: "test", Fingerprint: "ab"}}, Signatures: []model.BmcSecureBootSignature{}},
		}}},
	}})
	if err != nil {
		t.Fatal(err)
	}

	var keys client.BmcResponseModelBmcSecureBoot
	if err := keys.Decode(jx.DecodeBytes(body)); err != nil {
		t.Fatalf("decoding the keys %s: %v", body, err)
	}
	if got := keys.Results.Value[0].Value.Data.Value.Keys[0].Certificates[0].Subject.Value; got != "test" {
		t.Errorf("subject = %s", got)
	}
}

// a body without enabled must be refused, since reading it as false would turn secure boot off on every node
func TestSetSecureBootRequiresEnabled(t *testing.T) {
	s := fuego.NewServer(fuego.WithoutStartupMessages())
	h := &Handler{}
	fuego.Post(s, "/sb", h.BmcSetSecureBoot)

	for _, body := range []string{`{}`, `{"enabled": null}`} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/sb?nodeset=cpn-1", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		s.Mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400: %s", body, rec.Code, rec.Body)
		}
	}
}
