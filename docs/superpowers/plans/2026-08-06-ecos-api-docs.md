# ECOS API 명세 문서 작성 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 한국은행 ECOS OpenAPI 6개 서비스의 명세를 실 API 호출로 검증하며 `docs/api/*.md` 7개 문서(README 인덱스 + 서비스 6개)로 작성한다.

**Architecture:** 크롤러 없이 문서를 직접 작성한다. 먼저 실 API 호출로 fixture(응답 JSON)를 수집하고, 각 문서의 응답 필드 표를 fixture의 실제 키와 1:1 대조해 누락을 방지한다. 스펙: `docs/superpowers/specs/2026-08-06-ecos-api-docs-design.md`

**Tech Stack:** Markdown, curl + python3(json 검증), git. 코드 작성 없음(라이브러리 구현은 2단계 별도 플랜).

---

## 공통 준비 (모든 태스크에서 사용)

- 작업 디렉토리: `/Users/user/src/workspace_moneyflow/ecos-go` (브랜치 `feature/ecos-api-docs`)
- fixture 디렉토리: `/private/tmp/claude-501/-Users-user-src-workspace-moneyflow/14d0c594-3468-4ea5-8a0d-24728a1cd56c/scratchpad/ecos-fixtures` (없으면 Task 1이 생성)
- API 키 로드 (모든 셸 명령 앞에서):

```bash
K=$(grep -m1 'export ECOS_API_KEY' ~/.zshrc | cut -d= -f2 | tr -d '"' | tr -d "'")
FIX=/private/tmp/claude-501/-Users-user-src-workspace-moneyflow/14d0c594-3468-4ea5-8a0d-24728a1cd56c/scratchpad/ecos-fixtures
```

- **인증키 마스킹 규칙**: 문서에 넣는 URL 예시는 인증키 자리를 반드시 `{API_KEY}` 로 표기한다. 실제 키가 문서/커밋에 절대 들어가면 안 된다. 커밋 전 `grep -r "$K" docs/` 로 확인한다.
- **실측 우선 규칙**: 이 플랜의 표(요청 인자·응답 필드)는 사전 조사 결과다. fixture 실측과 다르면 **실측을 우선**하고 문서에 반영한다.

---

### Task 1: 실측 fixture 수집

**Files:**
- Create: `$FIX/*.json` (커밋하지 않음, scratchpad)

- [ ] **Step 1: fixture 수집 스크립트 실행 — 정상 케이스**

```bash
K=$(grep -m1 'export ECOS_API_KEY' ~/.zshrc | cut -d= -f2 | tr -d '"' | tr -d "'")
FIX=/private/tmp/claude-501/-Users-user-src-workspace-moneyflow/14d0c594-3468-4ea5-8a0d-24728a1cd56c/scratchpad/ecos-fixtures
mkdir -p "$FIX"
B="https://ecos.bok.or.kr/api"

# 1) StatisticTableList — 기본 / 통계표코드 지정
curl -s "$B/StatisticTableList/$K/json/kr/1/10/"        -o "$FIX/table_list.json"
curl -s "$B/StatisticTableList/$K/json/kr/1/5/102Y004"  -o "$FIX/table_list_code.json"
# 2) StatisticWord
curl -s "$B/StatisticWord/$K/json/kr/1/5/소비자동향지수"   -o "$FIX/word.json"
# 3) StatisticItemList
curl -s "$B/StatisticItemList/$K/json/kr/1/10/102Y004"  -o "$FIX/item_list.json"
# 4) StatisticSearch — D 주기(기준금리) / Q 주기(GDP) / 항목코드 생략
curl -s "$B/StatisticSearch/$K/json/kr/1/5/722Y001/D/20260101/20260131/0101000" -o "$FIX/search_d.json"
curl -s "$B/StatisticSearch/$K/json/kr/1/5/200Y102/Q/2023Q1/2023Q4/10111"       -o "$FIX/search_q.json"
curl -s "$B/StatisticSearch/$K/json/kr/1/5/722Y001/M/202601/202606/0101000"     -o "$FIX/search_m.json"
curl -s "$B/StatisticSearch/$K/json/kr/1/5/722Y001/A/2020/2025/0101000"         -o "$FIX/search_a.json"
# 통계항목코드 생략 시 동작 확인 (스펙: 선택 인자 생략 동작)
curl -s "$B/StatisticSearch/$K/json/kr/1/5/722Y001/D/20260101/20260131/"        -o "$FIX/search_no_item.json"
# 5) KeyStatisticList
curl -s "$B/KeyStatisticList/$K/json/kr/1/10/"          -o "$FIX/keystat.json"
# 6) StatisticMeta
curl -s "$B/StatisticMeta/$K/json/kr/1/10/경제심리지수"    -o "$FIX/meta.json"
# 7) 언어 en 확인 (표 목록)
curl -s "$B/StatisticTableList/$K/json/en/1/3/"         -o "$FIX/table_list_en.json"
ls -la "$FIX"
```

