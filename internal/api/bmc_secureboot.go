// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/pkg/model"
)

// BmcSecureBootBody turns secure boot on or off. Enabled is a pointer so a body that leaves it out is refused, rather than read as false and turning secure boot off
type BmcSecureBootBody struct {
	Enabled *bool `json:"enabled" validate:"required" oai3:"notnull" description:"Whether secure boot is in effect from the next boot"`
}

// BmcSecureBootKeyBody is the keys to enroll, either certificates or one signature
type BmcSecureBootKeyBody struct {
	Database      string   `json:"database" description:"PK, KEK, db or dbx" example:"db"`
	Certificates  []string `json:"certificates" description:"PEM encoded X.509 certificates, one per entry, enrolled in order. Mutually exclusive with signature"`
	Signature     string   `json:"signature" description:"Hex encoded signature data, such as the SHA-256 of a binary to revoke in dbx. Mutually exclusive with certificates"`
	SignatureType string   `json:"signature_type" enum:"EFI_CERT_SHA256_GUID,EFI_CERT_SHA384_GUID,EFI_CERT_SHA512_GUID,EFI_CERT_X509_SHA256_GUID" description:"UEFI type of the signature. Defaults to EFI_CERT_SHA256_GUID"`
	Owner         string   `json:"owner" description:"UEFI signature owner GUID. Empty lets the bmc choose"`
}

// BmcSecureBootDeleteBody names one key to remove, by a fingerprint that is the same on every node, or an id that is not
type BmcSecureBootDeleteBody struct {
	Database    string `json:"database" description:"PK, KEK, db or dbx" example:"db"`
	Fingerprint string `json:"fingerprint" description:"SHA-256 of a certificate, or the hash of a signature, in hex with or without colons. Mutually exclusive with id"`
	ID          string `json:"id" description:"Id of the certificate or signature on the bmc. Ids differ between nodes, so this suits a single node. Mutually exclusive with fingerprint"`
}

// BmcSecureBootResetBody resets the secure boot databases
type BmcSecureBootResetBody struct {
	Database string `json:"database" description:"Reset only this database, PK, KEK, db or dbx. Empty resets every database"`
	Type     string `json:"type" enum:"to-default,delete-all,delete-pk" description:"to-default restores the factory keys, delete-all removes every key, which leaves the system in setup mode, and delete-pk removes only the platform key. delete-pk applies to every database, not one"`
}

// BmcSecureBoot reads the secure boot state of node(s), and with keys the contents of their databases, over one connection to each bmc
func (h *Handler) BmcSecureBoot(c fuego.ContextNoBody) (model.BmcResponse[model.BmcSecureBoot], error) {
	var req redfish.SecureBootRequest
	for _, name := range strings.Split(c.QueryParam("databases"), ",") {
		if name = strings.TrimSpace(name); name != "" {
			req.Databases = append(req.Databases, name)
		}
	}
	var err error
	if req.Keys, err = queryBool(c, "keys"); err != nil {
		return model.BmcResponse[model.BmcSecureBoot]{}, err
	}
	if req.Defaults, err = queryBool(c, "defaults"); err != nil {
		return model.BmcResponse[model.BmcSecureBoot]{}, err
	}

	return runBmc(h, c, "bmc secureboot", false,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[model.BmcSecureBoot] {
			return h.Redfish.SecureBoot(ctx, t, req)
		},
		func(data *model.BmcSecureBoot) *model.BmcSecureBoot { return data },
	)
}

// queryBool reads an optional boolean query parameter, false when it is absent
func queryBool(c fuego.ContextNoBody, name string) (bool, error) {
	if c.QueryParam(name) == "" {
		return false, nil
	}
	v, err := c.QueryParamBoolErr(name)
	if err != nil {
		return false, badRequest(err, name+" must be true or false")
	}
	return v, nil
}

