//go:build integration

package ecos

import (
	"context"
	"errors"
	"testing"
	"time"
)

// 실 API 스모크. 실행: ECOS_API_KEY=... go test -tags integration ./...
func newIntegrationClient(t *testing.T) *Client {
	t.Helper()
	c, err := NewClientFromEnv()
	if err != nil {
		t.Skip("ECOS_API_KEY not set")
	}
	return c
}

func ctxWithTimeout(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestIntegrationStatisticTableList(t *testing.T) {
	c := newIntegrationClient(t)
	got, err := c.StatisticTableList(ctxWithTimeout(t), TableListParams{Page: Page{1, 5}})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount < 800 || len(got.Rows) != 5 {
		t.Errorf("total=%d rows=%d", got.TotalCount, len(got.Rows))
	}
}

func TestIntegrationStatisticWord(t *testing.T) {
	c := newIntegrationClient(t)
	got, err := c.StatisticWord(ctxWithTimeout(t), WordParams{Word: "소비자동향지수"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) == 0 || got.Rows[0].Word == "" {
		t.Errorf("rows = %+v", got.Rows)
	}
}

func TestIntegrationStatisticItemList(t *testing.T) {
	c := newIntegrationClient(t)
	got, err := c.StatisticItemList(ctxWithTimeout(t), ItemListParams{StatCode: "102Y004"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) == 0 || got.Rows[0].ItemCode == "" {
		t.Errorf("rows = %+v", got.Rows)
	}
}

func TestIntegrationStatisticSearch(t *testing.T) {
	c := newIntegrationClient(t)
	got, err := c.StatisticSearch(ctxWithTimeout(t), SearchParams{
		StatCode: "722Y001", Cycle: CycleDaily,
		Start: "20260101", End: "20260110", ItemCode1: "0101000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) == 0 {
		t.Fatal("no rows")
	}
	if _, err := got.Rows[0].Float64(); err != nil {
		t.Errorf("Float64: %v", err)
	}
}

func TestIntegrationStatisticSearchNoData(t *testing.T) {
	c := newIntegrationClient(t)
	_, err := c.StatisticSearch(ctxWithTimeout(t), SearchParams{
		StatCode: "722Y001", Cycle: CycleDaily,
		Start: "20991230", End: "20991231", ItemCode1: "0101000",
	})
	if !errors.Is(err, ErrNoData) {
		t.Errorf("err = %v, want ErrNoData", err)
	}
}

func TestIntegrationKeyStatisticList(t *testing.T) {
	c := newIntegrationClient(t)
	got, err := c.KeyStatisticList(ctxWithTimeout(t), KeyStatisticListParams{Page: Page{1, 10}})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount < 100 || got.RowCount != 10 {
		t.Errorf("total=%d rowCount=%d", got.TotalCount, got.RowCount)
	}
}

func TestIntegrationStatisticMeta(t *testing.T) {
	c := newIntegrationClient(t)
	got, err := c.StatisticMeta(ctxWithTimeout(t), MetaParams{Name: "경제심리지수"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) == 0 {
		t.Fatal("no rows")
	}
}
