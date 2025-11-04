package accesslog

import (
	"fmt"
	"strings"
)

func (r *Record) Print() {
	fmt.Println("===================================================")
	fmt.Println("time:", r.Time)
	fmt.Println("host:", r.Host)
	fmt.Println("status code:", r.StatusCode)
	fmt.Println("duration:", r.Duration)
}

func (ss Summaries) Print() {
	maxHostLen := 0
	for h := range ss {
		l := len(h)
		if l > maxHostLen {
			maxHostLen = l
		}
	}
	maxHostLen += 2

	fmt.Printf("%-*s %15s %15s %18s %18s\n",
		maxHostLen, "Host", "total_requests", "2xx_requests", "non_2xx_requests", "avg_duration_s")
	fmt.Println(strings.Repeat("-", maxHostLen+70))

	//TODO: maybe try to sort it
	for h, s := range ss {
		fmt.Printf("%-*s %15d %15d %18d %18.3f\n",
			maxHostLen, h, s.requestTotal, s.request2xx,
			s.requestTotal-s.request2xx, s.avgDuration)
	}
}
