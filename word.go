package ecos

import (
	"context"
	"errors"
)

// WordParams 는 StatisticWord(통계용어사전) 요청 인자.
type WordParams struct {
	Page Page   // 제로값은 1~100
	Word string // 필수 — 검색할 용어
}

// WordRow 는 통계용어사전 응답 행. docs/api/통계용어사전.md 참고.
type WordRow struct {
	Word    string `json:"WORD"`    // 용어
	Content string `json:"CONTENT"` // 용어 설명
}

// WordResult 는 통계용어사전 응답.
type WordResult struct {
	TotalCount int
	Rows       []WordRow
}

// StatisticWord 는 통계 용어 정의를 검색한다.
func (c *Client) StatisticWord(ctx context.Context, p WordParams) (*WordResult, error) {
	if p.Word == "" {
		return nil, errors.New("ecos: StatisticWord: Word is required")
	}
	start, end := p.Page.orDefault()
	var out struct {
		Body listBody[WordRow] `json:"StatisticWord"`
	}
	if err := c.http.Get(ctx, "StatisticWord", start, end, []string{p.Word}, &out); err != nil {
		return nil, err
	}
	return &WordResult{TotalCount: out.Body.TotalCount, Rows: out.Body.Rows}, nil
}
