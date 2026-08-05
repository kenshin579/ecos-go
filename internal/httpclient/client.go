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
