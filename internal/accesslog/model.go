package accesslog

import (
	"encoding/json"
	"fmt"
	"time"
)

type Record struct {
	Time       time.Time `json:"time"`
	Host       string    `json:"host"`
	StatusCode int       `json:"status_code"`
	Duration   float64   `json:"duration"`
}

type Summary struct {
	host         string
	requestTotal int
	request2xx   int
	avgDuration  float64 // in seconds
}

type Summaries []Summary

func NewSummary(host string) (*Summary, error) {
	return &Summary{host: host}, nil
}

func NewRecord(jb []byte) (*Record, error) {
	var al Record
	err := json.Unmarshal(jb, &al)
	if err != nil {
		return nil, fmt.Errorf("error jsonnya bang: %w", err)
	}
	return &al, nil
}

func (r *Record) Print() {
	fmt.Println("===================================================")
	fmt.Println("time:", r.Time)
	fmt.Println("host:", r.Host)
	fmt.Println("status code:", r.StatusCode)
	fmt.Println("duration:", r.Duration)
}

func (s *Summary) Print() {
	fmt.Println(s.host, s.requestTotal, s.request2xx, s.requestTotal - s.request2xx, s.avgDuration)
}

func (ss *Summaries) Print() {
	fmt.Println("Host		total_requests  2xx_requests  non_2xx_requests  avg_duration_s")
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	for _, s := range *ss {
		s.Print()
	}
}

