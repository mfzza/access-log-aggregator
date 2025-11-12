package accesslog

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func (ss Summaries) sort() ([]string, int) {
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

	return hosts, maxHostLen
}

func (ss Summaries) Format() string {
	hosts, maxHostLen := ss.sort()

	var b strings.Builder

	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "*** Access Log Summary as of", time.Now().Format("2006-01-02 15:04:05"), "***")
	fmt.Fprintln(&b, strings.Repeat("=", maxHostLen+72))
	fmt.Fprintf(&b, "%-*s %15s %15s %18s %18s\n",
		maxHostLen, "Host", "total_requests", "2xx_requests", "non_2xx_requests", "avg_duration_s")
	fmt.Fprintln(&b, strings.Repeat("-", maxHostLen+72))

	for _, h := range hosts {
		fmt.Fprintf(&b, "%-*s %15d %15d %18d %18.3f\n",
			maxHostLen, h,
			ss[h].requestTotal,
			ss[h].request2xx,
			ss[h].requestTotal-ss[h].request2xx,
			ss[h].avgDuration)
	}
	fmt.Fprintln(&b, strings.Repeat("=", maxHostLen+72))

	return b.String()
}
