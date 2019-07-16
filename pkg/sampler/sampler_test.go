// Copyright 2019 Ray Holder
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
