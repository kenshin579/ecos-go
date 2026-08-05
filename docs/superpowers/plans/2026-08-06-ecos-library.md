# ecos-go 라이브러리 구현 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `docs/api/` 명세를 근거로 한국은행 ECOS OpenAPI 6개 서비스의 Go 클라이언트 라이브러리(`github.com/kenshin579/ecos-go`)를 구현하고 v0.1.0 릴리스를 준비한다.

**Architecture:** 단일 flat 패키지 `ecos`(서비스 6개 = 메서드 6개 + `-All` 헬퍼 1개) + `internal/httpclient`(path segment URL 조립, JSON 디코드, RESULT 에러 매핑, 인증키 마스킹). 스펙: `docs/superpowers/specs/2026-08-06-ecos-api-docs-design.md`의 "(참고) 2단계" 섹션. API 명세: `docs/api/*.md`.

**Tech Stack:** Go 1.25, 표준 라이브러리만(외부 의존성 0 — **테스트도 testify 없이 stdlib로 작성**), httptest + testdata fixture 유닛 테스트, `-tags integration` 실 API 스모크.

---

## 공통 준비

- 저장소: `/Users/user/src/workspace_moneyflow/ecos-go` (main = 839df08, 문서 머지됨)
- fixture 원본: `/private/tmp/claude-501/-Users-user-src-workspace-moneyflow/14d0c594-3468-4ea5-8a0d-24728a1cd56c/scratchpad/ecos-fixtures/` — Task 1에서 `testdata/`로 복사. **scratchpad가 비워졌으면** 문서 플랜(`docs/superpowers/plans/2026-08-06-ecos-api-docs.md`) Task 1의 curl 명령으로 재수집한다.
- 커밋 author: `kenshin579@hotmail.com` (로컬 config 설정됨), 커밋 트레일러 `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`
- 검증 공통 커맨드: `go build ./... && go vet ./... && go test ./...`

## 파일 구조 (완성 시)

```
ecos-go/
├── go.mod                    # module github.com/kenshin579/ecos-go, go 1.25
├── LICENSE                   # MIT (opendart-go 것 복사)
├── client.go                 # Client, NewClient
├── config.go                 # Option 4종
├── from_env.go               # NewClientFromEnv (ECOS_API_KEY)
├── errors.go                 # APIError alias, ErrNoData
├── types.go                  # Cycle 상수, Page, listBody[T], parseFloat
├── table.go / word.go / item.go / search.go / keystat.go / meta.go
├── *_test.go                 # 서비스별 유닛 테스트 (httptest + testdata)
├── integration_test.go       # -tags integration
├── internal/httpclient/      # client.go + client_test.go
├── testdata/                 # Task 1에서 복사한 실응답 fixture
├── examples/basic/main.go
└── scripts/release.sh
```

---

### Task 1: 스캐폴딩 (브랜치·go.mod·LICENSE·testdata)

**Files:**
- Create: `go.mod`, `LICENSE`, `testdata/*.json`

- [ ] **Step 1: 브랜치 생성**

```bash
cd /Users/user/src/workspace_moneyflow/ecos-go
git checkout main && git pull origin main
git checkout -b feature/ecos-library
```

- [ ] **Step 2: go.mod + LICENSE**

```bash
go mod init github.com/kenshin579/ecos-go
cp ../opendart-go/LICENSE LICENSE
```

`go.mod` 내용 확인 — `go mod init` 이 `go 1.25.x` 처럼 패치 버전을 쓰면 `go 1.25` 로 수정한다 (opendart-go/fmp-go 와 동일 관례):

```
module github.com/kenshin579/ecos-go

go 1.25
```

- [ ] **Step 3: testdata 복사**

```bash
FIX=/private/tmp/claude-501/-Users-user-src-workspace-moneyflow/14d0c594-3468-4ea5-8a0d-24728a1cd56c/scratchpad/ecos-fixtures
mkdir -p testdata
cp "$FIX"/table_list.json "$FIX"/table_list_code.json "$FIX"/word.json \
   "$FIX"/item_list.json "$FIX"/search_d.json "$FIX"/keystat.json "$FIX"/meta.json \
   "$FIX"/err_info100.json "$FIX"/err_info200.json "$FIX"/err_error100.json testdata/
ls testdata/   # 10개
```

- [ ] **Step 4: 커밋**

```bash
git add go.mod LICENSE testdata/
git commit -m "chore: ecos-go 모듈 스캐폴딩 (go.mod, LICENSE, testdata fixture)"
```

---

### Task 2: internal/httpclient — buildPath (TDD)

**Files:**
- Create: `internal/httpclient/client.go` (일부), `internal/httpclient/client_test.go` (일부)

- [ ] **Step 1: 실패하는 테스트 작성** — `internal/httpclient/client_test.go`

