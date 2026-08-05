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
