package app

import (
	"accessAggregator/internal/config"
	"bytes"
	"errors"
	"testing"
	"time"
)

type mockSummarizer struct {
	aggregateLog [][]byte
	formatOut    string
	aggregateErr error
}

func (m *mockSummarizer) Aggregate(b []byte) error {
	m.aggregateLog = append(m.aggregateLog, append([]byte(nil), b...))
	return m.aggregateErr
}

func (m *mockSummarizer) Format() string {
	return m.formatOut
}

func waitOrTimeout(t *testing.T, ch <-chan struct{}, timeout time.Duration) {
	t.Helper()
	select {
	case <-ch:
		return
	case <-time.After(timeout):
		t.Fatal("timeout waiting for channel signal")
	}
}


func TestAggr_TickerPrintsSummary(t *testing.T) {
	out := &bytes.Buffer{}
	s := &mockSummarizer{formatOut: "SUMMARY\n"}

	data := make(chan []byte)
	done := make(chan struct{})

	// fast ticker
	flags := config.Flags{Interval: 15 * time.Millisecond}

	go aggr(done, flags, data, s, out)

	// Let ticker fire once
	time.Sleep(20 * time.Millisecond)

	close(data) // trigger exit
	waitOrTimeout(t, done, time.Second)

	got := out.String()
	if !bytes.Contains([]byte(got), []byte("SUMMARY")) {
		t.Fatalf("expected summary in output, got: %s", got)
	}
}

func TestAggr_AggregatesData(t *testing.T) {
	out := &bytes.Buffer{}
	s := &mockSummarizer{formatOut: "SUMMARY\n"}

	data := make(chan []byte, 2)
	done := make(chan struct{})

	flags := config.Flags{Interval: time.Hour} // disable ticker firing

	go aggr(done, flags, data, s, out)

	data <- []byte("A")
	data <- []byte("B")
	close(data)

	waitOrTimeout(t, done, time.Second)

	if len(s.aggregateLog) != 2 {
		t.Fatalf("expected 2 aggregated records, got %d", len(s.aggregateLog))
	}
}

func TestAggr_MalformedIncrementPrinted(t *testing.T) {
	out := &bytes.Buffer{}
	s := &mockSummarizer{
		formatOut:    "SUMMARY\n",
		aggregateErr: errors.New("bad"),
	}

	data := make(chan []byte, 1)
	done := make(chan struct{})

	flags := config.Flags{Interval: 15 * time.Millisecond}

	go aggr(done, flags, data, s, out)

	// Send malformed record
	data <- []byte("BAD")
	// Wait for ticker to print summary
	time.Sleep(20 * time.Millisecond)

	close(data)
	waitOrTimeout(t, done, time.Second)

	got := out.String()
	if !bytes.Contains([]byte(got), []byte("missing field or malformed log:")) {
		t.Fatalf("expected malformed warning in output, got: %s", got)
	}
}

func TestAggr_FinalSummaryAndDoneClose(t *testing.T) {
	out := &bytes.Buffer{}
	s := &mockSummarizer{formatOut: "SUMMARY\n"}

	data := make(chan []byte)
	done := make(chan struct{})

	flags := config.Flags{Interval: time.Hour}

	go aggr(done, flags, data, s, out)

	close(data) // trigger final summary + done close

	waitOrTimeout(t, done, time.Second)

	got := out.String()

	if !bytes.Contains([]byte(got), []byte("Printing final summary:")) {
		t.Fatalf("expected final summary banner, got: %s", got)
	}
	if !bytes.Contains([]byte(got), []byte("SUMMARY")) {
		t.Fatalf("expected summary in final output, got: %s", got)
	}
}

func TestAggr_DrainsAllRecordsBeforeEnd(t *testing.T) {
	out := &bytes.Buffer{}
	s := &mockSummarizer{formatOut: "SUMMARY\n"}

	data := make(chan []byte, 3)
	done := make(chan struct{})

	flags := config.Flags{Interval: time.Hour}

	go aggr(done, flags, data, s, out)

	data <- []byte("1")
	data <- []byte("2")
	data <- []byte("3")

	close(data)
	waitOrTimeout(t, done, time.Second)

	if len(s.aggregateLog) != 3 {
		t.Fatalf("expected all 3 records drained, got %d", len(s.aggregateLog))
	}
}

func TestAggr_NoGoroutineLeak(t *testing.T) {
	out := &bytes.Buffer{}
	s := &mockSummarizer{formatOut: "SUMMARY\n"}

	data := make(chan []byte)
	done := make(chan struct{})

	flags := config.Flags{Interval: 10 * time.Millisecond}

	go aggr(done, flags, data, s, out)

	close(data)

	waitOrTimeout(t, done, time.Second)
	// No hang mean success
}
