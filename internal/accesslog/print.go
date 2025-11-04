package accesslog

import (
	"fmt"
	"sort"
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
	// count longest host name, and sorting host name
	hosts := make([]string, 0, len(ss))
	maxHostLen := 0
	for h := range ss {
		hosts = append(hosts, h)
		l := len(h)
		if l > maxHostLen {
			maxHostLen = l
		}
	}
	sort.Strings(hosts)
	maxHostLen += 2

	fmt.Printf("%-*s %15s %15s %18s %18s\n",
		maxHostLen, "Host", "total_requests", "2xx_requests", "non_2xx_requests", "avg_duration_s")
	fmt.Println(strings.Repeat("-", maxHostLen+70))

	for _, h := range hosts {
		fmt.Printf("%-*s %15d %15d %18d %18.3f\n",
			maxHostLen, h,
			ss[h].requestTotal,
			ss[h].request2xx,
			ss[h].requestTotal-ss[h].request2xx,
			ss[h].avgDuration)
	}
}
