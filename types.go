package ecos

import (
	"strconv"
	"strings"
)

// Cycle 은 통계 주기 코드. 날짜 형식은 docs/api/README.md 참고.
type Cycle string

const (
	CycleAnnual      Cycle = "A"  // 년 (YYYY)
	CycleSemiAnnual  Cycle = "S"  // 반년 (YYYYSn) — 희소/레거시
	CycleQuarterly   Cycle = "Q"  // 분기 (YYYYQn)
	CycleMonthly     Cycle = "M"  // 월 (YYYYMM)
	CycleSemiMonthly Cycle = "SM" // 반월 (YYYYMMSn) — 희소/레거시
	CycleDaily       Cycle = "D"  // 일 (YYYYMMDD)
)

// Page 는 요청시작건수/요청종료건수 (건수 기반, 1부터). 제로값은 1~100.
type Page struct {
	Start int
	End   int
}

func (p Page) orDefault() (start, end int) {
	if p.Start == 0 && p.End == 0 {
		return 1, 100
	}
	return p.Start, p.End
}

// listBody 는 ECOS 성공 응답 공통 형태 {"list_total_count":N,"row":[...]}.
type listBody[T any] struct {
	TotalCount int `json:"list_total_count"`
	Rows       []T `json:"row"`
}

// parseFloat 은 ECOS 문자열 수치(DATA_VALUE 등)를 float64 로 변환한다 (쉼표 허용).
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
}
