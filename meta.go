package ecos

import (
	"context"
	"errors"
)

// MetaParams 는 StatisticMeta(통계메타DB) 요청 인자.
type MetaParams struct {
	Page Page   // 제로값은 1~100
	Name string // 필수 — 데이터명 (예: "경제심리지수")
}

// MetaRow 는 통계메타DB 응답 행. docs/api/통계메타DB.md 참고.
// 트리 구조이며 말단 노드의 MetaData 에 값이 들어 있다.
type MetaRow struct {
	Lvl       string `json:"LVL"`         // 트리 깊이 ("1"~"4", 문자열)
	PContCode string `json:"P_CONT_CODE"` // 상위 콘텐츠코드
	ContCode  string `json:"CONT_CODE"`   // 콘텐츠코드
	ContName  string `json:"CONT_NAME"`   // 콘텐츠명
	MetaData  string `json:"META_DATA"`   // 메타데이터 값 (그룹 노드는 null→"")
}

// MetaResult 는 통계메타DB 응답.
type MetaResult struct {
	TotalCount int
	Rows       []MetaRow
}

// StatisticMeta 는 통계메타DB 를 데이터명으로 조회한다.
func (c *Client) StatisticMeta(ctx context.Context, p MetaParams) (*MetaResult, error) {
	if p.Name == "" {
		return nil, errors.New("ecos: StatisticMeta: Name is required")
	}
	start, end := p.Page.orDefault()
	var out struct {
		Body listBody[MetaRow] `json:"StatisticMeta"`
	}
	if err := c.http.Get(ctx, "StatisticMeta", start, end, []string{p.Name}, &out); err != nil {
		return nil, err
	}
	return &MetaResult{TotalCount: out.Body.TotalCount, Rows: out.Body.Rows}, nil
}
