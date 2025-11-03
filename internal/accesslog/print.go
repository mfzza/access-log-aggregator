package accesslog

import "fmt"

func (r *Record) Print() {
	fmt.Println("===================================================")
	fmt.Println("time:", r.Time)
	fmt.Println("host:", r.Host)
	fmt.Println("status code:", r.StatusCode)
	fmt.Println("duration:", r.Duration)
}

func (ss *Summaries) Print() {
	fmt.Println("Host		total_requests  2xx_requests  non_2xx_requests  avg_duration_s")
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	for k, s := range *ss {
	// TODO: dont just use \t
		fmt.Println(k,
			"\t", s.requestTotal,
			"\t\t", s.request2xx,
			"\t\t", s.requestTotal-s.request2xx,
			"\t\t", s.avgDuration)
	}
}