```go
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
			args: []string{"722Y001", "D", "20260101", "20260131", "", "X", "", ""},
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
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test ./internal/httpclient/`
Expected: FAIL (컴파일 에러 — `New`, `Config`, `buildPath` 미정의)

- [ ] **Step 3: 최소 구현** — `internal/httpclient/client.go`

```go
// Package httpclient 는 ECOS OpenAPI REST 호출의 단일 GET 통로다.
// path segment URL 조립, RESULT 에러 매핑, 인증키 마스킹을 담당한다.
package httpclient

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL 은 ECOS OpenAPI 베이스 URL.
const DefaultBaseURL = "https://ecos.bok.or.kr"

// Config 는 Client 생성 인자.
type Config struct {
	APIKey     string
	BaseURL    string        // 빈 값이면 DefaultBaseURL
	Lang       string        // 빈 값이면 "kr"
	Timeout    time.Duration // 0이면 30s
	HTTPClient *http.Client  // nil이면 기본 클라이언트
}

// Client 는 ECOS HTTP 계층.
type Client struct {
	apiKey  string
	baseURL string
	lang    string
	http    *http.Client
}

// New 는 Config 로 Client 를 만든다.
func New(cfg Config) *Client {
	base := cfg.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	lang := cfg.Lang
	if lang == "" {
		lang = "kr"
	}
	hc := cfg.HTTPClient
	if hc == nil {
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		hc = &http.Client{Timeout: timeout}
	}
	return &Client{apiKey: cfg.APIKey, baseURL: base, lang: lang, http: hc}
}

// buildPath 는 ECOS path segment URL 경로를 조립한다.
// /api/{서비스}/{키}/json/{lang}/{start}/{end}/{인자...}
// 뒤쪽의 빈 선택 인자는 잘라내고, 중간에 빈 인자가 있으면 에러다.
// 서비스별 인자가 없으면 꼬리 슬래시를 붙인다 (ECOS 관례).
func (c *Client) buildPath(service string, start, end int, args []string) (string, error) {
	n := len(args)
	for n > 0 && args[n-1] == "" {
		n--
	}
	segs := []string{"api", service, c.apiKey, "json", c.lang,
		strconv.Itoa(start), strconv.Itoa(end)}
	for _, a := range args[:n] {
		if a == "" {
			return "", fmt.Errorf("ecos: %s: empty argument in the middle of path", service)
		}
		segs = append(segs, url.PathEscape(a))
	}
	p := "/" + strings.Join(segs, "/")
	if n == 0 {
		p += "/"
	}
	return p, nil
}
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test ./internal/httpclient/`
Expected: PASS

- [ ] **Step 5: 커밋**

```bash
git add internal/httpclient/
git commit -m "feat: httpclient path segment URL 조립 (escape·꼬리 인자 생략)"
```

---

### Task 3: internal/httpclient — Get + 에러 매핑 + 키 마스킹 (TDD)

**Files:**
- Modify: `internal/httpclient/client.go`, `internal/httpclient/client_test.go`

- [ ] **Step 1: 실패하는 테스트 추가** — `client_test.go` 에 append

```go
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
```

import 블록에 `"context"`, `"errors"`, `"net/http/httptest"` 추가.

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test ./internal/httpclient/`
Expected: FAIL (컴파일 에러 — `Get`, `APIError`, `ErrNoData` 미정의)

- [ ] **Step 3: 구현** — `client.go` 에 append

```go
// APIError 는 ECOS RESULT 응답 (INFO-200 제외).
type APIError struct {
	Code    string // 예: ERROR-100, INFO-100
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("ecos: [%s] %s", e.Code, e.Message)
}

// ErrNoData 는 INFO-200 (해당하는 데이터가 없습니다).
var ErrNoData = errors.New("ecos: no data (INFO-200)")

// resultEnvelope 는 ECOS 실패 응답 {"RESULT":{...}} 프로브.
type resultEnvelope struct {
	Result *struct {
		Code    string `json:"CODE"`
		Message string `json:"MESSAGE"`
	} `json:"RESULT"`
}

// Get 은 path segment GET 후 out 으로 디코드한다.
// 응답이 {"RESULT":...} 실패 형태면 INFO-200→ErrNoData, 그 외→*APIError.
// 반환하는 어떤 에러에도 인증키가 노출되지 않는다.
func (c *Client) Get(ctx context.Context, service string, start, end int, args []string, out any) error {
	path, err := c.buildPath(service, start, end, args)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("ecos: %s: %s", service, c.mask(err))
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("ecos: GET %s: %s", service, c.mask(err))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ecos: GET %s: %s", service, c.mask(err))
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ecos: GET %s: http status %d", service, resp.StatusCode)
	}
	var env resultEnvelope
	if json.Unmarshal(body, &env) == nil && env.Result != nil {
		if env.Result.Code == "INFO-200" {
			return ErrNoData
		}
		return &APIError{Code: env.Result.Code, Message: env.Result.Message}
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("ecos: decode %s: %s", service, c.mask(err))
	}
	return nil
}

