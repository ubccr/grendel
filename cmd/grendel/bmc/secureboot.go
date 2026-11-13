// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
	"golang.org/x/term"
)

var (
	secureBootCmd = &cobra.Command{
		Use:   "secureboot",
		Short: "UEFI secure boot commands",
		Long:  "Show and change UEFI secure boot, and manage the keys in its PK, KEK, db and dbx databases. Changes take effect at the node's next boot",
	}

	// secureBootAckColumns is the table every secure boot change prints
	secureBootAckColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Message"},
	}

	secureBootWritable = []string{"PK", "KEK", "db", "dbx"}
)

func init() {
	bmcCmd.AddCommand(secureBootCmd)
}

// completeDatabases completes the database argument at position pos, after the nodeset
func completeDatabases(pos int, names []string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == pos {
			return names, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveDefault
	}
}

// printAcks prints the result of a secure boot change and returns the error that sets the exit code
func printAcks(res *client.BmcResponseModelBmcAck, filter []string, options *shared.TableOptions) error {
	t := shared.NewTable(secureBootAckColumns, filter).Options(*options).Empty("-")
	for _, r := range res.Results.Value {
		result := r.Value
		t.AppendRow(result.Node.Value, colorState(result.Success.Value), result.Message.Value)
	}
	t.Print()

	s := res.Summary.Value
	return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
}

// errNoTerminal makes a script that forgot --yes fail, rather than exit 0 having changed nothing
var errNoTerminal = errors.New("no terminal to ask for confirmation on, pass --yes")

// confirmed asks before a change that can leave nodes unable to boot. Answering no is not an error, but having no terminal to ask on is
func confirmed(yes bool, label string) (bool, error) {
	if yes {
		return true, nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false, errNoTerminal
	}
	prompt := promptui.Prompt{Label: label, IsConfirm: true}
	if _, err := prompt.Run(); err != nil {
		fmt.Println("Cancelled, nothing was changed.")
		return false, nil
	}
	return true, nil
}

// readCertificates reads every certificate in a PEM file, or the one certificate in a DER file
func readCertificates(path string) ([]*x509.Certificate, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var certs []*x509.Certificate
	rest := raw
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		certs = append(certs, cert)
	}
	if len(certs) > 0 {
		return certs, nil
	}

	cert, err := x509.ParseCertificate(raw)
	if err != nil {
		return nil, fmt.Errorf("%s holds no PEM or DER certificate", path)
	}
	return []*x509.Certificate{cert}, nil
}

func certPEM(cert *x509.Certificate) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))
}

// certFingerprint is the SHA-256 of a certificate in hex, how the server matches a certificate across nodes
func certFingerprint(cert *x509.Certificate) string {
	sum := sha256.Sum256(cert.Raw)
	return hex.EncodeToString(sum[:])
}
