package process

import (
	"os/exec"
	"strconv"
	"strings"
)

func FindPidsByProcessName(exactProcessName string) ([]int, error) {
	// Detect with pgrep -x
	out, err := exec.Command("pgrep", "-x", exactProcessName).Output()
	if err != nil {
		return nil, err
	}
	pids := strings.Split(strings.TrimSpace(string(out)), "\n")
	numericPids := make([]int, len(pids))
	for i, pid := range pids {
		numericPids[i], err = strconv.Atoi(pid)
		if err != nil {
			return nil, err
		}
	}
	return numericPids, nil
}
