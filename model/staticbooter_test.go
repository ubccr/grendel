package model

import (
	"strings"
	"testing"
)

func TestStaticBooterLoadJSON(t *testing.T) {
	test := strings.NewReader(TestHostListJSON)

	staticBooter, err := NewStaticBooter("", []string{}, "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	err = staticBooter.LoadJSON(test)
	if err != nil {
		t.Fatal(err)
	}

	hostList, err := staticBooter.HostList()
	if err != nil {
		t.Fatal(err)
	}

	if len(hostList) != 1 {
		t.Errorf("Wrong size for host list from json")
	}
}

const TestHostListJSON = `[
    {
        "firmware": "",
        "id": "1VCnR6qevU5BbihTIvZEhX002CI",
        "interfaces": [
            {
                "bmc": false,
                "fqdn": "tux01.compute.local",
                "ifname": "",
                "ip": "10.10.1.2",
                "mac": "d0:93:ae:e1:b5:2e"
            }
        ],
        "name": "tux01",
        "provision": true
    }
]`
