package accesslog

import (
	"strings"
	"testing"
)

func TestSummaries_format(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		ss   Summaries
	}{
		{
			name: "multiple hosts with data",
			ss: Summaries{
				"example.com": {
					requestTotal: 5,
					request2xx:   4,
					avgDuration:  0.123,
				},
				"another.com": {
					requestTotal: 2,
					request2xx:   2,
					avgDuration:  0.456,
				},
			},
		},
		{
			name: "empty summaries should still produce header",
			ss:   Summaries{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ss.format()

			// Check header
			if !strings.Contains(got, "Access Log Summary") {
				t.Errorf("format() missing header: %v", got)
			}

			// Check column name
			expectedCols := []string{"Host", "total_requests", "2xx_requests", "non_2xx_requests", "avg_duration_s"}
			for _, col := range expectedCols {
				if !strings.Contains(got, col) {
					t.Errorf("format() missing column header %q", col)
				}
			}

			// For non-empty summaries, check host name
			if len(tt.ss) > 0 {
				for host := range tt.ss {
					if !strings.Contains(got, host) {
						t.Errorf("format() missing host name %q in output", host)
					}
				}
			}
		})
	}
}

