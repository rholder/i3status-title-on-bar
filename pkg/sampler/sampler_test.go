package sampler

import (
	"testing"
	"time"
)

func TestSampleLoopSingleEvent(t *testing.T) {
	events := make(chan interface{}, 100)
	s := NewSampler(events, 50)
	events <- "changed"
	count := 0
	s.Run(func(value interface{}) {
		s.Close()
		count++
	})

	if count != 1 {
		t.Fatal("Expected only 1 stop event")
	}
}

func TestSampleLoopMultipleEvents(t *testing.T) {
	events := make(chan interface{}, 100)
	s := NewSampler(events, 50)

	events <- "changed"
	events <- "changed"
	events <- "changed"
	events <- "changed"

	count := 0
	s.Run(func(value interface{}) {
		s.Close()
		count++
	})

	if count != 1 {
		t.Fatal("Expected only 1 stop event")
	}
}

func TestSampleLoopClosing(t *testing.T) {
	events := make(chan interface{}, 100)
	s := NewSampler(events, 50)

	events <- "changed"
	events <- "changed"
	events <- "changed"
	events <- "changed"

	count := 0
	go s.Run(func(value interface{}) {
		count++
	})
	time.Sleep(200 * time.Millisecond)
	s.Close()

	if count != 1 {
		t.Fatalf("Expected only 1 stop event, instead saw %v", count)
	}
}