// mask 는 에러 메시지에서 인증키를 지운다 (*url.Error 는 전체 URL 을 담으므로 필수).
func (c *Client) mask(err error) string {
	return strings.ReplaceAll(err.Error(), c.apiKey, "{API_KEY}")
}
```

import 블록에 `"context"`, `"encoding/json"`, `"errors"`, `"io"` 추가.

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test ./internal/httpclient/`
Expected: PASS (전 케이스)

- [ ] **Step 5: 커밋**

```bash
git add internal/httpclient/
git commit -m "feat: httpclient Get — RESULT 에러 매핑(INFO-200→ErrNoData)·인증키 마스킹"
```

---

### Task 4: 패키지 루트 골격 — errors/types/config/client/from_env (TDD)

**Files:**
- Create: `errors.go`, `types.go`, `config.go`, `client.go`, `from_env.go`, `client_test.go`

- [ ] **Step 1: 실패하는 테스트 작성** — `client_test.go`

```go
package ecos

import (
	"testing"
)

func TestNewClientRequiresKey(t *testing.T) {
	if _, err := NewClient(""); err == nil {
		t.Fatal("want error for empty apiKey")
	}
}

func TestNewClientDefaults(t *testing.T) {
	c, err := NewClient("KEY")
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("nil client")
	}
}

func TestNewClientFromEnv(t *testing.T) {
	t.Setenv("ECOS_API_KEY", "ENVKEY")
	if _, err := NewClientFromEnv(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ECOS_API_KEY", "")
	if _, err := NewClientFromEnv(); err == nil {
		t.Fatal("want error when ECOS_API_KEY unset")
	}
}

func TestPageOrDefault(t *testing.T) {
	s, e := (Page{}).orDefault()
	if s != 1 || e != 100 {
		t.Errorf("zero Page = %d..%d, want 1..100", s, e)
	}
	s, e = (Page{Start: 5, End: 20}).orDefault()
	if s != 5 || e != 20 {
		t.Errorf("Page{5,20} = %d..%d", s, e)
	}
}

func TestParseFloat(t *testing.T) {
	got, err := parseFloat("2.5")
	if err != nil || got != 2.5 {
		t.Errorf("parseFloat(2.5) = %v, %v", got, err)
	}
	got, err = parseFloat("1,397,923")
	if err != nil || got != 1397923 {
		t.Errorf("parseFloat(1,397,923) = %v, %v", got, err)
	}
	if _, err = parseFloat(""); err == nil {
		t.Error("want error for empty string")
	}
}
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test .`
Expected: FAIL (컴파일 에러)

- [ ] **Step 3: 구현**

`errors.go`:

```go
package ecos

import "github.com/kenshin579/ecos-go/internal/httpclient"

// APIError 는 ECOS RESULT 실패 응답. errors.As 로 Code/Message 접근.
type APIError = httpclient.APIError

// ErrNoData 는 INFO-200 (해당하는 데이터가 없습니다).
var ErrNoData = httpclient.ErrNoData
```

`types.go`:

```go
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
```

`config.go`:

```go
package ecos

import (
	"net/http"
	"time"
)

type clientOptions struct {
	baseURL    string
	lang       string
	timeout    time.Duration
	httpClient *http.Client
}

// Option 은 functional option.
type Option func(*clientOptions)

// WithBaseURL 은 API 베이스 URL 을 지정한다 (테스트/프록시용).
func WithBaseURL(u string) Option { return func(o *clientOptions) { o.baseURL = u } }

// WithTimeout 은 HTTP 타임아웃을 지정한다 (기본 30s).
func WithTimeout(d time.Duration) Option { return func(o *clientOptions) { o.timeout = d } }

// WithHTTPClient 는 사용자 정의 *http.Client 를 주입한다.
func WithHTTPClient(c *http.Client) Option { return func(o *clientOptions) { o.httpClient = c } }

// WithLanguage 는 언어구분을 지정한다 ("kr" 또는 "en", 기본 "kr").
func WithLanguage(lang string) Option { return func(o *clientOptions) { o.lang = lang } }
```

`client.go`:

```go
// Package ecos 는 한국은행 ECOS OpenAPI 의 Go 클라이언트다.
package ecos

import (
	"errors"
	"time"

	"github.com/kenshin579/ecos-go/internal/httpclient"
)

// Client 는 ecos 라이브러리의 단일 진입점.
type Client struct {
	http *httpclient.Client
}

// NewClient 는 API 키로 Client 를 만든다.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("ecos: apiKey is required")
	}
	cfg := clientOptions{timeout: 30 * time.Second}
	for _, opt := range opts {
		opt(&cfg)
	}
	hc := httpclient.New(httpclient.Config{
		APIKey:     apiKey,
		BaseURL:    cfg.baseURL,
		Lang:       cfg.lang,
		Timeout:    cfg.timeout,
		HTTPClient: cfg.httpClient,
	})
	return &Client{http: hc}, nil
}
```

