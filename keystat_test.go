package ecos

import (
	"context"
	"testing"
)

func TestKeyStatisticList(t *testing.T) {
	c := newTestClient(t, "/api/KeyStatisticList/TESTKEY/json/kr/1/10/", "keystat.json")
	got, err := c.KeyStatisticList(context.Background(), KeyStatisticListParams{Page: Page{1, 10}})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 101 || got.RowCount != 10 {
		t.Errorf("TotalCount=%d RowCount=%d, want 101/10", got.TotalCount, got.RowCount)
	}
	r := got.Rows[0]
	if r.ClassName != "통화량" || r.KeyStatName != "M1(협의통화, 평잔)" {
		t.Errorf("row[0] = %+v", r)
	}
	if r.Cycle != "202605" || r.UnitName != "십억원" {
		t.Errorf("row[0] cycle/unit = %+v", r)
	}
	f, err := r.Float64()
	if err != nil || f != 1397923 {
		t.Errorf("Float64() = %v, %v", f, err)
	}
}