Expected: json 파일 13개 생성. `search_no_item.json` 의 결과(전체 항목 반환 여부)는 Task 6 문서의 "생략 시 전체 항목" 설명의 실측 근거가 된다 — 다르게 동작하면 문서를 실측에 맞게 수정.

- [ ] **Step 2: S(반년)·SM(반월) 주기 실측**

전체 통계표 목록에서 S/SM 주기 표를 찾아 실측한다.

```bash
curl -s "$B/StatisticTableList/$K/json/kr/1/834/" -o "$FIX/table_list_all.json"
python3 - "$FIX/table_list_all.json" <<'EOF'
import json, sys
rows = json.load(open(sys.argv[1]))["StatisticTableList"]["row"]
for cyc in ("S", "SM"):
    hit = [r for r in rows if r.get("CYCLE") == cyc and r.get("SRCH_YN") == "Y"]
    print(cyc, len(hit), [h["STAT_CODE"] for h in hit[:5]])
EOF
```

찾은 표 코드로 `StatisticItemList` → 항목코드 확인 → `StatisticSearch` 를 호출해 `search_s.json` / `search_sm.json` 저장. **S·SM 주기의 날짜 형식(예: `2023S1`, `202401SM1` 등)을 여기서 확정**하고 기록한다. S/SM 주기 표가 하나도 없으면 "ECOS 데이터에 S/SM 주기 표 없음"으로 기록하고 문서의 주기 표에 그대로 명시한다.

- [ ] **Step 3: 에러 케이스 fixture 수집**

```bash
# INFO-100 인증키 오류
curl -s "$B/KeyStatisticList/WRONGKEY123/json/kr/1/2/"                    -o "$FIX/err_info100.json"
# ERROR-100 필수값 누락 (StatisticSearch 인자 없이)
curl -s "$B/StatisticSearch/$K/json/kr/1/2/"                              -o "$FIX/err_error100.json"
# ERROR-101 주기-날짜 형식 불일치
curl -s "$B/StatisticSearch/$K/json/kr/1/2/722Y001/D/2026/2026/0101000"   -o "$FIX/err_error101.json"
# INFO-200 데이터 없음
curl -s "$B/StatisticSearch/$K/json/kr/1/2/722Y001/D/20991231/20991231/0101000" -o "$FIX/err_info200.json"
# ERROR-200 요청유형 오류 (json/xml 이외 값)
curl -s "$B/KeyStatisticList/$K/badtype/kr/1/2/"                          -o "$FIX/err_error200.json"
# ERROR-301 조회건수 타입 오류 (숫자 아님)
curl -s "$B/KeyStatisticList/$K/json/kr/abc/xyz/"                         -o "$FIX/err_error301.json"
for f in "$FIX"/err_*.json; do echo "--- $f"; cat "$f"; echo; done
```

Expected: 각 파일에 `{"RESULT":{"CODE":"...","MESSAGE":"..."}}`. 예상과 다른 코드가 나오면 실측 코드·메시지를 기록(문서에는 실측값을 쓴다).

- [ ] **Step 4: fixture 유효성 확인**

```bash
python3 - "$FIX" <<'EOF'
import json, os, sys
d = sys.argv[1]
for f in sorted(os.listdir(d)):
    if not f.endswith(".json"): continue
    data = json.load(open(os.path.join(d, f)))
    root = list(data.keys())[0]
    n = data[root].get("list_total_count") if isinstance(data[root], dict) else None
    print(f"{f}: root={root} total={n}")
EOF
```

Expected: 모든 파일이 json 파싱 성공. 정상 fixture 는 root 가 서비스명, 에러 fixture 는 root 가 `RESULT`.

---

### Task 2: docs/api/README.md (인덱스 + 공통 규약)

**Files:**
- Create: `docs/api/README.md`

- [ ] **Step 1: README.md 작성**

아래 내용으로 작성하되, **주기·날짜 형식 표의 S·SM 행과 에러 코드 표는 Task 1 실측 결과로 보정**한다.

````markdown
# 한국은행 ECOS OpenAPI 문서

