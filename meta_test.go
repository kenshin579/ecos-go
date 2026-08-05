package ecos

import (
	"context"
	"testing"
)

func TestStatisticMeta(t *testing.T) {
	c := newTestClient(t,
		"/api/StatisticMeta/TESTKEY/json/kr/1/10/%EA%B2%BD%EC%A0%9C%EC%8B%AC%EB%A6%AC%EC%A7%80%EC%88%98",
		"meta.json")
	got, err := c.StatisticMeta(context.Background(), MetaParams{Page: Page{1, 10}, Name: "경제심리지수"})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 53 {
		t.Errorf("TotalCount = %d, want 53", got.TotalCount)
	}
	r := got.Rows[0]
	if r.Lvl != "2" || r.ContCode != "N13" || r.ContName != "참고 자료" || r.MetaData != "" {
		t.Errorf("row[0] = %+v", r)
	}
}

func TestStatisticMetaRequiresName(t *testing.T) {
	c, _ := NewClient("KEY")
	if _, err := c.StatisticMeta(context.Background(), MetaParams{}); err == nil {
		t.Fatal("want error for empty Name")
	}
}
