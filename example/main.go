package main

import (
	"log"

	"github.com/myoan/libalp"
)

func main() {
	p, err := libalp.NewAlpProfiler("./alp.log")
	if err != nil {
		log.Fatal(err)
	}

	err = p.Run(libalp.AlpOption{
		File:           "/Users/yoan/dev/github.com/myoan/myalp/access.log",
		Reverse:        true,
		Limit:          10000,
		MatchingGroups: "/api/estate/\\d+,/api/estate/req_doc/\\d+,/api/chair/.+,/api/recommended_estate/.+",
		SortType:       "p99",
		Percentiles:    []int{50, 90, 99},
	})

	if err != nil {
		log.Fatal(err)
	}
}
