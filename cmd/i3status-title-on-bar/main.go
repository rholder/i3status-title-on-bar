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

// override this via: go build "-ldflags main.Version=x.x.x", defaults to 0.0.0-dev if unset
var Version string = "0.0.0-dev"

const titleChangeSampleMs = 50
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
  i3status | i3status-title-on-bar --color #00EE00
  i3status | i3status-title-on-bar --append-right --fixed-width 64
  i3status-title-on-bar < i3status-output-example.json

Report bugs and find the latest updates at https://github.com/rholder/i3status-title-on-bar.
`

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
		return true, 2
	}

	if config.printHelp {
		fmt.Fprintln(stdout, helpText)
		return true, 1
	}

	if config.printVersion {
		fmt.Fprintln(stdout, Version)
		return true, 1
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

	titleChangeEvents := make(chan interface{}, titleChangeEventBufferSize)

	currentStatusPids := process.FindPidsByProcessName("i3status")
	if len(currentStatusPids) == 0 {
		fmt.Fprintln(stderr, "No i3status PID could be found")
	}

	windowAPI, err := window.NewX11()
	if err != nil {
		fmt.Fprintln(stderr, err)
	}
	titleChangeSampler := sampler.NewSampler(titleChangeEvents, titleChangeSampleMs)

	go windowAPI.BeginTitleChangeDetection(func() {
		titleChangeEvents <- "changed"
	}, func(err error) {
		fmt.Fprintln(stderr, err)
	})

	go titleChangeSampler.Run(func(value interface{}) {
		process.SignalPidsWithUSR1(currentStatusPids)
	})

	exitCode := i3.RunJsonParsingLoop(stdin, stdout, stderr, windowAPI, config.color, config.appendEnd, config.fixedWidth)
	os.Exit(exitCode)
}
