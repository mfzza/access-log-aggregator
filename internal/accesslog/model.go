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
	requestTotal int
	request2xx   int
	avgDuration  float64 // in seconds
}

// type Summaries []Summary
type Summaries map[string]Summary

func NewRecord(jb []byte) (*Record, error) {
	var r Record
	err := json.Unmarshal(jb, &r)

	// NOTE: ignore line should handled on caller
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if r.Time.IsZero() || r.Host == "" || r.StatusCode == 0 || r.Duration == 0{
		return nil, fmt.Errorf("missing or invalid required field")
	}

	return &r, nil
}
