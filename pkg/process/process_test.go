package process

import (
	"os/exec"
	"testing"
)

func TestFindPidsByProcessName(t *testing.T) {
	cmd := exec.Command("sleep", "2")
	cmd.Start()
	currentPid := cmd.Process.Pid
	pids, err := FindPidsByProcessName("sleep")
	if err != nil {
		t.Fatal(err)
	}

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
	pids, err := FindPidsByProcessName("FAKE_PROCESS_NAME")
	if err == nil {
		t.Fatal("Expected an error")
	}
	if len(pids) > 0 {
		t.Fatal("Expected no process ids")
	}
}
