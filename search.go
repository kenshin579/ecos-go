package ecos

import (
	"context"
	"errors"
)

// SearchParams 는 StatisticSearch(통계조회) 요청 인자.
// 항목코드 1~4 는 선택이며, 모두 생략하면 전체 항목이 반환된다 (응답이 매우 커질 수 있음 —
// docs/api/통계조회.md 참고).
type SearchParams struct {
	Page      Page   // 제로값은 1~100
	StatCode  string // 필수 — 통계표코드 (예: "722Y001")
	Cycle     Cycle  // 필수 — 주기
	Start     string // 필수 — 검색시작일자 (주기별 날짜 형식)
	End       string // 필수 — 검색종료일자
	ItemCode1 string // 선택 — 통계항목코드1
	ItemCode2 string // 선택
	ItemCode3 string // 선택
	ItemCode4 string // 선택
}

// SearchRow 는 통계조회 응답 행. docs/api/통계조회.md 참고.
type SearchRow struct {
	StatCode  string `json:"STAT_CODE"`  // 통계표코드
	StatName  string `json:"STAT_NAME"`  // 통계명
	ItemCode1 string `json:"ITEM_CODE1"` // 통계항목코드1
	ItemName1 string `json:"ITEM_NAME1"`
	ItemCode2 string `json:"ITEM_CODE2"` // null → ""
	ItemName2 string `json:"ITEM_NAME2"`
	ItemCode3 string `json:"ITEM_CODE3"`
	ItemName3 string `json:"ITEM_NAME3"`
	ItemCode4 string `json:"ITEM_CODE4"`
	ItemName4 string `json:"ITEM_NAME4"`
	UnitName  string `json:"UNIT_NAME"`  // 단위
	Weight    string `json:"WGT"`        // 가중치 (없으면 null→"")
	Time      string `json:"TIME"`       // 시점 (주기별 날짜 형식)
	DataValue string `json:"DATA_VALUE"` // 값 (문자열 원본)
}

// Float64 는 DataValue 를 float64 로 변환한다.
func (r SearchRow) Float64() (float64, error) { return parseFloat(r.DataValue) }

// SearchResult 는 통계조회 응답.
type SearchResult struct {
	TotalCount int
	Rows       []SearchRow
}

// StatisticSearch 는 통계 데이터를 조회한다.
func (c *Client) StatisticSearch(ctx context.Context, p SearchParams) (*SearchResult, error) {
	if p.StatCode == "" {
		return nil, errors.New("ecos: StatisticSearch: StatCode is required")
	}
	if p.Cycle == "" {
		return nil, errors.New("ecos: StatisticSearch: Cycle is required")
	}
	if p.Start == "" || p.End == "" {
		return nil, errors.New("ecos: StatisticSearch: Start and End are required")
	}
	start, end := p.Page.orDefault()
	args := []string{p.StatCode, string(p.Cycle), p.Start, p.End,
		p.ItemCode1, p.ItemCode2, p.ItemCode3, p.ItemCode4}
	var out struct {
		Body listBody[SearchRow] `json:"StatisticSearch"`
	}
	if err := c.http.Get(ctx, "StatisticSearch", start, end, args, &out); err != nil {
		return nil, err
	}
	return &SearchResult{TotalCount: out.Body.TotalCount, Rows: out.Body.Rows}, nil
}

// searchAllChunk 는 StatisticSearchAll 의 페이지 크기.
const searchAllChunk = 1000

// StatisticSearchAll 은 list_total_count 기준으로 전 페이지를 자동 수집한다.
// p.Page 는 무시된다. 페이지 경계에서 INFO-200 이 오면 수집분을 반환한다.
func (c *Client) StatisticSearchAll(ctx context.Context, p SearchParams) ([]SearchRow, error) {
	var all []SearchRow
	for start := 1; ; {
		p.Page = Page{Start: start, End: start + searchAllChunk - 1}
		res, err := c.StatisticSearch(ctx, p)
		if err != nil {
			if errors.Is(err, ErrNoData) && all != nil {
				return all, nil
			}
			return nil, err
		}
		all = append(all, res.Rows...)
		if len(res.Rows) == 0 || len(all) >= res.TotalCount {
			return all, nil
		}
		// 요청 범위보다 적게 반환되는 경우에도 행을 건너뛰지 않도록 실수신 행 수만큼 전진.
		start += len(res.Rows)
	}
}
