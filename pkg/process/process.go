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
