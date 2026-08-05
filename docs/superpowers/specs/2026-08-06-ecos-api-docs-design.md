# ecos-go 설계 — 1단계: API 명세 문서 작성

- 날짜: 2026-08-06
- 상태: 승인됨 (사용자 확인)
- 이번 작업 범위: **API 명세 문서(md) 작성에 집중**. 라이브러리 구현(2단계)은 문서 검수 완료 후 별도 플랜으로 진행한다.

## 배경

한국은행 ECOS OpenAPI(https://ecos.bok.or.kr/api/#/)의 Go 클라이언트 라이브러리 `github.com/kenshin579/ecos-go`를 만든다. fmp-go·opendart-go 자매 프로젝트다. opendart-go 방식대로 **API 명세를 md 문서로 먼저 정리하고, 그 문서를 근거로 라이브러리를 구현**한다.

opendart-go는 문서 사이트가 서버렌더 HTML이라 크롤러(`scripts/crawl`)로 124개 문서를 자동 생성했지만, ECOS 문서 사이트는 React SPA(내부 POST 프로토콜)라 크롤링 비용이 크다. ECOS는 서비스가 **6개뿐**이므로 크롤러 없이 **공식 명세 + 실 API 호출 응답을 대조해 문서를 직접 작성**한다(사용자 합의: 방식 A).

## 대상 API — ECOS OpenAPI 서비스 6개

사용자 API 키(`ECOS_API_KEY`, `~/.zshrc`)로 6개 서비스 전부 실 호출 검증 완료(2026-08-06).

| 서비스명 | 설명 | 문서 파일 |
|----------|------|-----------|
| StatisticTableList | 서비스 통계(통계표) 목록 — 부모·자식 트리 구조 | `docs/api/통계표목록.md` |
| StatisticWord | 통계용어사전 | `docs/api/통계용어사전.md` |
| StatisticItemList | 통계 세부항목 목록 | `docs/api/통계세부항목목록.md` |
| StatisticSearch | 통계 조회(데이터) — 핵심 서비스 | `docs/api/통계조회.md` |
| KeyStatisticList | 100대 통계지표 | `docs/api/100대통계지표.md` |
| StatisticMeta | 통계메타DB | `docs/api/통계메타DB.md` |

### API 공통 특성 (조사 결과)

- URL은 쿼리스트링이 아니라 **path segment 방식**:
  `https://ecos.bok.or.kr/api/{서비스명}/{인증키}/{요청유형}/{언어구분}/{요청시작건수}/{요청종료건수}/{서비스별 인자...}`
- 요청유형: `json` / `xml` (문서에는 둘 다 기재, 라이브러리는 JSON만 사용 예정)
- 언어구분: `kr` / `en`
- 페이지네이션: 시작건수/종료건수(건 수 기반) + 응답의 `list_total_count`
- 성공 응답: `{"서비스명": {"list_total_count": N, "row": [...]}}`
- 실패 응답: `{"RESULT": {"CODE": "ERROR-100", "MESSAGE": "..."}}` — 성공과 구조가 다름
- 주기 코드: A(년) · S(반년) · Q(분기) · M(월) · SM(반월) · D(일) — 주기별 날짜 형식이 다름(A: YYYY, Q: YYYYQn, M: YYYYMM, D: YYYYMMDD 등)
- 한글 인자(용어사전 검색어, 메타 통계명 등)는 URL path escape 필요

## 산출물 — 문서 구조

```
docs/api/
├── README.md              # 인덱스 + 공통 규약
├── 통계표목록.md            # StatisticTableList
├── 통계용어사전.md          # StatisticWord
├── 통계세부항목목록.md       # StatisticItemList
├── 통계조회.md             # StatisticSearch
├── 100대통계지표.md         # KeyStatisticList
└── 통계메타DB.md           # StatisticMeta
```

### README.md (인덱스) 내용

- 서비스 6개 표(서비스명 · 설명 · 문서 링크)
- 공통 규약:
  - 인증(키 발급, path 상 인증키 위치)
  - URL path segment 규칙(순서, 생략 규칙)
  - 요청유형·언어구분
  - 페이지네이션(시작건수/종료건수, `list_total_count`)
  - 주기 코드 ↔ 날짜 형식 대응표
  - **에러 코드 표 전체**(ERROR-xxx / INFO-xxx, 실 응답 포맷 포함)

### 각 서비스 문서 포맷 (opendart-go `docs/api` 포맷 준용)

1. 제목 + 요청 URL 한 줄 요약
2. **기본 정보**: 메서드 / URL 패턴 / 출력포맷(JSON·XML)
3. **요청 인자**: path segment 순서대로 표 — 위치 · 인자명 · 명칭 · 타입 · 필수여부 · 값 설명
4. **응답 필드**: 필드명 · 명칭 · 설명 (실 응답과 대조해 전수 기재)
5. **샘플**: 실 호출 요청 URL(인증키는 `{API_KEY}`로 마스킹) + 실 응답 JSON(일부 발췌)

## 검증 방법 (누락 방지)

문서 작성 시 서비스마다 다음을 수행한다:

1. **실 API 호출**로 응답 필드 전수 확인 — 공식 명세에만 의존하지 않고, 실제 JSON 응답의 키 목록과 문서의 응답 필드 표를 1:1 대조
2. **요청 인자 검증** — 필수 인자를 빼고 호출해 에러 확인, 선택 인자 생략 동작 확인
3. 공식 개발명세서(ECOS 사이트) 내용과 교차 확인
4. 작성 완료 후 **누락 점검 체크리스트** 실행:
   - [ ] 서비스 6개 문서가 모두 존재하는가
   - [ ] 각 문서의 요청 인자 표가 실 URL과 순서·개수 일치하는가
   - [ ] 각 문서의 응답 필드 표가 실 응답 JSON 키와 전수 일치하는가
   - [ ] 에러 코드 표가 공식 명세의 코드를 모두 포함하는가
   - [ ] README 인덱스 링크가 모두 유효한가
   - [ ] 파일 인코딩이 UTF-8인가 (`file -I`)

## 완료 기준 (이번 작업)

- `docs/api/` 문서 7개(README + 서비스 6개)가 작성되어 있다
- 누락 점검 체크리스트 전 항목 통과
- **사용자 검수**: 사용자가 문서를 확인하고 승인하면 → 2단계(라이브러리 구현) 플랜으로 진행

## (참고) 2단계 — 라이브러리 설계 합의 사항

문서 검수 후 별도 플랜으로 진행할 내용. 브레인스토밍에서 합의된 설계를 기록해 둔다.

- **구조: 단일 패키지 flat** (방식 A 합의) — `client.StatisticSearch(ctx, params)` 식. 서비스 6개라 fmp-go식 서브 클라이언트 계층은 두지 않는다.
- 파일 구성: `client.go` / `config.go` / `errors.go` / `types.go` / 서비스별 파일 6개 / `internal/httpclient` / `examples/` / `scripts/release.sh`
- 인증: `NewClientFromEnv()`(`ECOS_API_KEY`) 또는 `NewClient(apiKey)`
- 옵션: `WithBaseURL` / `WithTimeout`(기본 30s) / `WithHTTPClient` / `WithLanguage`(기본 kr)
- 요청유형은 JSON 고정, 외부 의존성 0(표준 라이브러리만), Go 1.25
- **편의 기능 포함** (방식 A 합의):
  - Cycle 상수: `CycleAnnual`(A) `CycleSemiAnnual`(S) `CycleQuarterly`(Q) `CycleMonthly`(M) `CycleSemiMonthly`(SM) `CycleDaily`(D)
  - 전체 조회 헬퍼: `StatisticSearchAll` 등 — `list_total_count` 기준 자동 페이지 반복
  - 숫자 파싱 헬퍼: `DATA_VALUE` 등에 `Float64()` — 원본 문자열 보존
  - 에러 매핑: `INFO-200`(데이터 없음) → `ErrNoData`, 그 외 → `*APIError{Code, Message}`; 에러 메시지에서 인증키 마스킹
- 테스트: `httptest` + `testdata/` 실 응답 fixture 유닛 테스트, `-tags integration` 실 호출 스모크
- 릴리스: `scripts/release.sh vX.Y.Z`(opendart-go 재사용) → v0.1.0
