package main

import (
	"context"
	"io"
	"testing"
	"time"

)

// mockTailer implements tailer.Tailer for testing
type mockTailer struct {
	recs [][]byte
	i    int
}

func (m *mockTailer) GetRawRecord() ([]byte, error) {
	if m.i >= len(m.recs) {
		return nil, io.EOF
	}
	r := m.recs[m.i]
	m.i++
	return r, nil
}
func (m *mockTailer) Close() error { return nil }

func TestRunStreamLoop(t *testing.T) {
	mock := &mockTailer{
		recs: [][]byte{
			[]byte(`{"msg":"first"}`),
			[]byte(`{"msg":"second"}`),
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	ch := make(chan []byte, 10)
	err := runStreamLoop(mock, ctx, ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	close(ch)

	var got int
	for range ch {
		got++
	}
	if got != 2 {
		t.Fatalf("expected 2 records, got %d", got)
	}
}