`from_env.go`:

```go
package ecos

import (
	"errors"
	"os"
)

// NewClientFromEnv 는 ECOS_API_KEY 환경변수로 Client 를 만든다.
func NewClientFromEnv(opts ...Option) (*Client, error) {
	key := os.Getenv("ECOS_API_KEY")
	if key == "" {
		return nil, errors.New("ecos: ECOS_API_KEY is not set")
	}
	return NewClient(key, opts...)
}
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test ./...`
Expected: PASS (`listBody`/`Cycle` 은 아직 미사용 — `go vet` 이 unused 로 잡지 않는 타입 선언이므로 OK)

- [ ] **Step 5: 커밋**

```bash
git add errors.go types.go config.go client.go from_env.go client_test.go
git commit -m "feat: ecos 패키지 골격 — Client·Option·Cycle·Page·에러 재노출"
```

---

### Task 5: StatisticTableList (TDD)

**Files:**
- Create: `table.go`, `table_test.go`

- [ ] **Step 1: 실패하는 테스트 작성** — `table_test.go`

```go
package ecos

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestClient 는 요청 경로를 검증하고 testdata fixture 를 서빙하는 Client 를 만든다.
// wantPath 가 비어 있지 않으면 요청 경로와 정확히 일치해야 한다.
func newTestClient(t *testing.T, wantPath, fixture string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wantPath != "" && r.URL.EscapedPath() != wantPath {
			t.Errorf("request path = %q, want %q", r.URL.EscapedPath(), wantPath)
		}
		b, err := os.ReadFile(filepath.Join("testdata", fixture))
		if err != nil {
			t.Fatal(err)
		}
		w.Write(b)
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient("TESTKEY", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestStatisticTableList(t *testing.T) {
	c := newTestClient(t, "/api/StatisticTableList/TESTKEY/json/kr/1/10/", "table_list.json")
	got, err := c.StatisticTableList(context.Background(), TableListParams{Page: Page{1, 10}})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 834 {
		t.Errorf("TotalCount = %d, want 834", got.TotalCount)
	}
	if len(got.Rows) == 0 {
		t.Fatal("no rows")
	}
	r := got.Rows[0]
	if r.PStatCode != "*" || r.StatCode != "0000000001" || !strings.Contains(r.StatName, "통화") {
		t.Errorf("row[0] = %+v", r)
	}
	if r.SrchYN != "N" || r.Cycle != "" { // CYCLE null → zero value
		t.Errorf("row[0] cycle/srch = %+v", r)
	}
}

func TestStatisticTableListWithStatCode(t *testing.T) {
	c := newTestClient(t, "/api/StatisticTableList/TESTKEY/json/kr/1/5/102Y004", "table_list_code.json")
	got, err := c.StatisticTableList(context.Background(), TableListParams{Page: Page{1, 5}, StatCode: "102Y004"})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 1 || got.Rows[0].StatCode != "102Y004" || got.Rows[0].Cycle != "M" {
		t.Errorf("got = %+v", got)
	}
}
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test . -run TestStatisticTableList`
Expected: FAIL (컴파일 에러 — `TableListParams` 미정의)

- [ ] **Step 3: 구현** — `table.go`

```go
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
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test . -run TestStatisticTableList`
Expected: PASS (2 케이스)

- [ ] **Step 5: 커밋**

```bash
git add table.go table_test.go
git commit -m "feat: StatisticTableList(통계표목록) 구현"
```

---

### Task 6: StatisticWord (TDD)

**Files:**
- Create: `word.go`, `word_test.go`

- [ ] **Step 1: 실패하는 테스트 작성** — `word_test.go`

```go
package ecos

import (
	"context"
	"strings"
	"testing"
)

func TestStatisticWord(t *testing.T) {
	c := newTestClient(t,
		"/api/StatisticWord/TESTKEY/json/kr/1/5/%EC%86%8C%EB%B9%84%EC%9E%90%EB%8F%99%ED%96%A5%EC%A7%80%EC%88%98",
		"word.json")
	got, err := c.StatisticWord(context.Background(), WordParams{Page: Page{1, 5}, Word: "소비자동향지수"})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 1 || got.Rows[0].Word != "소비자동향지수" || !strings.Contains(got.Rows[0].Content, "미시간대학") {
		t.Errorf("got = %+v", got)
	}
}

func TestStatisticWordRequiresWord(t *testing.T) {
	c, _ := NewClient("KEY")
	if _, err := c.StatisticWord(context.Background(), WordParams{}); err == nil {
		t.Fatal("want error for empty Word")
	}
}
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test . -run TestStatisticWord`
Expected: FAIL (컴파일 에러)

- [ ] **Step 3: 구현** — `word.go`

```go
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
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test . -run TestStatisticWord`
Expected: PASS

- [ ] **Step 5: 커밋**

