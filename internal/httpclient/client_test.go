package httpclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

// --- Get / 에러 매핑 ---

func serveJSON(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestGetDecodesSuccess(t *testing.T) {
	srv := serveJSON(t, `{"StatisticWord":{"list_total_count":1,"row":[{"WORD":"용어","CONTENT":"설명"}]}}`)
	c := New(Config{APIKey: "TESTKEY", BaseURL: srv.URL})
	var out struct {
		Body struct {
			Total int `json:"list_total_count"`
			Rows  []struct {
				Word string `json:"WORD"`
			} `json:"row"`
		} `json:"StatisticWord"`
	}
	if err := c.Get(context.Background(), "StatisticWord", 1, 5, []string{"용어"}, &out); err != nil {
		t.Fatal(err)
	}
	if out.Body.Total != 1 || len(out.Body.Rows) != 1 || out.Body.Rows[0].Word != "용어" {
		t.Errorf("decoded = %+v", out)
	}
}

func TestGetMapsInfo200ToErrNoData(t *testing.T) {
	srv := serveJSON(t, `{"RESULT":{"CODE":"INFO-200","MESSAGE":"해당하는 데이터가 없습니다."}}`)
	c := New(Config{APIKey: "TESTKEY", BaseURL: srv.URL})
	var out map[string]any
	err := c.Get(context.Background(), "StatisticSearch", 1, 5, nil, &out)
	if !errors.Is(err, ErrNoData) {
		t.Fatalf("err = %v, want ErrNoData", err)
	}
}

func TestGetMapsResultToAPIError(t *testing.T) {
	srv := serveJSON(t, `{"RESULT":{"CODE":"INFO-100","MESSAGE":"인증키가 유효하지 않습니다."}}`)
	c := New(Config{APIKey: "TESTKEY", BaseURL: srv.URL})
	var out map[string]any
	err := c.Get(context.Background(), "KeyStatisticList", 1, 2, nil, &out)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Code != "INFO-100" || !strings.Contains(apiErr.Message, "인증키") {
		t.Errorf("apiErr = %+v", apiErr)
	}
}

func TestGetNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)
	c := New(Config{APIKey: "TESTKEY", BaseURL: srv.URL})
	var out map[string]any
	err := c.Get(context.Background(), "StatisticWord", 1, 5, []string{"용어"}, &out)
	if err == nil || !strings.Contains(err.Error(), "502") {
		t.Fatalf("err = %v, want http status 502", err)
	}
}

func TestErrorsNeverContainAPIKey(t *testing.T) {
	// transport 오류(연결 불가)의 *url.Error 는 전체 URL(키 포함)을 담는다 — 마스킹 검증.
	c := New(Config{APIKey: "SECRETKEY123", BaseURL: "http://127.0.0.1:1", Timeout: time.Second})
	var out map[string]any
	err := c.Get(context.Background(), "KeyStatisticList", 1, 2, nil, &out)
	if err == nil {
		t.Fatal("want error")
	}
	if strings.Contains(err.Error(), "SECRETKEY123") {
		t.Fatalf("error leaks API key: %v", err)
	}
}
