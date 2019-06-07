package sampler

import (
	"time"
)

type Sampler struct {
	events chan interface{}
	stop   chan interface{}
	timeMs time.Duration
}

func NewSampler(events chan interface{}, timeMs int) Sampler {
	return Sampler{
		events: events,
		stop:   make(chan interface{}, 1),
		timeMs: time.Millisecond * time.Duration(timeMs),
	}
}

func (sampler Sampler) Close() {
	// send stop signal, then one final event in case it's blocking
	sampler.stop <- "stop"
	sampler.events <- "end"
}

func (sampler Sampler) Run(onSignal func(interface{})) {
	// main loop
	for poll(sampler.stop) == nil {
		// block here on next event
		value := <-sampler.events

		// non-blocking function to drain the channel
		for poll(sampler.events) != nil {
			// drain these events that may have piled up
		}
		// at this point, new events may be sent to channel
		onSignal(value)

		// while the signal function and this sleep run, new events may occur
		time.Sleep(sampler.timeMs)
	}
}

func poll(events <-chan interface{}) *interface{} {
	select {
	case msg := <-events:
		// next message
		return &msg
	default:
		// nil when there is no next message
		return nil
	}
}
