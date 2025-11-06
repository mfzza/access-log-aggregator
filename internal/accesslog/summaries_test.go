package accesslog

import (
	"math"
	"testing"
	"time"
)

func TestSummary_updateSummary(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		summary   *summary
		newRecord *Record
		want      *summary
	}{
		{
			name:    "first record with 2xx request",
			summary: &summary{},
			newRecord: &Record{
				Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
				Host:       "chatgpt.com",
				StatusCode: 299,
				Duration:   0.224254673,
			},
			want: &summary{
				requestTotal: 1,
				request2xx:   1,
				avgDuration:  0.224254673,
			},
		},
		{
			name:    "first record with lower than 2xx request",
			summary: &summary{},
			newRecord: &Record{
				Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
				Host:       "chatgpt.com",
				StatusCode: 199,
				Duration:   0.224254673,
			},
			want: &summary{
				requestTotal: 1,
				request2xx:   0,
				avgDuration:  0.224254673,
			},
		},
		{
			name:    "first record with higher than 2xx request",
			summary: &summary{},
			newRecord: &Record{
				Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
				Host:       "chatgpt.com",
				StatusCode: 300,
				Duration:   0.224254673,
			},
			want: &summary{
				requestTotal: 1,
				request2xx:   0,
				avgDuration:  0.224254673,
			},
		},
		{
			name:    "first record average",
			summary: &summary{},
			newRecord: &Record{
				Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
				Host:       "chatgpt.com",
				StatusCode: 400,
				Duration:   0.224254673,
			},
			want: &summary{
				requestTotal: 1,
				request2xx:   0,
				avgDuration:  0.224254673,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.summary.updateSummary(tt.newRecord)

			if tt.want.requestTotal != tt.summary.requestTotal {
				t.Errorf("expected requestTotal = %d, got %d", tt.want.requestTotal, tt.summary.requestTotal)
			}
			if tt.want.request2xx != tt.summary.request2xx {
				t.Errorf("expected request2xx = %d, got %d", tt.want.request2xx, tt.summary.request2xx)
			}
			if math.Abs(tt.want.avgDuration-tt.summary.avgDuration) > 1e-9 {
				t.Errorf("expected avgDuration = %f, got %f", tt.want.avgDuration, tt.summary.avgDuration)
			}
		})
	}
}

func TestSummary_updateSummary_multipleRecords(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		summary    *summary
		newRecords []Record
		want       *summary
	}{
		{
			name:    "mixed status code with edge cases",
			summary: &summary{},
			newRecords: []Record{
				{
					Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
					Host:       "chatgpt.com",
					StatusCode: 199,
					Duration:   0.224254673,
				}, {
					Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
					Host:       "chatgpt.com",
					StatusCode: 200,
					Duration:   0.224254673,
				}, {
					Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
					Host:       "chatgpt.com",
					StatusCode: 201,
					Duration:   0.224254673,
				}, {
					Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
					Host:       "chatgpt.com",
					StatusCode: 299,
					Duration:   0.224254673,
				}, {
					Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
					Host:       "chatgpt.com",
					StatusCode: 300,
					Duration:   0.224254673,
				}, {
					Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
					Host:       "chatgpt.com",
					StatusCode: 301,
					Duration:   0.224254673,
				},
			},
			want: &summary{
				requestTotal: 6,
				request2xx:   3,
				avgDuration:  0.224254673,
			},
		},
		{
			name:    "calculate average",
			summary: &summary{},
			newRecords: []Record{
				{
					Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
					Host:       "chatgpt.com",
					StatusCode: 199,
					Duration:   0.182374913,
				}, {
					Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
					Host:       "chatgpt.com",
					StatusCode: 200,
					Duration:   0.923194810,
				}, {
					Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
					Host:       "chatgpt.com",
					StatusCode: 201,
					Duration:   0.219318123,
				}, {
					Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
					Host:       "chatgpt.com",
					StatusCode: 299,
					Duration:   0.123987913,
				}, {
					Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
					Host:       "chatgpt.com",
					StatusCode: 300,
					Duration:   0.819283121,
				}, {
					Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
					Host:       "chatgpt.com",
					StatusCode: 301,
					Duration:   0.123974710,
				},
			},
			want: &summary{
				requestTotal: 6,
				request2xx:   3,
				avgDuration:  0.398688932,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, r := range tt.newRecords {
				tt.summary.updateSummary(&r)
			}

			if tt.want.requestTotal != tt.summary.requestTotal {
				t.Errorf("expected requestTotal = %d, got %d", tt.want.requestTotal, tt.summary.requestTotal)
			}
			if tt.want.request2xx != tt.summary.request2xx {
				t.Errorf("expected request2xx = %d, got %d", tt.want.request2xx, tt.summary.request2xx)
			}
			if math.Abs(tt.want.avgDuration-tt.summary.avgDuration) > 1e-9 {
				t.Errorf("expected avgDuration = %f, got %f", tt.want.avgDuration, tt.summary.avgDuration)
			}
		})
	}
}
