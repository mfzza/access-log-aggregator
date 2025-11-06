package accesslog_test

import (
	"accessAggregator/internal/accesslog"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestReader_GetRawRecord(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		r       io.Reader
		want    []byte
		wantErr bool
	}{
		{
			name:    "single line ending with newline",
			r:       strings.NewReader("hello world\n"),
			want:    []byte("hello world\n"),
			wantErr: false,
		},
		{
			name:    "multiple lines - reads first line only",
			r:       strings.NewReader("first line\nsecond line\n"),
			want:    []byte("first line\n"),
			wantErr: false,
		},
		{
			name:    "no newline - return EOF",
			r:       strings.NewReader("last line no newline"),
			want:    nil,
			wantErr: true, // expect it io.EOF
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lr := accesslog.NewReader(tt.r)
			got, err := lr.GetRawRecord()

			if (err != nil) != tt.wantErr {
				t.Fatalf("GetRawRecord() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !bytes.Equal(got, tt.want) {
				t.Errorf("GetRawRecord() = %q, want %q", got, tt.want)
			}
		})
	}
}

