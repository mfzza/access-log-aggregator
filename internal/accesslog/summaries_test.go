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
				requestTotal:  1,
				request2xx:    1,
				durationTotal: 0.224254673,
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
				requestTotal:  1,
				request2xx:    0,
				durationTotal: 0.224254673,
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
				requestTotal:  1,
				request2xx:    0,
				durationTotal: 0.224254673,
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
				requestTotal:  1,
				request2xx:    0,
				durationTotal: 0.224254673,
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
			if math.Abs(tt.want.durationTotal-tt.summary.durationTotal) > 1e-9 {
				t.Errorf("expected avgDuration = %f, got %f", tt.want.durationTotal, tt.summary.durationTotal)
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
				requestTotal:  6,
				request2xx:    3,
				durationTotal: 1.345528038,
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
				requestTotal:  6,
				request2xx:    3,
				durationTotal: 2.39213359,
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
			if math.Abs(tt.want.durationTotal-tt.summary.durationTotal) > 1e-9 {
				t.Errorf("expected avgDuration = %f, got %f", tt.want.durationTotal, tt.summary.durationTotal)
			}
		})
	}
}

func TestSummaries_Aggregate(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		summaries Summaries
		rawRecord []byte
		want      Summaries
	}{
		{name: "new host on empty summaries",
			summaries: Summaries{},
			rawRecord: []byte(`{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":299,"duration":0.224254673}`),
			want:      Summaries{"chatgpt.com": {requestTotal: 1, request2xx: 1, durationTotal: 0.224254673}},
		},
		{name: "existing host on existing summaries",
			summaries: Summaries{"chatgpt.com": {requestTotal: 1, request2xx: 1, durationTotal: 0.224254673}},
			rawRecord: []byte(`{"time":"2025-08-14T02:07:12.680651416Z","host":"chatgpt.com","status_code":300,"duration":0.224254673}`),
			want:      Summaries{"chatgpt.com": {requestTotal: 2, request2xx: 1, durationTotal: 0.448509346}},
		},
		{name: "new host on existing summaries",
			summaries: Summaries{"chatgpt.com": {requestTotal: 1, request2xx: 1, durationTotal: 0.224254673}},
			rawRecord: []byte(`{"time":"2025-08-14T02:07:12.680651416Z","host":"substrate.office.com","status_code":300,"duration":0.224254673}`),
			want:      Summaries{"chatgpt.com": {requestTotal: 1, request2xx: 1, durationTotal: 0.224254673}, "substrate.office.com": {requestTotal: 1, request2xx: 0, durationTotal: 0.224254673}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.summaries.Aggregate(tt.rawRecord)
			if len(tt.summaries) != len(tt.want) {
				t.Fatalf("expected map length %d, tt.summaries %d", len(tt.want), len(tt.summaries))
			}
			for k, v := range tt.want {
				gv, ok := tt.summaries[k]
				if !ok {
					t.Errorf("missing key %q", k)
					continue
				}
				if gv != v {
					t.Errorf("for key %q, tt.summaries %+v, want %+v", k, gv, v)
				}
			}
		})
	}
}
