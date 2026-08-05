package ecos

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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

func TestStatisticSearchAll(t *testing.T) {
	// 총 5행을 서빙하는 mock — 5행 < searchAllChunk 라 1회 호출로 끝나야 한다.
	rowJSON := func(day string) string {
		return `{"STAT_CODE":"722Y001","STAT_NAME":"기준금리","ITEM_CODE1":"0101000",` +
			`"ITEM_NAME1":"한국은행 기준금리","UNIT_NAME":"연%","TIME":"` + day + `","DATA_VALUE":"2.5"}`
	}
	all := []string{"20260101", "20260102", "20260103", "20260104", "20260105"}
	var gotPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.URL.Path)
		// 경로에서 start/end 파싱: /api/StatisticSearch/KEY/json/kr/{start}/{end}/...
		parts := strings.Split(r.URL.Path, "/")
		start, _ := strconv.Atoi(parts[6])
		end, _ := strconv.Atoi(parts[7])
		if start > len(all) {
			w.Write([]byte(`{"RESULT":{"CODE":"INFO-200","MESSAGE":"해당하는 데이터가 없습니다."}}`))
			return
		}
		if end > len(all) {
			end = len(all)
		}
		var rows []string
		for _, d := range all[start-1 : end] {
			rows = append(rows, rowJSON(d))
		}
		w.Write([]byte(`{"StatisticSearch":{"list_total_count":` + strconv.Itoa(len(all)) +
			`,"row":[` + strings.Join(rows, ",") + `]}}`))
	}))
	t.Cleanup(srv.Close)

	c, _ := NewClient("TESTKEY", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	got, err := c.StatisticSearchAll(context.Background(), SearchParams{
		StatCode: "722Y001", Cycle: CycleDaily, Start: "20260101", End: "20260105", ItemCode1: "0101000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 {
		t.Fatalf("rows = %d, want 5", len(got))
	}
	if got[0].Time != "20260101" || got[4].Time != "20260105" {
		t.Errorf("rows 순서 = %s..%s", got[0].Time, got[4].Time)
	}
	if len(gotPaths) != 1 { // 5행 < searchAllChunk → 1회 호출
		t.Errorf("호출 수 = %d, want 1", len(gotPaths))
	}
}

func TestStatisticSearchAllPaginates(t *testing.T) {
	// list_total_count 를 부풀려 2회째 호출을 유도하고, 2회째는 빈 row 를 반환 — 종료 검증.
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		if call == 1 {
			w.Write([]byte(`{"StatisticSearch":{"list_total_count":2000,"row":[` +
				`{"TIME":"20260101","DATA_VALUE":"1"}]}}`))
			return
		}
		w.Write([]byte(`{"StatisticSearch":{"list_total_count":2000,"row":[]}}`))
	}))
	t.Cleanup(srv.Close)
	c, _ := NewClient("TESTKEY", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	got, err := c.StatisticSearchAll(context.Background(), SearchParams{
		StatCode: "722Y001", Cycle: CycleDaily, Start: "20260101", End: "20261231",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || call != 2 {
		t.Errorf("rows=%d calls=%d, want rows=1 calls=2", len(got), call)
	}
}