```bash
git add word.go word_test.go
git commit -m "feat: StatisticWord(통계용어사전) 구현"
```

---

### Task 7: StatisticItemList (TDD)

**Files:**
- Create: `item.go`, `item_test.go`

- [ ] **Step 1: 실패하는 테스트 작성** — `item_test.go`

```go
package ecos

import (
	"context"
	"testing"
)

func TestStatisticItemList(t *testing.T) {
	c := newTestClient(t, "/api/StatisticItemList/TESTKEY/json/kr/1/10/102Y004", "item_list.json")
	got, err := c.StatisticItemList(context.Background(), ItemListParams{Page: Page{1, 10}, StatCode: "102Y004"})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 6 {
		t.Errorf("TotalCount = %d, want 6", got.TotalCount)
	}
	r := got.Rows[0]
	if r.StatCode != "102Y004" || r.GrpCode != "Group1" || r.ItemCode != "ABA1" {
		t.Errorf("row[0] = %+v", r)
	}
	if r.Cycle != "A" || r.StartTime != "2003" || r.EndTime != "2025" || r.DataCnt != 23 {
		t.Errorf("row[0] cycle/기간 = %+v", r)
	}
	if r.UnitName != "십억원" || r.PItemCode != "" || r.Weight != "" {
		t.Errorf("row[0] 단위/null 필드 = %+v", r)
	}
}

func TestStatisticItemListRequiresStatCode(t *testing.T) {
	c, _ := NewClient("KEY")
	if _, err := c.StatisticItemList(context.Background(), ItemListParams{}); err == nil {
		t.Fatal("want error for empty StatCode")
	}
}
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test . -run TestStatisticItemList`
Expected: FAIL (컴파일 에러)

- [ ] **Step 3: 구현** — `item.go`

```go
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
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test . -run TestStatisticItemList`
Expected: PASS

주의: `WEIGHT` 가 실 응답에서 JSON number 로 오는 경우가 발견되면 string 디코드가 실패한다.
그 경우 `json.Number` 로 바꾸지 말고 fixture 를 확인해 실측 우선으로 타입을 정하고 보고하라
(현재 fixture 는 null 만 존재).

- [ ] **Step 5: 커밋**

```bash
git add item.go item_test.go
git commit -m "feat: StatisticItemList(통계세부항목목록) 구현"
```

---

### Task 8: StatisticSearch + Float64 헬퍼 (TDD)

**Files:**
- Create: `search.go`, `search_test.go`

- [ ] **Step 1: 실패하는 테스트 작성** — `search_test.go`

```go
package ecos

import (
	"context"
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
		{Cycle: CycleDaily, Start: "20260101", End: "20260131"},        // StatCode 없음
		{StatCode: "722Y001", Start: "20260101", End: "20260131"},      // Cycle 없음
		{StatCode: "722Y001", Cycle: CycleDaily, End: "20260131"},      // Start 없음
		{StatCode: "722Y001", Cycle: CycleDaily, Start: "20260101"},    // End 없음
	}
	for i, p := range cases {
		if _, err := c.StatisticSearch(ctx, p); err == nil {
			t.Errorf("case %d: want validation error", i)
		}
	}
}
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test . -run TestStatisticSearch`
Expected: FAIL (컴파일 에러)

- [ ] **Step 3: 구현** — `search.go`

```go
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
	UnitName  string `json:"UNIT_NAME"` // 단위
	Weight    string `json:"WGT"`       // 가중치 (없으면 null→"")
	Time      string `json:"TIME"`      // 시점 (주기별 날짜 형식)
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
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test . -run TestStatisticSearch`
Expected: PASS

- [ ] **Step 5: 커밋**

```bash
git add search.go search_test.go
git commit -m "feat: StatisticSearch(통계조회)·Float64 헬퍼 구현"
```

---

### Task 9: StatisticSearchAll — 자동 페이지네이션 (TDD)

**Files:**
- Modify: `search.go`, `search_test.go`

- [ ] **Step 1: 실패하는 테스트 추가** — `search_test.go` 에 append

```go
func TestStatisticSearchAll(t *testing.T) {
	// 총 5행을 chunk 2 로 3번에 나눠 서빙하는 mock.
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
	// 테스트에서 chunk 를 줄이기 위해 내부 상수 대신 파라미터화하지 않는다 —
	// searchAllChunk 는 프로덕션 상수이므로, 여기서는 전체 5행 < chunk 라 1회 호출로 끝나는
	// 케이스와, chunk 를 넘는 케이스를 나눠 검증한다.
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
```

`search_test.go` import 블록에 `"net/http"`, `"net/http/httptest"`, `"strconv"`, `"strings"` 추가.

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test . -run TestStatisticSearchAll`
Expected: FAIL (컴파일 에러 — `StatisticSearchAll` 미정의)

- [ ] **Step 3: 구현** — `search.go` 에 append

```go
// searchAllChunk 는 StatisticSearchAll 의 페이지 크기.
const searchAllChunk = 1000

