package ecos

import "github.com/kenshin579/ecos-go/internal/httpclient"

// APIError 는 ECOS RESULT 실패 응답. errors.As 로 Code/Message 접근.
type APIError = httpclient.APIError

// ErrNoData 는 INFO-200 (해당하는 데이터가 없습니다).
var ErrNoData = httpclient.ErrNoData
