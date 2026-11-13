// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testDER(t *testing.T, cn string) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: cn}, NotBefore: time.Now(), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return der
}

func TestReadCertificates(t *testing.T) {
	dir := t.TempDir()
	a, b := testDER(t, "a"), testDER(t, "b")

	bundle := filepath.Join(dir, "bundle.pem")
	data := append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: a}), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("x")})...)
	data = append(data, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: b})...)
	if err := os.WriteFile(bundle, data, 0o600); err != nil {
		t.Fatal(err)
	}
	certs, err := readCertificates(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if len(certs) != 2 || certs[0].Subject.CommonName != "a" || certs[1].Subject.CommonName != "b" {
		t.Errorf("read %d certificates from the bundle, want a then b, skipping the key", len(certs))
	}

	der := filepath.Join(dir, "one.der")
	if err := os.WriteFile(der, a, 0o600); err != nil {
		t.Fatal(err)
	}
	fp, err := deleteFingerprint(der)
	if err != nil {
		t.Fatal(err)
	}
	if fp != certFingerprint(certs[0]) {
		t.Errorf("a DER file and the same certificate in PEM give different fingerprints")
	}

	if fp, _ := deleteFingerprint("AB:CD"); fp != "AB:CD" {
		t.Errorf("an argument that is not a file is taken as a fingerprint, got %s", fp)
	}
	if _, err := deleteFingerprint(filepath.Join(dir, "missing.pem")); err == nil {
		t.Error("a missing file is reported as missing, not passed on as a fingerprint")
	}

	junk := filepath.Join(dir, "junk")
	if err := os.WriteFile(junk, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readCertificates(junk); err == nil {
		t.Error("expected an error for a file with no certificate")
	}
}

func TestHashKind(t *testing.T) {
	for typ, want := range map[string]string{"EFI_CERT_SHA256_GUID": "sha256", "EFI_CERT_X509_SHA256_GUID": "x509_sha256", "": "hash"} {
		if got := hashKind(typ); got != want {
			t.Errorf("hashKind(%q) = %s, want %s", typ, got, want)
		}
	}
}

func TestConfirmedWithoutATerminal(t *testing.T) {
	if ok, err := confirmed(true, "x"); !ok || err != nil {
		t.Errorf("--yes confirms without asking, got %t, %v", ok, err)
	}
	// go test runs with stdin that is not a terminal
	if ok, err := confirmed(false, "x"); ok || !errors.Is(err, errNoTerminal) {
		t.Errorf("with no terminal and no --yes, got %t, %v, want errNoTerminal", ok, err)
	}
}
