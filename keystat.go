package ecos

import "context"

// KeyStatisticListParams 는 KeyStatisticList(100대통계지표) 요청 인자.
// 서비스별 추가 인자가 없다.
type KeyStatisticListParams struct {
	Page Page // 제로값은 1~100
}

// KeyStatRow 는 100대통계지표 응답 행. docs/api/100대통계지표.md 참고.
type KeyStatRow struct {
	ClassName   string `json:"CLASS_NAME"`   // 통계 분류명
	KeyStatName string `json:"KEYSTAT_NAME"` // 지표명
	DataValue   string `json:"DATA_VALUE"`   // 값 (문자열 원본)
	Cycle       string `json:"CYCLE"`        // 수록 시점 — 주기 코드가 아니라 날짜
	UnitName    string `json:"UNIT_NAME"`    // 단위
}

// Float64 는 DataValue 를 float64 로 변환한다.
func (r KeyStatRow) Float64() (float64, error) { return parseFloat(r.DataValue) }

// KeyStatisticListResult 는 100대통계지표 응답.
// RowCount 는 이 서비스에만 있는 상위 필드 (이번 응답의 행 수).
type KeyStatisticListResult struct {
	TotalCount int
	RowCount   int
	Rows       []KeyStatRow
}

// KeyStatisticList 는 100대 통계지표의 최신 값을 조회한다.
func (c *Client) KeyStatisticList(ctx context.Context, p KeyStatisticListParams) (*KeyStatisticListResult, error) {
	start, end := p.Page.orDefault()
	var out struct {
		Body struct {
			TotalCount int          `json:"list_total_count"`
			RowCount   int          `json:"row_count"`
			Rows       []KeyStatRow `json:"row"`
		} `json:"KeyStatisticList"`
	}
	if err := c.http.Get(ctx, "KeyStatisticList", start, end, nil, &out); err != nil {
		return nil, err
	}
	return &KeyStatisticListResult{
		TotalCount: out.Body.TotalCount,
		RowCount:   out.Body.RowCount,
		Rows:       out.Body.Rows,
	}, nil
}
