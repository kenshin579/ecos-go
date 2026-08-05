package ecos

import (
	"context"
	"errors"
)

// ItemListParams 는 StatisticItemList(통계세부항목목록) 요청 인자.
type ItemListParams struct {
	Page     Page   // 제로값은 1~100
	StatCode string // 필수 — 통계표코드
}

// ItemRow 는 통계세부항목목록 응답 행. docs/api/통계세부항목목록.md 참고.
// 같은 항목이 주기(A/M 등)별로 별도 행으로 반환된다.
type ItemRow struct {
	StatCode  string `json:"STAT_CODE"`   // 통계표코드
	StatName  string `json:"STAT_NAME"`   // 통계명
	GrpCode   string `json:"GRP_CODE"`    // 항목 그룹코드 (Group1~ = 항목코드1~4 위치)
	GrpName   string `json:"GRP_NAME"`    // 항목 그룹명
	ItemCode  string `json:"ITEM_CODE"`   // 통계항목코드 (StatisticSearch 에 사용)
	ItemName  string `json:"ITEM_NAME"`   // 통계항목명
	PItemCode string `json:"P_ITEM_CODE"` // 상위 통계항목코드 (최상위는 null→"")
	PItemName string `json:"P_ITEM_NAME"` // 상위 통계항목명
	Cycle     string `json:"CYCLE"`       // 주기
	StartTime string `json:"START_TIME"`  // 수록 시작일자 (주기별 날짜 형식)
	EndTime   string `json:"END_TIME"`    // 수록 종료일자
	DataCnt   int    `json:"DATA_CNT"`    // 자료 수
	UnitName  string `json:"UNIT_NAME"`   // 단위
	Weight    string `json:"WEIGHT"`      // 가중치 (없으면 null→"")
}

// ItemListResult 는 통계세부항목목록 응답.
type ItemListResult struct {
	TotalCount int
	Rows       []ItemRow
}

// StatisticItemList 는 통계표의 세부 항목 목록을 조회한다.
func (c *Client) StatisticItemList(ctx context.Context, p ItemListParams) (*ItemListResult, error) {
	if p.StatCode == "" {
		return nil, errors.New("ecos: StatisticItemList: StatCode is required")
	}
	start, end := p.Page.orDefault()
	var out struct {
		Body listBody[ItemRow] `json:"StatisticItemList"`
	}
	if err := c.http.Get(ctx, "StatisticItemList", start, end, []string{p.StatCode}, &out); err != nil {
		return nil, err
	}
	return &ItemListResult{TotalCount: out.Body.TotalCount, Rows: out.Body.Rows}, nil
}
