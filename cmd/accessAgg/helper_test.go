package main

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

// mockTailer implements tailer.Tailer
type mockTailer struct {
	records [][]byte
	errs    []error
	calls   int
}

func (m *mockTailer) GetRawRecord() ([]byte, error) {
	if m.calls < len(m.records) {
		rec := m.records[m.calls]
		m.calls++
		return rec, nil
	}
	if m.calls-len(m.records) < len(m.errs) {
		err := m.errs[m.calls-len(m.records)]
		m.calls++
		return nil, err
	}
	m.calls++
	return nil, io.EOF
}

func (m *mockTailer) Close() error {
	return nil
}

func TestRunStreamLoop(t *testing.T) {
	tests := []struct {
		name        string
		mock        *mockTailer
		ctxSetup    func() (context.Context, context.CancelFunc)
		rawCap      int
		expectErr   string
		expectRecs  [][]byte
		cancelAfter time.Duration // optional delayed cancel
	}{
		{
			name: "normal flow (sends records and exits on cancel)",
			mock: &mockTailer{
				records: [][]byte{[]byte("one"), []byte("two")},
				errs:    []error{io.EOF},
			},
			ctxSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				return ctx, cancel
			},
			rawCap:      2,
			cancelAfter: 50 * time.Millisecond,
			expectRecs:  [][]byte{[]byte("one"), []byte("two")},
		},
		{
			name: "context canceled before start",
			mock: &mockTailer{},
			ctxSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			rawCap:     1,
			expectRecs: nil,
		},
		{
			name: "handles EOF and resumes",
			mock: &mockTailer{
				errs:    []error{io.EOF},
				records: [][]byte{[]byte("later")},
			},
			ctxSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				return ctx, cancel
			},
			rawCap:      1,
			cancelAfter: 100 * time.Millisecond,
			expectRecs:  [][]byte{[]byte("later")},
		},
		{
			name: "returns wrapped error on failure",
			mock: &mockTailer{
				errs: []error{errors.New("boom")},
			},
			ctxSetup: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			rawCap:    1,
			expectErr: "reading record: boom",
		},
		{
			name: "context canceled while sending",
			mock: &mockTailer{
				records: [][]byte{[]byte("last")},
			},
			ctxSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				return ctx, cancel
			},
			rawCap: 0, // unbuffered channel → send will block
			// we cancel after short delay so it exits
			cancelAfter: 50 * time.Millisecond,
			expectRecs:  [][]byte{[]byte("last")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.ctxSetup()
			defer cancel()

			rawRecords := make(chan []byte, tt.rawCap)

			done := make(chan error, 1)
			go func() {
				done <- runStreamLoop(tt.mock, ctx, rawRecords)
			}()

			if tt.cancelAfter > 0 {
				time.AfterFunc(tt.cancelAfter, cancel)
			}

			var got [][]byte
			timeout := time.After(300 * time.Millisecond)

		loop:
			for {
				select {
				case r := <-rawRecords:
					got = append(got, r)
				case err := <-done:
					// check result
					if tt.expectErr == "" && err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if tt.expectErr != "" && (err == nil || !strings.Contains(err.Error(), tt.expectErr)) {
						t.Fatalf("expected error %q, got %v", tt.expectErr, err)
					}
					break loop
				case <-timeout:
					t.Fatalf("test timeout, got records so far: %q", got)
				}
			}

			if !reflect.DeepEqual(got, tt.expectRecs) {
				t.Errorf("got records %q, want %q", got, tt.expectRecs)
			}
		})
	}
}
