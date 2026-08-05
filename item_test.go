package ecos

import (
	"context"
	"testing"
)

func TestStatisticItemList(t *testing.T) {
	c := newTestClient(t, "/api/StatisticItemList/TESTKEY/json/kr/1/10/102Y004", "item_list.json")
	got, err := c.StatisticItemList(context.Background(), ItemListParams{Page: Page{1, 10}, StatCode: "102Y004"})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 6 {
		t.Errorf("TotalCount = %d, want 6", got.TotalCount)
	}
	r := got.Rows[0]
	if r.StatCode != "102Y004" || r.GrpCode != "Group1" || r.ItemCode != "ABA1" {
		t.Errorf("row[0] = %+v", r)
	}
	if r.Cycle != "A" || r.StartTime != "2003" || r.EndTime != "2025" || r.DataCnt != 23 {
		t.Errorf("row[0] cycle/기간 = %+v", r)
	}
	if r.UnitName != "십억원" || r.PItemCode != "" || r.Weight != "" {
		t.Errorf("row[0] 단위/null 필드 = %+v", r)
	}
}

func TestStatisticItemListRequiresStatCode(t *testing.T) {
	c, _ := NewClient("KEY")
	if _, err := c.StatisticItemList(context.Background(), ItemListParams{}); err == nil {
		t.Fatal("want error for empty StatCode")
	}
}
