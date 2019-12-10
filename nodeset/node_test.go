package nodeset

import (
	"strings"
	"testing"
)

func TestNodeSet(t *testing.T) {
	ns, err := NewNodeSet("cpn-d13", "1-10,22,24")
	if err != nil {
		t.Fatal(err)
	}

	if ns.Len() != 12 {
		t.Errorf("Incorrect nodeset size: got %d should be %d", ns.Len(), 12)
	}

	count := 0
	for ns.Next() {
		val := ns.Value()
		if !strings.HasPrefix(val, "cpn-d13-") {
			t.Errorf("Invalid prefix: %s", val)
		}
		count++
	}

	if count != 12 {
		t.Errorf("Incorrect nodeset size from iterator: got %d should be %d", count, 12)
	}
}
