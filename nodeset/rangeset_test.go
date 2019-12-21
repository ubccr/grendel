package nodeset

import (
	"testing"
)

func TestRangeSet(t *testing.T) {
	rs, err := NewRangeSet("1-10")
	if err != nil {
		t.Fatal(err)
	}

	if rs.Len() != 10 {
		t.Errorf("Incorrect rangeset size: got %d should be %d", rs.Len(), 10)
	}
}
