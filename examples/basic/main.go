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