// BmcSetSecureBoot turns secure boot on or off on node(s), from their next boot
func (h *Handler) BmcSetSecureBoot(c fuego.ContextWithBody[BmcSecureBootBody]) (model.BmcResponse[model.BmcAck], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, badRequest(err, "enabled must be true or false")
	}
	enable := *body.Enabled

	operation := "bmc secureboot disable"
	if enable {
		operation = "bmc secureboot enable"
	}
	return runAck(h, c, operation, func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.SetSecureBoot(ctx, t, enable)
	})
}

// BmcAddSecureBootKey enrolls certificates or a signature in a secure boot database on node(s)
func (h *Handler) BmcAddSecureBootKey(c fuego.ContextWithBody[BmcSecureBootKeyBody]) (model.BmcResponse[model.BmcAck], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, badRequest(err, "failed to parse the secure boot key")
	}

	req, operation, err := secureBootKey(body)
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}

	return runAck(h, c, operation, func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.AddSecureBootKey(ctx, t, req)
	})
}

// BmcDeleteSecureBootKey removes a certificate or signature from a secure boot database on node(s)
func (h *Handler) BmcDeleteSecureBootKey(c fuego.ContextWithBody[BmcSecureBootDeleteBody]) (model.BmcResponse[model.BmcAck], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, badRequest(err, "failed to parse the secure boot key")
	}

	req, operation, err := secureBootDelete(body)
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}

	return runAck(h, c, operation, func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.DeleteSecureBootKey(ctx, t, req)
	})
}

// BmcResetSecureBootKeys resets or empties the secure boot databases of node(s)
func (h *Handler) BmcResetSecureBootKeys(c fuego.ContextWithBody[BmcSecureBootResetBody]) (model.BmcResponse[model.BmcAck], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, badRequest(err, "failed to parse the secure boot reset")
	}

	req, operation, err := secureBootReset(body)
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}

	return runAck(h, c, operation, func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.ResetSecureBootKeys(ctx, t, req)
	})
}

// signatureLengths is the size in bytes of each signature type's data. An X509_SHA256 entry is a hash followed by a 16 byte revocation time. The type names are also the enum on BmcSecureBootKeyBody.SignatureType
var signatureLengths = map[string]int{
	"EFI_CERT_SHA256_GUID":      32,
	"EFI_CERT_SHA384_GUID":      48,
	"EFI_CERT_SHA512_GUID":      64,
	"EFI_CERT_X509_SHA256_GUID": 48,
}

func writableDatabase(name string) (string, error) {
	db, ok := redfish.WritableDatabase(name)
	if !ok {
		return "", badRequest(fmt.Errorf("unknown database %q", name), "database must be PK, KEK, db or dbx")
	}
	return db, nil
}

// secureBootKey validates the keys before any node is contacted, and hands the redfish layer what it would otherwise work out again on every node. The event names each key by its fingerprint, never its contents
func secureBootKey(body BmcSecureBootKeyBody) (redfish.AddSecureBootKeyRequest, string, error) {
	var req redfish.AddSecureBootKeyRequest

	db, err := writableDatabase(body.Database)
	if err != nil {
		return req, "", err
	}
	req.Database = db

	if body.Owner != "" {
		// uuid accepts the braced and urn forms too, and the bmc gets the canonical one
		owner, err := uuid.Parse(strings.TrimSpace(body.Owner))
		if err != nil {
			return req, "", badRequest(err, "owner must be a GUID such as 26dc4851-195f-4ae1-9a19-fbf883bbb35e")
		}
		req.Owner = owner.String()
	}

	var certs []string
	for _, c := range body.Certificates {
		if c = strings.TrimSpace(c); c != "" {
			certs = append(certs, c)
		}
	}
	sig := strings.TrimSpace(body.Signature)
	switch {
	case len(certs) > 0 && sig != "":
		return req, "", badRequest(errors.New("both certificates and signature given"), "give certificates or a signature, not both")
	case len(certs) > 0:
		fps := make([]string, 0, len(certs))
		for i, c := range certs {
			parsed, err := redfish.ParseCertificate(c)
			if err != nil {
				return req, "", badRequest(fmt.Errorf("certificate %d: %w", i+1, err), "each certificate must be one PEM encoded X.509 certificate")
			}
			fp := redfish.Fingerprint(parsed.Raw)
			req.Certificates = append(req.Certificates, redfish.SecureBootCert{PEM: c, Fingerprint: fp})
			fps = append(fps, fp)
		}
		return req, fmt.Sprintf("bmc secureboot add %s certificate %s", db, strings.Join(fps, ",")), nil
	case sig != "":
		typ := body.SignatureType
		if typ == "" {
			typ = "EFI_CERT_SHA256_GUID"
		}
		size, ok := signatureLengths[typ]
		if !ok {
			return req, "", badRequest(fmt.Errorf("unknown signature type %q", typ), "signature_type must be EFI_CERT_SHA256_GUID, EFI_CERT_SHA384_GUID, EFI_CERT_SHA512_GUID or EFI_CERT_X509_SHA256_GUID")
		}
		hash := redfish.NormalizeFingerprint(sig)
		data, err := hex.DecodeString(hash)
		if err != nil {
			return req, "", badRequest(err, "signature must be hex encoded")
		}
		if len(data) != size {
			return req, "", badRequest(fmt.Errorf("signature is %d bytes", len(data)), fmt.Sprintf("a %s signature is %d bytes, %d hex digits", typ, size, size*2))
		}
		req.Signature, req.SignatureType = hash, typ
		return req, fmt.Sprintf("bmc secureboot add %s signature %s", db, hash), nil
	}
	return req, "", badRequest(errors.New("no key given"), "give certificates or a signature")
}