// StatisticSearchAll 은 list_total_count 기준으로 전 페이지를 자동 수집한다.
// p.Page 는 무시된다. 페이지 경계에서 INFO-200 이 오면 수집분을 반환한다.
func (c *Client) StatisticSearchAll(ctx context.Context, p SearchParams) ([]SearchRow, error) {
	var all []SearchRow
	for start := 1; ; start += searchAllChunk {
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
	}
}
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test . -run TestStatisticSearchAll`
Expected: PASS (2 케이스)

- [ ] **Step 5: 커밋**

```bash
git add search.go search_test.go
git commit -m "feat: StatisticSearchAll — list_total_count 기반 자동 페이지네이션"
```

---

### Task 10: KeyStatisticList (TDD)

**Files:**
- Create: `keystat.go`, `keystat_test.go`

- [ ] **Step 1: 실패하는 테스트 작성** — `keystat_test.go`

```go
package ecos

import (
	"context"
	"testing"
)

func TestKeyStatisticList(t *testing.T) {
	c := newTestClient(t, "/api/KeyStatisticList/TESTKEY/json/kr/1/10/", "keystat.json")
	got, err := c.KeyStatisticList(context.Background(), KeyStatisticListParams{Page: Page{1, 10}})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 101 || got.RowCount != 10 {
		t.Errorf("TotalCount=%d RowCount=%d, want 101/10", got.TotalCount, got.RowCount)
	}
	r := got.Rows[0]
	if r.ClassName != "통화량" || r.KeyStatName != "M1(협의통화, 평잔)" {
		t.Errorf("row[0] = %+v", r)
	}
	if r.Cycle != "202605" || r.UnitName != "십억원" {
		t.Errorf("row[0] cycle/unit = %+v", r)
	}
	f, err := r.Float64()
	if err != nil || f != 1397923 {
		t.Errorf("Float64() = %v, %v", f, err)
	}
}
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test . -run TestKeyStatisticList`
Expected: FAIL (컴파일 에러)

- [ ] **Step 3: 구현** — `keystat.go`

```go
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
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test . -run TestKeyStatisticList`
Expected: PASS

- [ ] **Step 5: 커밋**

```bash
git add keystat.go keystat_test.go
git commit -m "feat: KeyStatisticList(100대통계지표) 구현 — row_count 포함"
```

---

### Task 11: StatisticMeta (TDD)

**Files:**
- Create: `meta.go`, `meta_test.go`

- [ ] **Step 1: 실패하는 테스트 작성** — `meta_test.go`

```go
package ecos

import (
	"context"
	"testing"
)

func TestStatisticMeta(t *testing.T) {
	c := newTestClient(t,
		"/api/StatisticMeta/TESTKEY/json/kr/1/10/%EA%B2%BD%EC%A0%9C%EC%8B%AC%EB%A6%AC%EC%A7%80%EC%88%98",
		"meta.json")
	got, err := c.StatisticMeta(context.Background(), MetaParams{Page: Page{1, 10}, Name: "경제심리지수"})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount != 53 {
		t.Errorf("TotalCount = %d, want 53", got.TotalCount)
	}
	r := got.Rows[0]
	if r.Lvl != "2" || r.ContCode != "N13" || r.ContName != "참고 자료" || r.MetaData != "" {
		t.Errorf("row[0] = %+v", r)
	}
}

func TestStatisticMetaRequiresName(t *testing.T) {
	c, _ := NewClient("KEY")
	if _, err := c.StatisticMeta(context.Background(), MetaParams{}); err == nil {
		t.Fatal("want error for empty Name")
	}
}
```

- [ ] **Step 2: 테스트 실패 확인**

Run: `go test . -run TestStatisticMeta`
Expected: FAIL (컴파일 에러)

- [ ] **Step 3: 구현** — `meta.go`

```go
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
```

- [ ] **Step 4: 테스트 통과 확인**

Run: `go test ./...`
Expected: PASS (전체 — httpclient 포함)

- [ ] **Step 5: 커밋**

```bash
git add meta.go meta_test.go
git commit -m "feat: StatisticMeta(통계메타DB) 구현"
```

---

### Task 12: 통합 테스트 (-tags integration)

**Files:**
- Create: `integration_test.go`

- [ ] **Step 1: 통합 테스트 작성** — `integration_test.go`

```go
//go:build integration

package ecos

import (
	"context"
	"errors"
	"testing"
	"time"
)

// 실 API 스모크. 실행: ECOS_API_KEY=... go test -tags integration ./...
func newIntegrationClient(t *testing.T) *Client {
	t.Helper()
	c, err := NewClientFromEnv()
	if err != nil {
		t.Skip("ECOS_API_KEY not set")
	}
	return c
}

