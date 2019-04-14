package sampler

import (
	"testing"
)

func TestSampleLoopSingleEvent(t *testing.T) {
	events := make(chan interface{}, 100)
	s := NewSampler(events)
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
	s := NewSampler(events)

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