func secureBootDelete(body BmcSecureBootDeleteBody) (redfish.DeleteSecureBootKeyRequest, string, error) {
	var req redfish.DeleteSecureBootKeyRequest

	db, err := writableDatabase(body.Database)
	if err != nil {
		return req, "", err
	}
	req.Database = db

	fp := redfish.NormalizeFingerprint(body.Fingerprint)
	id := strings.TrimSpace(body.ID)
	switch {
	case fp != "" && id != "":
		return req, "", badRequest(errors.New("both fingerprint and id given"), "give a fingerprint or an id, not both")
	case fp != "":
		data, err := hex.DecodeString(fp)
		if err != nil {
			return req, "", badRequest(err, "fingerprint must be hex encoded")
		}
		// a shortened fingerprint would match nothing, and the run would report every node as having nothing to remove
		if len(data) < 32 {
			return req, "", badRequest(fmt.Errorf("fingerprint is %d hex digits", len(fp)), "fingerprint must be the whole SHA-256 or hash, at least 64 hex digits")
		}
		req.Fingerprint = fp
		return req, fmt.Sprintf("bmc secureboot delete %s %s", db, fp), nil
	case id != "":
		req.ID = id
		return req, fmt.Sprintf("bmc secureboot delete %s id %s", db, id), nil
	}
	return req, "", badRequest(errors.New("no key named"), "give the fingerprint or id of the key to remove")
}

var secureBootResetTypes = map[string]schemas.ResetKeysType{
	"to-default": schemas.ResetAllKeysToDefaultResetKeysType,
	"delete-all": schemas.DeleteAllKeysResetKeysType,
	"delete-pk":  schemas.DeletePKResetKeysType,
}

func secureBootReset(body BmcSecureBootResetBody) (redfish.ResetSecureBootKeysRequest, string, error) {
	var req redfish.ResetSecureBootKeysRequest

	typ, ok := secureBootResetTypes[body.Type]
	if !ok {
		return req, "", badRequest(fmt.Errorf("unknown reset type %q", body.Type), "type must be to-default, delete-all or delete-pk")
	}
	req.Type = typ

	operation := "bmc secureboot reset " + body.Type
	if body.Database != "" {
		db, err := writableDatabase(body.Database)
		if err != nil {
			return req, "", err
		}
		if typ == schemas.DeletePKResetKeysType {
			return req, "", badRequest(errors.New("delete-pk given a database"), "delete-pk applies to every database, leave database empty")
		}
		req.Database = db
		operation += " " + db
	}
	return req, operation, nil
}
