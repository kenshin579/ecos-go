package httpclient

import (
	"strings"
	"testing"
)

func newBare(lang string) *Client {
	return New(Config{APIKey: "TESTKEY", Lang: lang})
}

func TestBuildPath(t *testing.T) {
	c := newBare("kr")
	tests := []struct {
		name    string
		service string
		start   int
		end     int
		args    []string
		want    string
		wantErr string
	}{
		{name: "인자 없음 — 꼬리 슬래시", service: "KeyStatisticList", start: 1, end: 10,
			args: nil, want: "/api/KeyStatisticList/TESTKEY/json/kr/1/10/"},
		{name: "인자 1개", service: "StatisticItemList", start: 1, end: 10,
			args: []string{"102Y004"}, want: "/api/StatisticItemList/TESTKEY/json/kr/1/10/102Y004"},
		{name: "한글 인자 escape", service: "StatisticWord", start: 1, end: 5,
			args: []string{"소비자동향지수"},
			want: "/api/StatisticWord/TESTKEY/json/kr/1/5/" + "%EC%86%8C%EB%B9%84%EC%9E%90%EB%8F%99%ED%96%A5%EC%A7%80%EC%88%98"},
		{name: "뒤쪽 빈 인자 잘림", service: "StatisticSearch", start: 1, end: 5,
			args: []string{"722Y001", "D", "20260101", "20260131", "0101000", "", "", ""},
			want: "/api/StatisticSearch/TESTKEY/json/kr/1/5/722Y001/D/20260101/20260131/0101000"},
		{name: "전부 빈 선택 인자 잘림", service: "StatisticSearch", start: 1, end: 5,
			args: []string{"722Y001", "D", "20260101", "20260131", "", "", "", ""},
			want: "/api/StatisticSearch/TESTKEY/json/kr/1/5/722Y001/D/20260101/20260131"},
		{name: "중간 빈 인자 에러", service: "StatisticSearch", start: 1, end: 5,
			args:    []string{"722Y001", "D", "20260101", "20260131", "", "X", "", ""},
			wantErr: "empty argument"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.buildPath(tt.service, tt.start, tt.end, tt.args)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want contains %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestBuildPathLangEn(t *testing.T) {
	c := newBare("en")
	got, err := c.buildPath("StatisticTableList", 1, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := "/api/StatisticTableList/TESTKEY/json/en/1/3/"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
