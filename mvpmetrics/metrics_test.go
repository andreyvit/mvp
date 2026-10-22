package mvpmetrics

import (
	"fmt"
	"io"
	"sync"
	"testing"
)

func TestEnumCallbackMayUpdateMetric(t *testing.T) {
	counter := NewCounter("jobs_total", []string{"state"}, &Registry{})
	counter.Add(1, "ok")

	counter.Enum(func(labelValues []string, value uint64) {
		if value != 1 || labelValues[0] != "ok" {
			t.Errorf("enum yielded %v=%d", labelValues, value)
		}
		counter.Inc(labelValues[0])
	})

	got := map[string]uint64{}
	counter.Enum(func(labelValues []string, value uint64) {
		got[labelValues[0]] = value
	})
	if len(got) != 1 || got["ok"] != 2 {
		t.Fatalf("after nested update: %v", got)
	}
}

func TestEnumConcurrentWithUpdates(t *testing.T) {
	counter := NewCounter("events_total", []string{"kind"}, &Registry{})
	gauge := NewGauge("in_flight", []string{"kind"}, &Registry{})

	const workers = 4
	const updates = 1000

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			label := fmt.Sprint(i)
			for j := 0; j < updates; j++ {
				counter.Inc(label)
				gauge.Add(1, label)
				gauge.Add(-1, label)
			}
		}(i)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()

	reader := &Writer{}
	reader.Reset(io.Discard)

	for running := true; running; {
		select {
		case <-done:
			running = false
		default:
			counter.WriteMetricTo(reader)
			gauge.WriteMetricTo(reader)
		}
	}

	var sum uint64
	counter.Enum(func(_ []string, value uint64) {
		sum += value
	})
	if want := uint64(workers * updates); sum != want {
		t.Fatalf("counter sum = %d, want %d", sum, want)
	}
}
