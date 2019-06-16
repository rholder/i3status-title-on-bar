package sampler

import (
	"time"
)

// Sampler can be used to return a single message on a channel when multiple
// events may have occurred during a configurable time frame. For instance, if
// 1000 messages land in the channel that is used by the running Sampler, only
// one of those messages ever makes it to the onSignal function of Run(). This
// effectively filters out all messages that occur in a given time period and
// allows for a process to only respond when at least one message has appeared
// on the given channel in the period between channel pollings. Additionally,
// because of the blocking nature of channels, we don't need to poll
// continuously, but instead block waiting for at least one sample to appear
// instead of creating a busy wait.
type Sampler struct {
	events chan interface{}
	stop   chan interface{}
	timeMs time.Duration
}

// NewSampler creates a new Sampler for the given channel, sleeping the given
// number of milliseconds after firing off the onSignal function.
func NewSampler(events chan interface{}, timeMs int) Sampler {
	return Sampler{
		events: events,
		stop:   make(chan interface{}, 1),
		timeMs: time.Millisecond * time.Duration(timeMs),
	}
}

// Close the current Sampler and shut down its Run loop.
func (sampler Sampler) Close() {
	// send stop signal, then one final event in case it's blocking
	sampler.stop <- nil
	sampler.events <- nil
}

// Run the Sampler. Messages appearing on the channel the Sampler is sampling
// will begin to flow through the running Sampler with the following behavior:
// at least one message within the given sample frequency will be used to call
// the onSignal function. It is assumed that the Sampler is the only consumer of
// its configured channel and additionally that it will consume all messages
// from the channel when they become available.
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

// Poll for events on the given channel instead of blocking. Returns the event
// if it exists or nil when it does not.
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
