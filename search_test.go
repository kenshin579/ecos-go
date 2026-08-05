package ecos

import (
	"context"
	"testing"
)

func TestStatisticSearch(t *testing.T) {
	c := newTestClient(t,
		"/api/StatisticSearch/TESTKEY/json/kr/1/5/722Y001/D/20260101/20260131/0101000",
		"search_d.json")
	got, err := c.StatisticSearch(context.Background(), SearchParams{
		Page: Page{1, 5}, StatCode: "722Y001", Cycle: CycleDaily,
		Start: "20260101", End: "20260131", ItemCode1: "0101000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 31 {
		t.Errorf("TotalCount = %d, want 31", got.TotalCount)
	}
	r := got.Rows[0]
	if r.StatCode != "722Y001" || r.ItemCode1 != "0101000" || r.ItemName1 != "한국은행 기준금리" {
		t.Errorf("row[0] = %+v", r)
	}
	if r.Time != "20260101" || r.DataValue != "2.5" || r.UnitName != "연%" {
		t.Errorf("row[0] time/value = %+v", r)
	}
	if r.ItemCode2 != "" { // null → ""
		t.Errorf("ItemCode2 = %q, want empty", r.ItemCode2)
	}
	f, err := r.Float64()
	if err != nil || f != 2.5 {
		t.Errorf("Float64() = %v, %v", f, err)
	}
}

func TestStatisticSearchValidation(t *testing.T) {
	c, _ := NewClient("KEY")
	ctx := context.Background()
	cases := []SearchParams{
		{Cycle: CycleDaily, Start: "20260101", End: "20260131"},     // StatCode 없음
		{StatCode: "722Y001", Start: "20260101", End: "20260131"},   // Cycle 없음
		{StatCode: "722Y001", Cycle: CycleDaily, End: "20260131"},   // Start 없음
		{StatCode: "722Y001", Cycle: CycleDaily, Start: "20260101"}, // End 없음
	}
	for i, p := range cases {
		if _, err := c.StatisticSearch(ctx, p); err == nil {
			t.Errorf("case %d: want validation error", i)
		}
	}
}
