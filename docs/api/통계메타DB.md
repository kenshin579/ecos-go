# 통계메타DB (StatisticMeta)

> `GET https://ecos.bok.or.kr/api/StatisticMeta/{API_KEY}/{요청유형}/{언어구분}/{시작건수}/{종료건수}/{데이터명}`

통계메타DB 의 메타데이터를 데이터명으로 조회한다. 트리 구조(레벨·부모코드)로 반환되며,
말단 노드의 `META_DATA` 에 실제 메타데이터 값이 들어 있다.

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
| LVL | 레벨 | 트리 깊이 (실측: `"1"`~`"4"`, 문자열) |
| P_CONT_CODE | 상위 콘텐츠코드 | |
| CONT_CODE | 콘텐츠코드 | |
| CONT_NAME | 콘텐츠명 | 예: `통계명`, `참고 자료` |
| META_DATA | 메타데이터 | 값. 상위(그룹) 노드는 `null` |

## 샘플

요청:

```
GET https://ecos.bok.or.kr/api/StatisticMeta/{API_KEY}/json/kr/1/10/경제심리지수
```

응답 (발췌 — 그룹 노드와 말단 노드 예):

```json
{
  "StatisticMeta": {
    "list_total_count": 53,
    "row": [
      {
        "LVL": "2",
        "P_CONT_CODE": "0000000108",
        "CONT_CODE": "N13",
        "CONT_NAME": "참고 자료",
        "META_DATA": null
      },
      {
        "LVL": "4",
        "P_CONT_CODE": "N031",
        "CONT_CODE": "0000000002",
        "CONT_NAME": "통계명",
        "META_DATA": "경제심리지수"
      }
    ]
  }
}
```
