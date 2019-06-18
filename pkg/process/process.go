package process

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// FindPidsByProcessName finds the list of process identifiers exactly matching
// the given name.
func FindPidsByProcessName(exactProcessName string) []int {
	// Detect with pgrep -x
	out, err := exec.Command("pgrep", "-x", exactProcessName).Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	pids := []int{}
	for _, line := range lines {
		pid, err := strconv.Atoi(line)
		if err == nil {
			pids = append(pids, pid)
		}
	}
	return pids
}

// SignalPidsWithUSR1 sends a USR1 signal to each process identifier in the
// given list.
func SignalPidsWithUSR1(pids []int) {
	for _, pid := range pids {
		syscall.Kill(pid, syscall.SIGUSR1)
	}
}
