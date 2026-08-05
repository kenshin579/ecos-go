package ecos

import "context"

// TableListParams 는 StatisticTableList(통계표목록) 요청 인자.
type TableListParams struct {
	Page     Page   // 제로값은 1~100
	StatCode string // 선택 — 지정 시 해당 통계표만
}

// TableRow 는 통계표목록 응답 행. docs/api/통계표목록.md 참고.
type TableRow struct {
	PStatCode string `json:"P_STAT_CODE"` // 상위 통계표코드 (최상위는 "*")
	StatCode  string `json:"STAT_CODE"`   // 통계표코드
	StatName  string `json:"STAT_NAME"`   // 통계명
	Cycle     string `json:"CYCLE"`       // 주기 (그룹 노드는 null→"")
	SrchYN    string `json:"SRCH_YN"`     // "Y"면 StatisticSearch 조회 가능
	OrgName   string `json:"ORG_NAME"`    // 출처 (대부분 null→"")
}

// TableListResult 는 통계표목록 응답.
type TableListResult struct {
	TotalCount int
	Rows       []TableRow
}

// StatisticTableList 는 통계표 목록을 조회한다.
func (c *Client) StatisticTableList(ctx context.Context, p TableListParams) (*TableListResult, error) {
	start, end := p.Page.orDefault()
	var out struct {
		Body listBody[TableRow] `json:"StatisticTableList"`
	}
	if err := c.http.Get(ctx, "StatisticTableList", start, end, []string{p.StatCode}, &out); err != nil {
		return nil, err
	}
	return &TableListResult{TotalCount: out.Body.TotalCount, Rows: out.Body.Rows}, nil
}