func ctxWithTimeout(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestIntegrationStatisticTableList(t *testing.T) {
	c := newIntegrationClient(t)
	got, err := c.StatisticTableList(ctxWithTimeout(t), TableListParams{Page: Page{1, 5}})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount < 800 || len(got.Rows) != 5 {
		t.Errorf("total=%d rows=%d", got.TotalCount, len(got.Rows))
	}
}

func TestIntegrationStatisticWord(t *testing.T) {
	c := newIntegrationClient(t)
	got, err := c.StatisticWord(ctxWithTimeout(t), WordParams{Word: "소비자동향지수"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) == 0 || got.Rows[0].Word == "" {
		t.Errorf("rows = %+v", got.Rows)
	}
}

func TestIntegrationStatisticItemList(t *testing.T) {
	c := newIntegrationClient(t)
	got, err := c.StatisticItemList(ctxWithTimeout(t), ItemListParams{StatCode: "102Y004"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) == 0 || got.Rows[0].ItemCode == "" {
		t.Errorf("rows = %+v", got.Rows)
	}
}

func TestIntegrationStatisticSearch(t *testing.T) {
	c := newIntegrationClient(t)
	got, err := c.StatisticSearch(ctxWithTimeout(t), SearchParams{
		StatCode: "722Y001", Cycle: CycleDaily,
		Start: "20260101", End: "20260110", ItemCode1: "0101000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) == 0 {
		t.Fatal("no rows")
	}
	if _, err := got.Rows[0].Float64(); err != nil {
		t.Errorf("Float64: %v", err)
	}
}

func TestIntegrationStatisticSearchNoData(t *testing.T) {
	c := newIntegrationClient(t)
	_, err := c.StatisticSearch(ctxWithTimeout(t), SearchParams{
		StatCode: "722Y001", Cycle: CycleDaily,
		Start: "20991230", End: "20991231", ItemCode1: "0101000",
	})
	if !errors.Is(err, ErrNoData) {
		t.Errorf("err = %v, want ErrNoData", err)
	}
}

func TestIntegrationKeyStatisticList(t *testing.T) {
	c := newIntegrationClient(t)
	got, err := c.KeyStatisticList(ctxWithTimeout(t), KeyStatisticListParams{Page: Page{1, 10}})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalCount < 100 || got.RowCount != 10 {
		t.Errorf("total=%d rowCount=%d", got.TotalCount, got.RowCount)
	}
}

func TestIntegrationStatisticMeta(t *testing.T) {
	c := newIntegrationClient(t)
	got, err := c.StatisticMeta(ctxWithTimeout(t), MetaParams{Name: "경제심리지수"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) == 0 {
		t.Fatal("no rows")
	}
}
```

- [ ] **Step 2: 통합 테스트 실행**

```bash
export ECOS_API_KEY=$(grep -m1 'export ECOS_API_KEY' ~/.zshrc | cut -d= -f2 | tr -d '"' | tr -d "'")
go test -tags integration -run TestIntegration -v ./...
```

Expected: 7개 전부 PASS (실 API — 네트워크 필요)

- [ ] **Step 3: 유닛 테스트에 영향 없는지 확인**

Run: `go test ./...`
Expected: PASS (integration 태그 없이는 integration_test.go 제외됨)

- [ ] **Step 4: 커밋**

```bash
git add integration_test.go
git commit -m "test: 실 API 통합 테스트 추가 (-tags integration)"
```

---

### Task 13: 예제 + README

**Files:**
- Create: `examples/basic/main.go`, `README.md`

- [ ] **Step 1: 예제 작성** — `examples/basic/main.go`

```go
// 기준금리·100대 통계지표 조회 예제.
//
//	ECOS_API_KEY=... go run ./examples/basic
package main

import (
	"context"
	"fmt"
	"log"

	ecos "github.com/kenshin579/ecos-go"
)

func main() {
	client, err := ecos.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// 한국은행 기준금리 (일별, 최근 열흘치 예시)
	res, err := client.StatisticSearch(ctx, ecos.SearchParams{
		StatCode: "722Y001", Cycle: ecos.CycleDaily,
		Start: "20260101", End: "20260110", ItemCode1: "0101000",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range res.Rows {
		v, _ := r.Float64()
		fmt.Printf("%s  %s = %.2f%s\n", r.Time, r.ItemName1, v, r.UnitName)
	}

	// 100대 통계지표 상위 5개
	keys, err := client.KeyStatisticList(ctx, ecos.KeyStatisticListParams{Page: ecos.Page{Start: 1, End: 5}})
	if err != nil {
		log.Fatal(err)
	}
	for _, k := range keys.Rows {
		fmt.Printf("[%s] %s = %s%s (%s)\n", k.ClassName, k.KeyStatName, k.DataValue, k.UnitName, k.Cycle)
	}
}
```

- [ ] **Step 2: 예제 빌드·실행 확인**

```bash
go build ./...
export ECOS_API_KEY=$(grep -m1 'export ECOS_API_KEY' ~/.zshrc | cut -d= -f2 | tr -d '"' | tr -d "'")
go run ./examples/basic
```

Expected: 기준금리 값들과 지표 5개 출력.

- [ ] **Step 3: README 작성** — `README.md` (기존 빈 파일 덮어쓰기)

````markdown
# ecos-go

한국은행 경제통계시스템(ECOS) OpenAPI 의 Go 클라이언트 라이브러리.

## 설치

```bash
go get github.com/kenshin579/ecos-go@latest
```

Go 1.25+, 외부 의존성 없음(표준 라이브러리만).

## 사용

```go
client, _ := ecos.NewClientFromEnv() // ECOS_API_KEY
ctx := context.Background()

// 한국은행 기준금리 (일별)
res, _ := client.StatisticSearch(ctx, ecos.SearchParams{
    StatCode: "722Y001", Cycle: ecos.CycleDaily,
    Start: "20260101", End: "20260131", ItemCode1: "0101000",
})
for _, r := range res.Rows {
    v, _ := r.Float64()
    fmt.Println(r.Time, v, r.UnitName)
}

// 전 페이지 자동 수집
rows, _ := client.StatisticSearchAll(ctx, ecos.SearchParams{
    StatCode: "722Y001", Cycle: ecos.CycleDaily,
    Start: "20200101", End: "20261231", ItemCode1: "0101000",
})

// 100대 통계지표
keys, _ := client.KeyStatisticList(ctx, ecos.KeyStatisticListParams{})
```

## 인증

https://ecos.bok.or.kr/api/ 에서 발급받은 인증키를 `ECOS_API_KEY` 환경변수로 두거나
`ecos.NewClient(apiKey)` 로 전달한다.

## 옵션

```go
client, _ := ecos.NewClient(apiKey,
    ecos.WithTimeout(10*time.Second),  // HTTP 타임아웃 (기본 30s)
    ecos.WithLanguage("en"),           // 언어구분 kr/en (기본 kr)
    ecos.WithBaseURL("https://..."),   // 베이스 URL 교체 (테스트/프록시)
    ecos.WithHTTPClient(custom),       // *http.Client 주입
)
```

## 커버리지

ECOS OpenAPI 6개 서비스 전체.

| 서비스 | 메서드 | 설명 |
|--------|--------|------|
| StatisticTableList | `client.StatisticTableList` | 통계표 목록 |
| StatisticWord | `client.StatisticWord` | 통계용어사전 |
| StatisticItemList | `client.StatisticItemList` | 통계 세부항목 목록 |
| StatisticSearch | `client.StatisticSearch` / `StatisticSearchAll` | 통계 데이터 조회 |
| KeyStatisticList | `client.KeyStatisticList` | 100대 통계지표 |
| StatisticMeta | `client.StatisticMeta` | 통계메타DB |

요청·응답 명세는 [`docs/api/`](docs/api/README.md) 참고 (실 API 응답 대조로 작성).

## 에러 처리

- `errors.Is(err, ecos.ErrNoData)` — 데이터 없음 (INFO-200)
- `errors.As(err, &apiErr)` — 그 외 ECOS 오류 (`*ecos.APIError`, Code/Message)

에러 메시지에 인증키는 노출되지 않는다.

## 예제

```bash
ECOS_API_KEY=... go run ./examples/basic
```

## License

Released under the [MIT License](LICENSE).
````

- [ ] **Step 4: 커밋**

```bash
git add examples/ README.md
git commit -m "docs: 사용 예제 및 README 작성"
```

---

### Task 14: release.sh + 최종 검증

**Files:**
- Create: `scripts/release.sh`

- [ ] **Step 1: release.sh 복사·수정**

```bash
mkdir -p scripts
cp ../opendart-go/scripts/release.sh scripts/release.sh
# 스크립트 내 문자열 치환: opendart → ecos
sed -i '' 's/opendart-release-check/ecos-release-check/; s/release.sh — opendart 릴리스/release.sh — ecos-go 릴리스/' scripts/release.sh
chmod +x scripts/release.sh
```

- [ ] **Step 2: 최종 검증 일괄 실행**

```bash
gofmt -l .            # 출력 없어야 함
go build ./...
go vet ./...
go test ./...
export ECOS_API_KEY=$(grep -m1 'export ECOS_API_KEY' ~/.zshrc | cut -d= -f2 | tr -d '"' | tr -d "'")
go test -tags integration ./...
```

Expected: 전부 PASS, gofmt 출력 없음.

- [ ] **Step 3: 커밋**

```bash
git add scripts/release.sh
git commit -m "chore: release.sh 추가 (opendart-go 재사용)"
```

- [ ] **Step 4: 사용자 검수 요청**

전체 커밋 목록·테스트 결과를 정리해 사용자에게 보고하고 PR 생성 여부를 확인한다.
v0.1.0 릴리스(`./scripts/release.sh v0.1.0`)는 **PR 이 main 에 머지된 후** main 에서 실행한다.
