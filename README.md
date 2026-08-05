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
