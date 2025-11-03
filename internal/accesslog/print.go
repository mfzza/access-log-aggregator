package accesslog

import "fmt"

func (r *Record) Print() {
	fmt.Println("===================================================")
	fmt.Println("time:", r.Time)
	fmt.Println("host:", r.Host)
	fmt.Println("status code:", r.StatusCode)
	fmt.Println("duration:", r.Duration)
}

func (s *Summary) Print() {
	// TODO: dont just use \t
	fmt.Println(s.host,
		"\t", s.requestTotal,
		"\t\t", s.request2xx,
		"\t\t", s.requestTotal-s.request2xx,
		"\t\t", s.avgDuration)
}

func (ss *Summaries) Print() {
	fmt.Println("Host		total_requests  2xx_requests  non_2xx_requests  avg_duration_s")
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	for _, s := range *ss {
		s.Print()
	}
}
