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
