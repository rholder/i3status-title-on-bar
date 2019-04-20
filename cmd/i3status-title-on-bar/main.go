package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/rholder/i3status-title-on-bar/pkg/i3"
	"github.com/rholder/i3status-title-on-bar/pkg/process"
	"github.com/rholder/i3status-title-on-bar/pkg/sampler"
	"github.com/rholder/i3status-title-on-bar/pkg/window"
)

func main() {
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr

	titleChangeEvents := make(chan interface{}, 1000)

	currentStatusPids := process.FindPidsByProcessName("i3status")
	if len(currentStatusPids) == 0 {
		fmt.Fprintln(stderr, "No i3status PID could be found")
	}

	windowAPI := window.NewX11()
	titleChangeSampler := sampler.NewSampler(titleChangeEvents)

	go windowAPI.BeginTitleChangeDetection(stderr, func() {
		titleChangeEvents <- "changed"
	})

	go titleChangeSampler.Run(func(value interface{}) {
		for _, pid := range currentStatusPids {
			syscall.Kill(pid, syscall.SIGUSR1)
		}
	})

	// TODO make color a flag
	color := "#00FF00"
	exitCode := i3.RunJsonParsingLoop(stdin, stdout, stderr, windowAPI, color)
	os.Exit(exitCode)
}
