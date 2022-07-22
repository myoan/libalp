package main

import (
	"log"

	"github.com/myoan/libalp"
)

func main() {
	p, err := libalp.NewAlpProfiler("/tmp/alp.log")
	if err != nil {
		log.Fatal(err)
	}

	err = p.Run(libalp.AlpOption{
		File:        "/var/log/nginx/access.log",
		Reverse:     true,
		SortType:    "p99",
		Percentiles: []int{50, 90, 99},
	})

	if err != nil {
		log.Fatal(err)
	}
}
