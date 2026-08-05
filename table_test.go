package ecos

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestClient 는 요청 경로를 검증하고 testdata fixture 를 서빙하는 Client 를 만든다.
// wantPath 가 비어 있지 않으면 요청 경로와 정확히 일치해야 한다.
func newTestClient(t *testing.T, wantPath, fixture string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wantPath != "" && r.URL.EscapedPath() != wantPath {
			t.Errorf("request path = %q, want %q", r.URL.EscapedPath(), wantPath)
		}
		b, err := os.ReadFile(filepath.Join("testdata", fixture))
		if err != nil {
			t.Fatal(err)
		}
		w.Write(b)
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient("TESTKEY", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestStatisticTableList(t *testing.T) {
	c := newTestClient(t, "/api/StatisticTableList/TESTKEY/json/kr/1/10/", "table_list.json")
	got, err := c.StatisticTableList(context.Background(), TableListParams{Page: Page{1, 10}})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 834 {
		t.Errorf("TotalCount = %d, want 834", got.TotalCount)
	}
	if len(got.Rows) == 0 {
		t.Fatal("no rows")
	}
	r := got.Rows[0]
	if r.PStatCode != "*" || r.StatCode != "0000000001" || !strings.Contains(r.StatName, "통화") {
		t.Errorf("row[0] = %+v", r)
	}
	if r.SrchYN != "N" || r.Cycle != "" { // CYCLE null → zero value
		t.Errorf("row[0] cycle/srch = %+v", r)
	}
}

func TestStatisticTableListWithStatCode(t *testing.T) {
	c := newTestClient(t, "/api/StatisticTableList/TESTKEY/json/kr/1/5/102Y004", "table_list_code.json")
	got, err := c.StatisticTableList(context.Background(), TableListParams{Page: Page{1, 5}, StatCode: "102Y004"})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 1 || got.Rows[0].StatCode != "102Y004" || got.Rows[0].Cycle != "M" {
		t.Errorf("got = %+v", got)
	}
}
