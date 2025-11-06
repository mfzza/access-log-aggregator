package accesslog_test

import (
	"accessAggregator/internal/accesslog"
	"testing"
	"time"
)

func TestNewRecord(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		rawRecord []byte
		want      *accesslog.Record
		wantErr   bool
	}{
		{
			name:      "valid JSON with all required fields",
			rawRecord: []byte(`{"time":"2025-08-14T02:07:12.680651416Z","level":"INFO","msg":"access","scheme":"https","method":"POST","request_uri":"/ces/v1/t","status_code":200,"size":16,"action":"passthrough","host":"chatgpt.com","client_ip":"192.168.2.42","server_ip":"104.18.32.47","duration":0.224254673,"details":""}`),
			want: &accesslog.Record{
				Time:       time.Date(2025, 8, 14, 2, 7, 12, 680651416, time.UTC),
				Host:       "chatgpt.com",
				StatusCode: 200,
				Duration:   0.224254673,
			},
			wantErr: false,
		},
		{
			name:      "invalid JSON format",
			rawRecord: []byte(`{"time":"2025-08-14T02:07:12.680651416Z","level":"INFO","msg":"access","scheme":"https","method":"POST","request_uri":"/ces/v1/t","status_code":200,"size":16,"action":"passthrough","host":"chatgpt.com","client_ip":"192.168.2.42","server_ip":"104.18.32.47","duration":0.224254673,"details":""},`),
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "missing Time field",
			rawRecord: []byte(`{"level":"INFO","msg":"access","scheme":"https","method":"POST","request_uri":"/ces/v1/t","status_code":200,"size":16,"action":"passthrough","host":"chatgpt.com","client_ip":"192.168.2.42","server_ip":"104.18.32.47","duration":0.224254673,"details":""}`),
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "missing Host field",
			rawRecord: []byte(`{"time":"2025-08-14T02:07:12.680651416Z","level":"INFO","msg":"access","scheme":"https","method":"POST","request_uri":"/ces/v1/t","status_code":200,"size":16,"action":"passthrough","client_ip":"192.168.2.42","server_ip":"104.18.32.47","duration":0.224254673,"details":""}`),
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "missing StatusCode field",
			rawRecord: []byte(`{"time":"2025-08-14T02:07:12.680651416Z","level":"INFO","msg":"access","scheme":"https","method":"POST","request_uri":"/ces/v1/t","size":16,"action":"passthrough","host":"chatgpt.com","client_ip":"192.168.2.42","server_ip":"104.18.32.47","duration":0.224254673,"details":""}`),
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "missing Duration field",
			rawRecord: []byte(`{"time":"2025-08-14T02:07:12.680651416Z","level":"INFO","msg":"access","scheme":"https","method":"POST","request_uri":"/ces/v1/t","status_code":200,"size":16,"action":"passthrough","host":"chatgpt.com","client_ip":"192.168.2.42","server_ip":"104.18.32.47","details":""}`),
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "zero StatusCode",
			rawRecord: []byte(`{"time":"2025-08-14T02:07:12.680651416Z","level":"INFO","msg":"access","scheme":"https","method":"POST","request_uri":"/ces/v1/t","status_code":0,"size":16,"action":"passthrough","host":"chatgpt.com","client_ip":"192.168.2.42","server_ip":"104.18.32.47","duration":0.224254673,"details":""}`),
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "zero Duration",
			rawRecord: []byte(`{"time":"2025-08-14T02:07:12.680651416Z","level":"INFO","msg":"access","scheme":"https","method":"POST","request_uri":"/ces/v1/t","status_code":200,"size":16,"action":"passthrough","host":"chatgpt.com","client_ip":"192.168.2.42","server_ip":"104.18.32.47","duration":0,"details":""}`),
			want:      nil,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := accesslog.NewRecord(tt.rawRecord)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewRecord() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewRecord() succeeded unexpectedly")
			}
			if !got.Time.Equal(tt.want.Time) || got.Host != tt.want.Host || got.StatusCode != tt.want.StatusCode || got.Duration != tt.want.Duration {
				t.Errorf("NewRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}
