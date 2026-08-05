// Package httpclient 는 ECOS OpenAPI REST 호출의 단일 GET 통로다.
// path segment URL 조립, RESULT 에러 매핑, 인증키 마스킹을 담당한다.
package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	if c.apiKey == "" { // ReplaceAll 의 빈 문자열 치환(모든 문자 사이 삽입) 방어
		return err.Error()
	}
	return strings.ReplaceAll(err.Error(), c.apiKey, "{API_KEY}")
}
