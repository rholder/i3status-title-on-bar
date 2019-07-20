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

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/rholder/i3status-title-on-bar/pkg/i3"
	"github.com/rholder/i3status-title-on-bar/pkg/process"
	"github.com/rholder/i3status-title-on-bar/pkg/sampler"
	"github.com/rholder/i3status-title-on-bar/pkg/window"
)

// Version override via: go build "-ldflags main.Version=x.x.x", defaults to 0.0.0-dev if unset
var Version = "0.0.0-dev"

// No title will change with a frequency higher than this value. See https://www.nngroup.com/articles/response-times-3-important-limits/
const titleChangeSampleMs = 100
const titleChangeEventBufferSize = 1000
const defaultColor = "#00FF00"
const helpText = `Usage: i3status-title-on-bar [OPTIONS...]

  Use i3status-title-on-bar to prepend the currently active X11 window title
  to the beginning (left) of the i3status output JSON as a new node. From there,
  i3status should be able to pick it up and display it on the bar.

Options:
  --color [i3_color_code]  Set the text color of the JSON node (Defaults to #00FF00)
  --append-end             Append window title JSON node to the end instead of the beginning
  --fixed-width [integer]  Truncate and pad to a fixed width, useful with append-end
  --help                   Print this help text and exit
  --version                Print the version and exit

Examples:
  i3status | i3status-title-on-bar --color '#00EE00'
  i3status | i3status-title-on-bar --append-end --fixed-width 64
  i3status-title-on-bar < i3status-output-example.json

Report bugs and find the latest updates at https://github.com/rholder/i3status-title-on-bar.
`

// Non-zero error codes signal different bad exit conditions. Zero is ok.
const (
	PrintErrorCode                int = 1
	BadConfigErrorCode            int = 2
	MissingStatusProcessErrorCode int = 8
	BadDisplayErrorCode           int = 9
)

// Config stores a bit of configuration for the CLI.
type Config struct {
	color        string
	appendEnd    bool
	fixedWidth   int
	printHelp    bool
	printVersion bool
}

func newConfig(name string, args []string) (*Config, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	var (
		color        = fs.String("color", defaultColor, "Set the text color of the JSON node")
		appendEnd    = fs.Bool("append-end", false, "Append window title JSON node to the end")
		fixedWidth   = fs.Int("fixed-width", 0, "Trucate and pad to a fixed width")
		printHelp    = fs.Bool("help", false, "Print additional help text and exit")
		printVersion = fs.Bool("version", false, "Print the version and exit")
	)
	// disable default output
	fs.SetOutput(ioutil.Discard)
	err := fs.Parse(args)

	return &Config{*color, *appendEnd, *fixedWidth, *printHelp, *printVersion}, err
}

func shouldExit(stdout io.Writer, config *Config, err error) (bool, int) {
	if err != nil {
		fmt.Fprintln(stdout, err.Error()+"\n")
		fmt.Fprintln(stdout, helpText)
		return true, BadConfigErrorCode
	}

	if config.printHelp {
		fmt.Fprintln(stdout, helpText)
		return true, PrintErrorCode
	}

	if config.printVersion {
		fmt.Fprintln(stdout, Version)
		return true, PrintErrorCode
	}

	return false, 0
}

func main() {
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr

	config, err := newConfig(os.Args[0], os.Args[1:])
	exit, code := shouldExit(stdout, config, err)
	if exit {
		os.Exit(code)
	}

	// Grab every PID of i3status currently running. There should be only one
	// but just in case let's use all of them.
	currentStatusPids := process.FindPidsByProcessName("i3status")
	if len(currentStatusPids) == 0 {
		// no i3status means nothing to update on window title change
		fmt.Fprintln(stderr, "No i3status PID could be found")
		os.Exit(MissingStatusProcessErrorCode)
	}

	// This window.API is for the current X11 display.
	windowAPI, err := window.NewX11()
	if err != nil {
		// any display error on creation is fatal
		fmt.Fprintln(stderr, err)
		os.Exit(BadDisplayErrorCode)
	}

	// Changes are sampled and an update for i3status is only done every
	// titleChangeSampleMs milliseconds instead of every time X11 decides to
	// change a property. This minimizes signal sending to i3status which forces
	// an update to everything it may be polling.
	titleChangeEvents := make(chan interface{}, titleChangeEventBufferSize)
	titleChangeSampler := sampler.NewSampler(titleChangeEvents, titleChangeSampleMs)
	go titleChangeSampler.Run(func(value interface{}) {
		process.SignalPidsWithUSR1(currentStatusPids)
	})

	// Whenever a change to a window title is detected, send it to this channel
	// to be sampled.
	go windowAPI.DetectWindowTitleChanges(func() {
		titleChangeEvents <- "changed"
	}, func(err error) {
		fmt.Fprintln(stderr, err)
	})

	// With everything set up and running, start processing the output from
	// i3status and injecting the window titles.
	exitCode := i3.RunJSONParsingLoop(stdin, stdout, stderr, windowAPI, config.color, config.appendEnd, config.fixedWidth)
	os.Exit(exitCode)
}