한국은행 경제통계시스템(ECOS) OpenAPI(https://ecos.bok.or.kr/api/#/)의 API 명세입니다.
공식 개발명세서와 실 API 호출 응답을 대조하여 작성했습니다. (작성일 2026-08-06)

## 서비스 목록

| 서비스명 | 설명 | 문서 |
| --- | --- | --- |
| StatisticTableList | 서비스 통계(통계표) 목록 — 부모·자식 트리 구조 | [통계표목록](통계표목록.md) |
| StatisticWord | 통계용어사전 | [통계용어사전](통계용어사전.md) |
| StatisticItemList | 통계 세부항목 목록 | [통계세부항목목록](통계세부항목목록.md) |
| StatisticSearch | 통계 조회(데이터) — 핵심 서비스 | [통계조회](통계조회.md) |
| KeyStatisticList | 100대 통계지표 | [100대통계지표](100대통계지표.md) |
| StatisticMeta | 통계메타DB | [통계메타DB](통계메타DB.md) |

## 공통 규약

### 인증

- https://ecos.bok.or.kr/api/ 에서 인증키를 발급받는다.
- 인증키는 요청 URL 의 path segment 로 전달한다 (아래 URL 구조 참고).

### URL 구조 — path segment 방식

쿼리스트링이 아니라 **path segment** 로 인자를 전달한다.

```
https://ecos.bok.or.kr/api/{서비스명}/{인증키}/{요청유형}/{언어구분}/{요청시작건수}/{요청종료건수}/{서비스별 인자...}
```

| 위치 | 인자 | 필수 | 설명 |
| --- | --- | --- | --- |
| 1 | 서비스명 | Y | 위 서비스 목록의 서비스명 |
| 2 | 인증키 | Y | 발급받은 인증키 |
| 3 | 요청유형 | Y | `json` 또는 `xml` |
| 4 | 언어구분 | Y | `kr`(국문) 또는 `en`(영문) |
| 5 | 요청시작건수 | Y | 전체 결과 중 몇 번째부터 (1부터) |
| 6 | 요청종료건수 | Y | 전체 결과 중 몇 번째까지 |
| 7+ | 서비스별 인자 | 서비스별 | 각 서비스 문서 참고. 뒤쪽의 선택 인자는 생략 가능 |

- 한글 인자(용어, 통계명 등)는 URL 인코딩(percent-encoding)하여 전달한다.

### 페이지네이션

- 요청시작건수/요청종료건수는 **건수 기반**이다 (페이지 번호 아님). 예: `1/100` → 1~100번째.
- 응답의 `list_total_count` 가 전체 건수다. 이를 이용해 나눠서 조회한다.

### 응답 구조

성공 (JSON):

```json
{"서비스명": {"list_total_count": 834, "row": [ { ... }, ... ]}}
```

실패 — 성공과 구조가 다르다:

```json
{"RESULT": {"CODE": "ERROR-100", "MESSAGE": "필수 값이 누락되어 있습니다. ..."}}
```

### 주기 코드와 날짜 형식

| 주기 코드 | 의미 | 날짜 형식 | 예 |
| --- | --- | --- | --- |
| A | 년 | YYYY | `2025` |
| S | 반년 | (Task 1 실측 결과로 기재) | |
| Q | 분기 | YYYYQn | `2023Q1` |
| M | 월 | YYYYMM | `202601` |
| SM | 반월 | (Task 1 실측 결과로 기재) | |
| D | 일 | YYYYMMDD | `20260805` |

주기와 날짜 형식이 일치하지 않으면 `ERROR-101` 이 반환된다.

### 에러 코드

(Task 1 실측으로 확인한 코드는 실측 메시지를 기재하고, 실측 불가한 코드는 공식 명세 기준으로 기재 후 `(공식 명세)` 표기)

| 코드 | 메시지 | 비고 |
| --- | --- | --- |
| INFO-100 | 인증키가 유효하지 않습니다. ... | 실측 확인 |
| INFO-200 | 해당하는 데이터가 없습니다. | 실측 확인 |
| ERROR-100 | 필수 값이 누락되어 있습니다. ... | 실측 확인 |
| ERROR-101 | 주기와 다른 형식의 날짜 형식입니다. ... | 실측 확인 |
| ERROR-200 | 파일타입 값이 누락 혹은 유효하지 않습니다. | Task 1 실측 시도 |
| ERROR-300 | 조회건수 값이 누락되어 있습니다. | Task 1 실측 시도 |
| ERROR-301 | 조회건수 값의 타입이 유효하지 않습니다. | Task 1 실측 시도 |
| ERROR-400 | 검색범위가 적정범위를 초과하여 조회할 수 없습니다. | (공식 명세) |
| ERROR-500 | 서버 오류입니다. | (공식 명세) |
| ERROR-600 | DB Connection 오류입니다. | (공식 명세) |
| ERROR-601 | SQL 오류입니다. | (공식 명세) |
| ERROR-602 | 과도한 OpenAPI 호출로 이용이 제한되었습니다. | (공식 명세) |
````

- [ ] **Step 2: 인코딩·키 유출 확인 후 커밋**

```bash
file -I docs/api/README.md          # 기대: charset=utf-8
grep -rn "$K" docs/ && echo "키 유출!!" || echo "키 유출 없음"
git add docs/api/README.md
git commit -m "docs: ECOS API 문서 인덱스 및 공통 규약 추가"
```

---

### Task 3: 통계표목록.md (StatisticTableList)

**Files:**
- Create: `docs/api/통계표목록.md`

- [ ] **Step 1: fixture 와 필드 대조 준비**

```bash
python3 - "$FIX/table_list.json" <<'EOF'
import json, sys
d = json.load(open(sys.argv[1]))
print(sorted(d["StatisticTableList"]["row"][0].keys()))
EOF
```

Expected: `['CYCLE', 'ORG_NAME', 'P_STAT_CODE', 'SRCH_YN', 'STAT_CODE', 'STAT_NAME']` — 다르면 아래 문서 표를 실측에 맞게 수정.

- [ ] **Step 2: 문서 작성**

````markdown
# 통계표목록 (StatisticTableList)

> `GET https://ecos.bok.or.kr/api/StatisticTableList/{API_KEY}/{요청유형}/{언어구분}/{시작건수}/{종료건수}/{통계표코드}`

ECOS 에서 서비스되는 통계표 목록을 제공한다. 부모·자식(그룹·통계표) 트리 구조이며,
`SRCH_YN=Y` 인 행이 실제 데이터 조회(StatisticSearch)가 가능한 통계표다.

## 기본 정보

| 메서드 | 출력포맷 | 비고 |
| --- | --- | --- |
| GET | JSON, XML | 요청유형 path 인자로 선택 |

## 요청 인자 (path segment 순서)

| 위치 | 인자 | 필수 | 값 설명 |
| --- | --- | --- | --- |
| 1~6 | 공통 인자 | Y | [README 공통 규약](README.md) 참고 |
| 7 | 통계표코드 | N | 지정 시 해당 통계표만 조회 (예: `102Y004`) |

## 응답 필드 (`StatisticTableList.row[]`)

| 필드 | 명칭 | 설명 |
| --- | --- | --- |
| P_STAT_CODE | 상위 통계표코드 | 트리 부모 노드. 최상위는 `*` |
| STAT_CODE | 통계표코드 | |
| STAT_NAME | 통계명 | 번호 체계 포함 (예: `1.1.1.1.1. 본원통화 구성내역...`) |
| CYCLE | 주기 | A/S/Q/M/SM/D. 그룹 노드는 null |
| SRCH_YN | 검색 가능 여부 | `Y` 면 StatisticSearch 로 데이터 조회 가능 |
| ORG_NAME | 출처 | 그룹 노드는 null |

## 샘플

요청:

```
GET https://ecos.bok.or.kr/api/StatisticTableList/{API_KEY}/json/kr/1/10/
```

응답 (발췌 — fixture 에서 삽입):

```json
(Task 1 의 table_list.json 앞부분 발췌를 여기에 삽입)
```
````

- [ ] **Step 3: 문서 응답 표 ↔ fixture 키 전수 대조**

```bash
python3 - "$FIX/table_list.json" docs/api/통계표목록.md <<'EOF'
import json, re, sys
keys = set()
for r in json.load(open(sys.argv[1]))["StatisticTableList"]["row"]:
    keys |= set(r.keys())
doc = open(sys.argv[2]).read()
missing = [k for k in sorted(keys) if not re.search(rf"^\| {re.escape(k)} \|", doc, re.M)]
print("실응답 키:", sorted(keys))
print("문서 누락:", missing or "없음")
EOF
```

Expected: `문서 누락: 없음`

- [ ] **Step 4: 인코딩 확인 후 커밋**

```bash
file -I docs/api/통계표목록.md && grep -rn "$K" docs/ || echo "키 유출 없음"
git add docs/api/통계표목록.md
git commit -m "docs: StatisticTableList(통계표목록) API 문서 추가"
```

---

### Task 4: 통계용어사전.md (StatisticWord)

**Files:**
- Create: `docs/api/통계용어사전.md`

- [ ] **Step 1: fixture 필드 확인**

```bash
python3 -c "import json; print(sorted(json.load(open('$FIX/word.json'))['StatisticWord']['row'][0].keys()))"
```

Expected: `['CONTENT', 'WORD']`

- [ ] **Step 2: 문서 작성**

````markdown
# 통계용어사전 (StatisticWord)

> `GET https://ecos.bok.or.kr/api/StatisticWord/{API_KEY}/{요청유형}/{언어구분}/{시작건수}/{종료건수}/{용어}`

통계 용어의 정의를 검색한다.

## 기본 정보

| 메서드 | 출력포맷 | 비고 |
| --- | --- | --- |
| GET | JSON, XML | 요청유형 path 인자로 선택 |

## 요청 인자 (path segment 순서)

| 위치 | 인자 | 필수 | 값 설명 |
| --- | --- | --- | --- |
| 1~6 | 공통 인자 | Y | [README 공통 규약](README.md) 참고 |
| 7 | 용어 | Y | 검색할 용어 (예: `소비자동향지수`). URL 인코딩 필요 |

## 응답 필드 (`StatisticWord.row[]`)

| 필드 | 명칭 | 설명 |
| --- | --- | --- |
| WORD | 용어 | |
| CONTENT | 용어 설명 | |

## 샘플

요청:

```
GET https://ecos.bok.or.kr/api/StatisticWord/{API_KEY}/json/kr/1/5/소비자동향지수
```

응답 (발췌 — fixture 에서 삽입):

```json
(Task 1 의 word.json 발췌를 여기에 삽입)
```
````

- [ ] **Step 3: 대조 검증 + 커밋**

Task 3 Step 3 과 같은 python 대조 스크립트를 `word.json`/`StatisticWord`/`docs/api/통계용어사전.md` 에 대해 실행. Expected: `문서 누락: 없음`

```bash
file -I docs/api/통계용어사전.md
git add docs/api/통계용어사전.md
git commit -m "docs: StatisticWord(통계용어사전) API 문서 추가"
```

---

### Task 5: 통계세부항목목록.md (StatisticItemList)

**Files:**
- Create: `docs/api/통계세부항목목록.md`

- [ ] **Step 1: fixture 필드 확인**

```bash
python3 -c "import json; print(sorted(json.load(open('$FIX/item_list.json'))['StatisticItemList']['row'][0].keys()))"
```

Expected: `['CYCLE', 'DATA_CNT', 'END_TIME', 'GRP_CODE', 'GRP_NAME', 'ITEM_CODE', 'ITEM_NAME', 'P_ITEM_CODE', 'P_ITEM_NAME', 'START_TIME', 'STAT_CODE', 'STAT_NAME', 'UNIT_NAME', 'WEIGHT']`

- [ ] **Step 2: 문서 작성**

````markdown
# 통계세부항목목록 (StatisticItemList)

> `GET https://ecos.bok.or.kr/api/StatisticItemList/{API_KEY}/{요청유형}/{언어구분}/{시작건수}/{종료건수}/{통계표코드}`

통계표의 세부 항목(item) 목록을 제공한다. StatisticSearch 호출에 필요한
통계항목코드(`ITEM_CODE`)와 주기·수록기간을 여기서 확인한다.
같은 항목이 주기(A/M 등)별로 별도 행으로 반환된다.

## 기본 정보

| 메서드 | 출력포맷 | 비고 |
| --- | --- | --- |
| GET | JSON, XML | 요청유형 path 인자로 선택 |

## 요청 인자 (path segment 순서)

| 위치 | 인자 | 필수 | 값 설명 |
| --- | --- | --- | --- |
| 1~6 | 공통 인자 | Y | [README 공통 규약](README.md) 참고 |
| 7 | 통계표코드 | Y | 예: `102Y004` |

## 응답 필드 (`StatisticItemList.row[]`)

| 필드 | 명칭 | 설명 |
| --- | --- | --- |
| STAT_CODE | 통계표코드 | |
| STAT_NAME | 통계명 | |
| GRP_CODE | 항목 그룹코드 | 예: `Group1` — StatisticSearch 의 항목코드1~4 위치와 대응 |
| GRP_NAME | 항목 그룹명 | 예: `계정항목` |
| ITEM_CODE | 통계항목코드 | StatisticSearch 요청에 사용 |
| ITEM_NAME | 통계항목명 | |
| P_ITEM_CODE | 상위 통계항목코드 | 트리 구조. 최상위는 null |
| P_ITEM_NAME | 상위 통계항목명 | |
| CYCLE | 주기 | A/S/Q/M/SM/D |
| START_TIME | 수록 시작일자 | 주기별 날짜 형식 (예: A → `2003`, M → `200310`) |
| END_TIME | 수록 종료일자 | |
| DATA_CNT | 자료 수 | |
| UNIT_NAME | 단위 | 예: `십억원` |
| WEIGHT | 가중치 | 없으면 null |

## 샘플

요청:

```
GET https://ecos.bok.or.kr/api/StatisticItemList/{API_KEY}/json/kr/1/10/102Y004
```

응답 (발췌 — fixture 에서 삽입):

```json
(Task 1 의 item_list.json 발췌를 여기에 삽입)
```
````

- [ ] **Step 3: 대조 검증 + 커밋**

Task 3 Step 3 과 같은 대조 스크립트를 `item_list.json`/`StatisticItemList`/`docs/api/통계세부항목목록.md` 에 대해 실행. Expected: `문서 누락: 없음`

```bash
file -I docs/api/통계세부항목목록.md
git add docs/api/통계세부항목목록.md
git commit -m "docs: StatisticItemList(통계세부항목목록) API 문서 추가"
```

---

### Task 6: 통계조회.md (StatisticSearch) — 핵심 서비스

**Files:**
- Create: `docs/api/통계조회.md`

- [ ] **Step 1: fixture 필드 확인 (주기 4종 전부)**

```bash
for f in search_d search_q search_m search_a; do
python3 -c "import json; print('$f', sorted(json.load(open('$FIX/$f.json'))['StatisticSearch']['row'][0].keys()))"
done
```

Expected (4개 동일): `['DATA_VALUE', 'ITEM_CODE1', 'ITEM_CODE2', 'ITEM_CODE3', 'ITEM_CODE4', 'ITEM_NAME1', 'ITEM_NAME2', 'ITEM_NAME3', 'ITEM_NAME4', 'STAT_CODE', 'STAT_NAME', 'TIME', 'UNIT_NAME', 'WGT']`

- [ ] **Step 2: 문서 작성**

````markdown
# 통계조회 (StatisticSearch)

> `GET https://ecos.bok.or.kr/api/StatisticSearch/{API_KEY}/{요청유형}/{언어구분}/{시작건수}/{종료건수}/{통계표코드}/{주기}/{검색시작일자}/{검색종료일자}/{통계항목코드1}/{통계항목코드2}/{통계항목코드3}/{통계항목코드4}`

통계표의 실제 데이터를 조회한다. ECOS OpenAPI 의 핵심 서비스.
통계표코드·주기·항목코드는 [통계표목록](통계표목록.md)과 [통계세부항목목록](통계세부항목목록.md)에서 확인한다.

## 기본 정보

| 메서드 | 출력포맷 | 비고 |
| --- | --- | --- |
| GET | JSON, XML | 요청유형 path 인자로 선택 |

## 요청 인자 (path segment 순서)

| 위치 | 인자 | 필수 | 값 설명 |
| --- | --- | --- | --- |
| 1~6 | 공통 인자 | Y | [README 공통 규약](README.md) 참고 |
| 7 | 통계표코드 | Y | 예: `722Y001` |
| 8 | 주기 | Y | A/S/Q/M/SM/D — [README 주기 표](README.md) 참고 |
| 9 | 검색시작일자 | Y | 주기별 날짜 형식 (D → `20260101`, Q → `2023Q1`) |
| 10 | 검색종료일자 | Y | 〃 |
| 11 | 통계항목코드1 | N | 생략 시 전체 항목. 예: `0101000` |
| 12 | 통계항목코드2 | N | 다차원 통계표의 2번째 그룹 항목 |
| 13 | 통계항목코드3 | N | 〃 3번째 그룹 |
| 14 | 통계항목코드4 | N | 〃 4번째 그룹 |

- 주기와 날짜 형식이 다르면 `ERROR-101`, 기간에 데이터가 없으면 `INFO-200` 이 반환된다.

## 응답 필드 (`StatisticSearch.row[]`)

| 필드 | 명칭 | 설명 |
| --- | --- | --- |
| STAT_CODE | 통계표코드 | |
| STAT_NAME | 통계명 | |
| ITEM_CODE1 | 통계항목코드1 | |
| ITEM_NAME1 | 통계항목명1 | |
| ITEM_CODE2 | 통계항목코드2 | 없으면 null |
| ITEM_NAME2 | 통계항목명2 | 〃 |
| ITEM_CODE3 | 통계항목코드3 | 〃 |
| ITEM_NAME3 | 통계항목명3 | 〃 |
| ITEM_CODE4 | 통계항목코드4 | 〃 |
| ITEM_NAME4 | 통계항목명4 | 〃 |
| UNIT_NAME | 단위 | 예: `연%` |
| WGT | 가중치 | 없으면 null |
| TIME | 시점 | 주기별 날짜 형식 |
| DATA_VALUE | 값 | 문자열로 반환됨 (예: `"2.5"`) |

## 샘플

요청 (한국은행 기준금리, 일별):

```
GET https://ecos.bok.or.kr/api/StatisticSearch/{API_KEY}/json/kr/1/5/722Y001/D/20260101/20260131/0101000
```

응답 (발췌 — fixture 에서 삽입):

```json
(Task 1 의 search_d.json 발췌를 여기에 삽입)
```

요청 (분기 GDP — 주기별 날짜 형식 예):

```
GET https://ecos.bok.or.kr/api/StatisticSearch/{API_KEY}/json/kr/1/5/200Y102/Q/2023Q1/2023Q4/10111
```
````

- [ ] **Step 3: 대조 검증 + 커밋**

Task 3 Step 3 과 같은 대조 스크립트를 `search_d.json`/`StatisticSearch`/`docs/api/통계조회.md` 에 대해 실행. Expected: `문서 누락: 없음`

```bash
file -I docs/api/통계조회.md
git add docs/api/통계조회.md
git commit -m "docs: StatisticSearch(통계조회) API 문서 추가"
```

---

### Task 7: 100대통계지표.md (KeyStatisticList)

**Files:**
- Create: `docs/api/100대통계지표.md`

- [ ] **Step 1: fixture 필드 확인**

```bash
python3 -c "import json; print(sorted(json.load(open('$FIX/keystat.json'))['KeyStatisticList']['row'][0].keys()))"
```

Expected: `['CLASS_NAME', 'CYCLE', 'DATA_VALUE', 'KEYSTAT_NAME', 'UNIT_NAME']`
(실측에서 `row_count` 같은 상위 필드가 추가로 있었음 — 상위 레벨 필드도 문서에 기재)

- [ ] **Step 2: 문서 작성**

````markdown
# 100대통계지표 (KeyStatisticList)

> `GET https://ecos.bok.or.kr/api/KeyStatisticList/{API_KEY}/{요청유형}/{언어구분}/{시작건수}/{종료건수}/`

한국은행이 선정한 100대 통계지표의 최신 값을 제공한다. 서비스별 추가 인자가 없다.

## 기본 정보

| 메서드 | 출력포맷 | 비고 |
| --- | --- | --- |
| GET | JSON, XML | 요청유형 path 인자로 선택 |

## 요청 인자 (path segment 순서)

| 위치 | 인자 | 필수 | 값 설명 |
| --- | --- | --- | --- |
| 1~6 | 공통 인자 | Y | [README 공통 규약](README.md) 참고. 추가 인자 없음 |

## 응답 필드 (`KeyStatisticList.row[]`)

| 필드 | 명칭 | 설명 |
| --- | --- | --- |
| CLASS_NAME | 통계 분류명 | 예: `환율`, `국민소득` |
| KEYSTAT_NAME | 지표명 | 예: `원/달러 환율(종가)` |
| DATA_VALUE | 값 | 문자열로 반환됨 |
| CYCLE | 수록 시점 | 최신값 기준 시점 (예: `20260805`) — 주기 코드가 아니라 날짜임에 주의 |
| UNIT_NAME | 단위 | |

- 상위 레벨에 `list_total_count` 외에 `row_count` 필드가 추가로 반환된다 (실측 확인).

## 샘플

요청:

```
GET https://ecos.bok.or.kr/api/KeyStatisticList/{API_KEY}/json/kr/1/10/
```

응답 (발췌 — fixture 에서 삽입):

```json
(Task 1 의 keystat.json 발췌를 여기에 삽입)
```
````

- [ ] **Step 3: 대조 검증 + 커밋**

Task 3 Step 3 과 같은 대조 스크립트를 `keystat.json`/`KeyStatisticList`/`docs/api/100대통계지표.md` 에 대해 실행. Expected: `문서 누락: 없음`

```bash
file -I docs/api/100대통계지표.md
git add docs/api/100대통계지표.md
git commit -m "docs: KeyStatisticList(100대통계지표) API 문서 추가"
```

---

### Task 8: 통계메타DB.md (StatisticMeta)

**Files:**
- Create: `docs/api/통계메타DB.md`

- [ ] **Step 1: fixture 필드 확인**

```bash
python3 -c "import json; print(sorted(json.load(open('$FIX/meta.json'))['StatisticMeta']['row'][0].keys()))"
```

Expected: `['CONT_CODE', 'CONT_NAME', 'LVL', 'META_DATA', 'P_CONT_CODE']`

- [ ] **Step 2: 문서 작성**

````markdown
# 통계메타DB (StatisticMeta)

> `GET https://ecos.bok.or.kr/api/StatisticMeta/{API_KEY}/{요청유형}/{언어구분}/{시작건수}/{종료건수}/{데이터명}`

통계메타DB 의 메타데이터를 데이터명으로 조회한다. 트리 구조(레벨·부모코드)로 반환된다.

## 기본 정보

| 메서드 | 출력포맷 | 비고 |
| --- | --- | --- |
| GET | JSON, XML | 요청유형 path 인자로 선택 |

## 요청 인자 (path segment 순서)

| 위치 | 인자 | 필수 | 값 설명 |
| --- | --- | --- | --- |
| 1~6 | 공통 인자 | Y | [README 공통 규약](README.md) 참고 |
| 7 | 데이터명 | Y | 예: `경제심리지수`. URL 인코딩 필요 |

## 응답 필드 (`StatisticMeta.row[]`)

| 필드 | 명칭 | 설명 |
| --- | --- | --- |
| LVL | 레벨 | 트리 깊이 |
| P_CONT_CODE | 상위 콘텐츠코드 | |
| CONT_CODE | 콘텐츠코드 | |
| CONT_NAME | 콘텐츠명 | |
| META_DATA | 메타데이터 | 본문. 상위 노드는 null |

## 샘플

요청:

```
GET https://ecos.bok.or.kr/api/StatisticMeta/{API_KEY}/json/kr/1/10/경제심리지수
```

응답 (발췌 — fixture 에서 삽입):

```json
(Task 1 의 meta.json 발췌를 여기에 삽입)
```
````

- [ ] **Step 3: 대조 검증 + 커밋**

Task 3 Step 3 과 같은 대조 스크립트를 `meta.json`/`StatisticMeta`/`docs/api/통계메타DB.md` 에 대해 실행. Expected: `문서 누락: 없음`

```bash
file -I docs/api/통계메타DB.md
git add docs/api/통계메타DB.md
git commit -m "docs: StatisticMeta(통계메타DB) API 문서 추가"
```

---

### Task 9: 누락 점검 체크리스트 + 최종 검증

**Files:**
- Modify: (수정 필요 발견 시) `docs/api/*.md`

- [ ] **Step 1: 스펙의 누락 점검 체크리스트 실행**

```bash
# 1) 문서 7개 존재
ls docs/api/  # 기대: README.md + 서비스 6개 md

# 2) README 링크 유효성
python3 - <<'EOF'
import os, re
doc = open("docs/api/README.md").read()
for m in re.finditer(r"\]\(([^)#h][^)]*)\)", doc):
    p = os.path.join("docs/api", m.group(1))
    print(("OK " if os.path.exists(p) else "MISSING ") + m.group(1))
EOF

# 3) 인코딩 전수 확인
file -I docs/api/*.md   # 전부 charset=utf-8

# 4) 인증키 유출 전수 확인
grep -rn "$K" docs/ && echo "키 유출!!" || echo "키 유출 없음"
```

- [ ] **Step 2: 응답 필드 전수 대조 일괄 재실행**

Task 3~8 의 대조 스크립트를 6개 서비스 전부에 대해 한 번에 실행하는 스크립트:

```bash
python3 - "$FIX" <<'EOF'
import json, re, os, sys
pairs = [
    ("table_list.json", "StatisticTableList", "docs/api/통계표목록.md"),
    ("word.json", "StatisticWord", "docs/api/통계용어사전.md"),
    ("item_list.json", "StatisticItemList", "docs/api/통계세부항목목록.md"),
    ("search_d.json", "StatisticSearch", "docs/api/통계조회.md"),
    ("keystat.json", "KeyStatisticList", "docs/api/100대통계지표.md"),
    ("meta.json", "StatisticMeta", "docs/api/통계메타DB.md"),
]
fail = False
for fx, root, doc_path in pairs:
    keys = set()
    for r in json.load(open(os.path.join(sys.argv[1], fx)))[root]["row"]:
        keys |= set(r.keys())
    doc = open(doc_path).read()
    missing = [k for k in sorted(keys) if not re.search(rf"^\| {re.escape(k)} \|", doc, re.M)]
    print(f"{root}: {'OK' if not missing else 'MISSING ' + str(missing)}")
    fail = fail or bool(missing)
sys.exit(1 if fail else 0)
EOF
```

Expected: 6개 전부 `OK`, exit 0.

- [ ] **Step 3: 에러코드 표와 실측 대조**

`$FIX/err_*.json` 의 CODE/MESSAGE 가 README 에러 코드 표와 일치하는지 확인. 불일치 시 README 수정.

- [ ] **Step 4: 수정분 커밋 + 플랜 체크박스 갱신**

```bash
git add docs/
git commit -m "docs: ECOS API 문서 최종 검증 및 보정"   # 수정이 있을 때만
git log --oneline
```

- [ ] **Step 5: 사용자 검수 요청**

문서 7개 목록과 검증 결과(체크리스트 통과 내역)를 정리해 사용자에게 검수를 요청한다.
**사용자 승인 후에만 2단계(라이브러리 구현) 플랜을 작성한다.**
