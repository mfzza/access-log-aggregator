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

// TODO: handle missing field
func NewRecord(jb []byte) (*Record, error) {
	var al Record
	err := json.Unmarshal(jb, &al)
	if err != nil {
		// TODO: just skip if malformed, dont return error
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return &al, nil
}

