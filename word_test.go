package ecos

import (
	"context"
	"strings"
	"testing"
)

func TestStatisticWord(t *testing.T) {
	c := newTestClient(t,
		"/api/StatisticWord/TESTKEY/json/kr/1/5/%EC%86%8C%EB%B9%84%EC%9E%90%EB%8F%99%ED%96%A5%EC%A7%80%EC%88%98",
		"word.json")
	got, err := c.StatisticWord(context.Background(), WordParams{Page: Page{1, 5}, Word: "소비자동향지수"})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 1 || got.Rows[0].Word != "소비자동향지수" || !strings.Contains(got.Rows[0].Content, "미시간대학") {
		t.Errorf("got = %+v", got)
	}
}

func TestStatisticWordRequiresWord(t *testing.T) {
	c, _ := NewClient("KEY")
	if _, err := c.StatisticWord(context.Background(), WordParams{}); err == nil {
		t.Fatal("want error for empty Word")
	}
}
