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
