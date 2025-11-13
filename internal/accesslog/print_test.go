package accesslog

import (
	"reflect"
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
					durationTotal:  0.123,
				},
				"another.com": {
					requestTotal: 2,
					request2xx:   2,
					durationTotal:  0.456,
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
			got := tt.ss.Format()

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

func TestSummaries_sort(t *testing.T) {
	tests := []struct {
		name    string
		ss      Summaries
		want    []string
		wantMax int
	}{
		{
			name: "empty summaries returns empty slice and base len 2",
			ss:   Summaries{},
			want: []string{},
			// maxHostLen starts 0 + 2 padding
			wantMax: 2,
		},
		{
			name: "single host",
			ss: Summaries{
				"example.com": {},
			},
			want:    []string{"example.com"},
			wantMax: len("example.com") + 2,
		},
		{
			name: "multiple hosts sorted alphabetically",
			ss: Summaries{
				"zulu.com":   {},
				"alpha.com":  {},
				"middle.org": {},
			},
			want:    []string{"alpha.com", "middle.org", "zulu.com"},
			wantMax: len("middle.org") + 2, // longest key + 2
		},
		{
			name: "hosts with equal length",
			ss: Summaries{
				"aaa": {},
				"ccc": {},
				"bbb": {},
			},
			want:    []string{"aaa", "bbb", "ccc"},
			wantMax: len("aaa") + 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotMax := tt.ss.sort()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sort() got hosts = %v, want %v", got, tt.want)
			}
			if gotMax != tt.wantMax {
				t.Errorf("sort() got maxHostLen = %v, want %v", gotMax, tt.wantMax)
			}
		})
	}
}
