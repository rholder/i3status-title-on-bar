package process

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

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

func SignalPidsWithUSR1(pids []int) {
	for _, pid := range pids {
		syscall.Kill(pid, syscall.SIGUSR1)
	}
}
