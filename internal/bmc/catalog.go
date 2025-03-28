// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
	"golang.org/x/crypto/openpgp"
)

const (
	Dell_Catalog_Download_location = "https://downloads.dell.com/catalog/Catalog.xml.gz"
	Dell_GPG_0x1285491434D8786F    = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1.4.12 (GNU/Linux)

mQINBE9RLYYBEADEAmJvn2y182B6ZUr+u9I29f2ue87p6HQreVvPbTjiXG4z2/k0
l/Ov0DLImXFckaeVSSrqjFnEGUd3DiRr9pPb1FqxOseHRZv5IgjCTKZyj9Jvu6bx
U9WL8u4+GIsFzrgS5G44g1g5eD4Li4sV46pNBTp8d7QEF4e2zg9xk2mcZKaT+STl
O0Q2WKI7qN8PAoGd1SfyW4XDsyfaMrJKmIJTgUxe9sHGj+UmTf86ZIKYh4pRzUQC
WBOxMd4sPgqVfwwykg/y2CQjrorZcnUNdWucZkeXR0+UCR6WbDtmGfvN5H3htTfm
Nl84Rwzvk4NT/By4bHy0nnX+WojeKuygCZrxfpSqJWOKhQeH+YHKm1oVqg95jvCl
vBYTtDNkpJDbt4eBAaVhuEPwjCBsfff/bxGCrzocoKlh0+hgWDrr2S9ePdrwv+rv
2cgYfUcXEHltD5Ryz3u5LpiC5zDzNYGFfV092xbpG/B9YJz5GGj8VKMslRhYpUjA
IpBDlYhOJ+0uVAAKPeeZGBuFx0A1y/9iutERinPx8B9jYjO9iETzhKSHCWEov/yp
X6k17T8IHfVj4TSwL6xTIYFGtYXIzhInBXa/aUPIpMjwt5OpMVaJpcgHxLam6xPN
FYulIjKAD07FJ3U83G2fn9W0lmr11hVsFIMvo9JpQq9aryr9CRoAvRv7OwARAQAB
tGBEZWxsIEluYy4sIFBHUkUgMjAxMiAoUEcgUmVsZWFzZSBFbmdpbmVlcmluZyBC
dWlsZCBHcm91cCAyMDEyKSA8UEdfUmVsZWFzZV9FbmdpbmVlcmluZ0BEZWxsLmNv
bT6JAjcEEwEKACEFAk9RLYYCGwMFCwkIBwMFFQoJCAsFFgIDAQACHgECF4AACgkQ
EoVJFDTYeG9eBw//asbM4KRxBfFi9RmzRNitOiFEN1FqTbE5ujjN+9m9OEb+tB3Z
Fxv0bEPb2kUdpEwtMq6CgC5n8UcLbe5TF82Ho8r2mVYNRh5RltdvAtDK2pQxCOh+
i2b9im6GoIZa1HWNkKvKiW0dmiYYBvWlu78iQ8JpIixRIHXwEdd1nQIgWxjVix11
VDr+hEXPRFRMIyRzMteiq2w/XNTUZAh275BaZTmLdMLoYPhHO99AkYgsca9DK9f0
z7SYBmxgrKAs9uoNnroo4UxodjCFZHDu+UG2efP7SvJnq9v6XaC7ZxqBG8AObEsw
qGaLv9AN3t4oLjWhrAIoNWwIM1LWpYLmKjFYlLHaf30MYhJ8J7GHzgxANnkOP4g0
RiXeYNLcNvsZGXZ61/KzuvE6YcsGXSMVKRVaxLWkgS559OSjEcQV1TD65b+bttIe
EEYmcS8jLKL+q2T1qTKnmD6VuNCtZwlsxjR5wHnxORjumtC5kbkt1lxjb0l2gNvT
3ccA6FEWKS/uvtleQDeGFEA6mrKEGoD4prQwljPV0MZwyzWqclOlM7g21i/+SUj8
ND2Iw0dCs4LvHkf4F1lNdV3QB41ZQGrbQqcCcJFm3qRsYhi4dg8+24j3bNrSHjxo
sGtcmOLv15jXA1bxyXHkn0HPG6PZ27dogsJnAD1GXEH2S8yhJclYuL0JE0CIRgQQ
EQoABgUCT1E0sQAKCRDKd5UdI7ZqnSh9AJ9jXsuabnqEfz5DQwWbmMDgaLGXiwCf
XA9nDiBc1oyCXVabfbcMs8J0ktqIXgQQEQoABgUCT1E0yQAKCRB1a6cLEBnO1iQA
AP98ZGIFya5HOUt6RAxL3TpMRSP4ihFVg8EUwZi9m9IVnwD/SXskcNW1PsZJO/bR
aNVUZIUniDIxbYuj5++8KwBksZiJAhwEEAEIAAYFAk9ROHAACgkQ2XsrqIahDMCl
CRAAhY59a8BEIQUR9oVeQG8XNZjaIAnybq7/IxeFMkYKr0ZsoxFy+BDHXl2bajql
ILnd9IYaxsLDh+8lwOTBiHhWfNg4b96gDPg5h4XaHgZ+zPmLMuEL/hQoKdYKZDmM
1b0YinoV5KisovpC5IZi1AtAFs5EL++NysGeY3RffIpynFRsUomZmBx2Gz99xkiU
XgbT9aXAJTKfsQrFLASM6LVib/oA3Sx1MQXGFU3IA65ye/UXA4A53dSbE3m10RYB
ZoeS6BUQ9yFtmRybZtibW5RNOGZCD6/Q3Py65tyWeUUeRiKyksAKl1IGpb2awA3r
AbrNd/xe3qAfR+NMlnidtU4nJO3GG6B7HTPQfGp8c69+YVaMML3JcyvACCJfVC0a
Lg+ru6UkCDSfWpuqgdMJrhm12FM16r1X3aFwDA1qwnCQcsWJWManqD8ljHl3S2Vd
0nyPcLZsGGuZfTCsK9pvhd3FANC5yncwe5oi1ueiU3KrIWfvI08NzCsj8H2ZCAPK
pz51zZfDgblMFXHTmDNZWj4QrHG01LODe+mZnsCFrBWbiP13EwsJ9WAMZ6L+/iwJ
jjoi9e4IDmTOBJdGUoWKELYMfglpF5EPGUcsYaA9FfcSCgm9QR31Ixy+F95bhCTV
T26xwTtNMYFdZ2rMRjA/TeTNfl5KHLi6YvAgtMaBT8nYKweIRgQTEQIABgUCT1FC
zwAKCRAhq+73kvD8CSnUAJ4j3Q6r+DESBbvISTD4cX3WcpMepwCfX8oc1nHL4bFb
VBS6BP9aHFcBqJ65Ag0ET1Ev4QEQANlcF8dbXMa6vXSmznnESEotJ2ORmvr5R1zE
gqQJOZ9DyML9RAc0dmt7IwgwUNX+EfY8LhXLKvHWrj2mBXm261A9SU8ijQOPHFAg
/SYyP16JqfSx2jsvWGBIjEXF4Z3SW/JD0yBNAXlWLWRGn3dx4cHyxmeGjCAc/6t3
22Tyi5XLtwKGxA/vEHeuGmTuKzNIEnWZbdnqALcrT/xK6PGjDo45VKx8mzLal/mn
cXmvaNVEyld8MMwQfkYJHvZXwpWYXaWTgAiMMm+yEd0gaBZJRPBSCETYz9bENePW
EMnrd9I65pRl4X27stDQ91yO2dIdfamVqti436ZvLc0L4EZ7HWtjN53vgXobxMzz
4/6eH71BRJujG1yYEk2J1DUJKV1WUfV8Ow0TsJVNQRM/L9v8imSMdiR12BjzHism
ReMvaeAWfUL7Q1tgwvkZEFtt3sl8o0eoB39R8xP4p1ZApJFRj6N3ryCTVQw536QF
GEb+C51MdJbXFSDTRHFlBFVsrSE6PxB24RaQ+37w3lQZp/yCoGqA57S5VVIAjAll
4Yl347WmNX9THogjhhzuLkXW+wNGIPX9SnZopVAfuc4hj0TljVa6rbYtiw6HZNmv
vr1/vSQMuAyl+HkEmqaAhDgVknb3MQqUQmzeO/WtgSqYSLb7pPwDKYy7I1BojNiO
t+qMj6P5ABEBAAGJAh4EGAEKAAkFAk9RL+ECGwwACgkQEoVJFDTYeG/6mA/4q6DT
SLwgKDiVYIRpqacUwQLySufOoAxGSEde8vGRpcGEC+kWt1aqIiE4jdlxFH7Cq5Sn
wojKpcBLIAvIYk6x9wofz5cx10s5XHq1Ja2jKJV2IPT5ZdJqWBc+M8K5LJelemYR
Zoe50aT0jbN5YFRUkuU0cZZyqv98tZzTYO9hdG4sH4gSZg4OOmUtnP1xwSqLWdDf
0RpnjDuxMwJM4m6G3UbaQ4w1K8hvUtZo9uC9+lLHq4eP9gcxnvi7Xg6mI3UXAXiL
YXXWNY09kYXQ/jjrpLxvWIPwk6zb02jsuD08j4THp5kU4nfujj/GklerGJJp1ypI
OEwV4+xckAeKGUBIHOpyQq1fn5bz8IituSF3xSxdT2qfMGsoXmvfo2l8T9QdmPyd
b4ZGYhv24GFQZoyMAATLbfPmKvXJAqomSbp0RUjeRCom7dbD1FfLRbtpRD73zHar
BhYYZNLDMls3IIQTFuRvNeJ7XfGwhkSE4rtY91J93eM77xNr4sXeYG+RQx4y5Hz9
9Q/gLas2celP6Zp8Y4OECdveX3BA0ytI8L02wkoJ8ixZnpGskMl4A0UYI4w4jZ/z
dqdpc9wPhkPj9j+eF2UInzWOavuCXNmQz1WkLP/qlR8DchJtUKlgZq9ThshK4gTE
SNnmxzdpR6pYJGbEDdFyZFe5xHRWSlrC3WTbzg==
=w0ey
-----END PGP PUBLIC KEY BLOCK-----`
)

type DellCatalog struct {
	BaseLocation                string               `xml:"baseLocation,attr"`
	BaseLocationAccessProtocols string               `xml:"baseLocationAccessProtocols,attr"`
	DateTime                    string               `xml:"dateTime,attr"`
	Version                     string               `xml:"version,attr"`
	SoftwareComponents          []SoftwareComponents `xml:"SoftwareComponent"`
}
type SoftwareComponents struct {
	Name            string `xml:"Name>Display"`
	DateTime        string `xml:"dateTime,attr"`
	DellVersion     string `xml:"dellVersion,attr"`
	Path            string `xml:"path,attr"`
	RebootRequired  string `xml:"rebootRequired,attr"`
	ReleaseDate     string `xml:"releaseDate,attr"`
	Size            string `xml:"size,attr"`
	VendorVersion   string `xml:"vendorVersion,attr"`
	HashMD5         string `xml:"hashMD5,attr"`
	ComponentType   string `xml:"ComponentType>Display"`
	Description     string `xml:"Description>Display"`
	LUCategory      string `xml:"LUCategory>Display"`
	Category        string `xml:"Category>Display"`
	RevisionHistory string `xml:"RevisionHistory>Display"`
	Criticality     string `xml:"Criticality>Display"`
	ImportantInfo   struct {
		Info string `xml:"Display"`
		URL  string `xml:"URL,attr"`
	}
	SupportedDevices struct {
		Device []struct {
			Name        string `xml:"Display"`
			ComponentID string `xml:"componentID,attr"`
			Embedded    string `xml:"embedded,attr"`
			PCIInfo     struct {
				DeviceID    string `xml:"deviceID,attr"`
				SubDeviceID string `xml:"subDeviceID,attr"`
				VendorID    string `xml:"vendorID,attr"`
				SubVendorID string `xml:"subVendorID,attr"`
			}
		}
	}
	SupportedSystems []struct {
		Brand []struct {
			Key    string `xml:"key,attr"`
			Prefix string `xml:"prefix,attr"`
			Name   string `xml:"Display"`
			Model  []struct {
				SystemID     string `xml:"systemID,attr"`
				SystemIDType string `xml:"systemIDType,attr"`
				Name         string `xml:"Display"`
			}
		}
	}
}

func DownloadFirmware(url, output, component, catalogSum string, pw progress.Writer) error {
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer out.Close()

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Wget/1.21.2")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var download bytes.Buffer
	go io.Copy(&download, resp.Body)

	tracker := progress.Tracker{
		Message: component,
		Total:   resp.ContentLength,
	}
	pw.AppendTracker(&tracker)

	for !tracker.IsDone() {
		tracker.SetValue(int64(download.Len()))
		time.Sleep(time.Millisecond * 50)
	}

	downloaded, err := io.ReadAll(&download)
	if err != nil {
		return nil
	}

	h := md5.New()
	_, err = h.Write(downloaded)
	if err != nil {
		return err
	}

	if resp.Header.Get("Content-Type") == "application/x-gzip" {
		err = VerifyPGP(Dell_Catalog_Download_location+".sha512.sign", downloaded)
		if err != nil {
			return err
		}

		uncompressed, err := gzip.NewReader(bytes.NewReader(downloaded))
		if err != nil {
			return err
		}

		downloaded, err = io.ReadAll(uncompressed)
		if err != nil {
			return err
		}
	}

	_, err = out.Write(downloaded)
	if err != nil {
		return err
	}

	downloadedSum := hex.EncodeToString(h.Sum(nil))
	if catalogSum != "" && downloadedSum != catalogSum {
		return fmt.Errorf("%s warning: md5sum mismatch. expected: %s, downloaded: %s", component, catalogSum, downloadedSum)
	}

	return nil
}

func VerifyPGP(signatureURL string, file []byte) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", signatureURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Wget/1.21.2")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	kr, err := openpgp.ReadArmoredKeyRing(strings.NewReader(Dell_GPG_0x1285491434D8786F))
	if err != nil {
		return err
	}
	_, err = openpgp.CheckArmoredDetachedSignature(kr, bytes.NewReader(file), resp.Body)

	return err
}
