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

package process

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func poll(events chan os.Signal) *os.Signal {
	select {
	case msg := <-events:
		// next message
		return &msg
	default:
		// nil when there is no next message
		return nil
	}
}

func TestFindPidsByProcessName(t *testing.T) {
	cmd := exec.Command("sleep", "2")
	cmd.Start()
	currentPid := cmd.Process.Pid
	pids := FindPidsByProcessName("sleep")

	found := false
	for _, pid := range pids {
		if pid == currentPid {
			found = true
		}
	}
	cmd.Wait()
	if !found {
		t.Fatal("Expected to find matching PID in search")
	}
}

func TestFindPidsByProcessNameNoProcessExists(t *testing.T) {
	pids := FindPidsByProcessName("FAKE_PROCESS_NAME")
	if len(pids) > 0 {
		t.Fatal("Expected no process ids")
	}
}

func TestSignalPidsWithUSR1(t *testing.T) {
	// attempt to catch a signal from this current process
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGUSR1)
	currentPid := os.Getpid()

	SignalPidsWithUSR1([]int{currentPid})

	// poll for this signal a few times, sometimes it takes a bit
	expectedSignal := poll(signalChannel)
	for i := 0; i < 3; i++ {
		if expectedSignal == nil {
			time.Sleep(2 * time.Second)
			expectedSignal = poll(signalChannel)
		}
	}
	if expectedSignal == nil {
		t.Fatal("Expected to receive signal USR1")
	}
}
