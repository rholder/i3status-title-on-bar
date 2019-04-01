package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"github.com/rholder/i3status-title-on-bar/pkg/window"
	"github.com/rholder/i3status-title-on-bar/pkg/x11"
)

func patchWifiBug(allJson []interface{}) {
	for _, rawEntry := range allJson {
		switch entry := rawEntry.(type) {
		case map[string]interface{}:
			if entry["name"] == "wireless" {
				full_text := entry["full_text"].(string)
				percent := full_text[len(full_text)-4:]
				if strings.HasPrefix(percent, "0") {
					// W: SOME_WIFI_SSID 067%"
					fixed_text := full_text[:len(full_text)-4] + full_text[len(full_text)-3:]
					entry["full_text"] = fixed_text
				}
			}
		}
	}
}

func i3statusPids() []int {
	// Detect with pgrep -x i3status
	out, err := exec.Command("pgrep", "-x", "i3status").Output()
	if err != nil {
		log.Fatal(err)
	}
	pids := strings.Split(strings.TrimSpace(string(out)), "\n")
	numericPids := make([]int, len(pids))
	for i, pid := range pids {
		numericPids[i], err = strconv.Atoi(pid)
		if err != nil {
			log.Fatal(err)
		}
	}
	return numericPids
}

func signalStatusUpdate(pids []int) {
	for _, pid := range pids {
		syscall.Kill(pid, syscall.SIGUSR1)
	}
}

func poll(events <-chan string) *string {
	select {
	case msg := <-events:
		// next message
		return &msg
	default:
		// nil when there is no next message
		return nil
	}
}

func runSignalLoop(titleChangeEvents <-chan string) {
	currentStatusPids := i3statusPids()

	// main loop
	for {
		// block here on next event
		<-titleChangeEvents

		// non-blocking function to drain the channel
		for poll(titleChangeEvents) != nil {
			// drain these events
		}
		// at this point, new events may be sent to the channel while this runs
		signalStatusUpdate(currentStatusPids)
		time.Sleep(50 * time.Millisecond)
	}
}

func scannerError(out io.Writer, scanner *bufio.Scanner, errorCode int) int {
	if scanner.Err() != nil {
		fmt.Fprintf(out, "ERROR from bufio.Scanner: %s\n", scanner.Err())
	}
	return errorCode
}

func runJsonParsingLoop(stdin io.Reader, stdout io.Writer, stderr io.Writer, windowAPI window.WindowAPI) int {

	// read from input
	scanner := bufio.NewScanner(stdin)

	// Skip the first line which contains the version header.
	// {"version":1}
	if !scanner.Scan() {
		// TODO happens way too often, be more resilient to bad scanner starts from stdin
		return scannerError(stderr, scanner, 3)
	}
	line := strings.TrimSpace(scanner.Text())
	fmt.Fprintf(stdout, "%s\n", line)

	// The second line contains the start of the infinite array.
	// [
	if !scanner.Scan() {
		return scannerError(stderr, scanner, 4)
	}
	line = strings.TrimSpace(scanner.Text())
	fmt.Fprintf(stdout, "%s\n", line)

	// Start the main loop.
	var parsed []interface{}
	prefix := ""
	for scanner.Scan() {
		// read from stdin
		line = strings.TrimSpace(scanner.Text())
		prefix = ""

		// clip off the comma if it exists, save it to add back later
		if strings.HasPrefix(line, ",") {
			line, prefix = line[1:], ","
		}

		// parse the original JSON
		err := json.Unmarshal([]byte(line), &parsed)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 5
		}

		// TODO make color a flag
		// inject window title node first
		title := windowAPI.ActiveWindowTitle()
		titleNode := map[string]string{
			"name":      "window_title",
			"full_text": title,
			"color":     "#00FF00"}

		// bolt together the JSON
		var allJson []interface{}
		allJson = append(allJson, titleNode)
		allJson = append(allJson, parsed...)

		// TODO make this a flag
		patchWifiBug(allJson)

		// parsed = append(titleNode, parsed...) // TODO figure out how to do this cleanly
		parsedJson, err := json.Marshal(allJson)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 6
		}

		// output i3bar JSON
		fmt.Fprintf(stdout, "%s%s\n", prefix, parsedJson)
	}

	if scanner.Err() != nil {
		return scannerError(stderr, scanner, 7)
	} else {
		// we hit EOL normally, everything is fine
		return 0
	}
}

func main() {
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr

	windowAPI := x11.New()
	titleChangeEvents := make(chan string, 100)

	go windowAPI.BeginTitleChangeDetection(stderr, func() {
		titleChangeEvents <- "changed"
	})
	go runSignalLoop(titleChangeEvents)

	exitCode := runJsonParsingLoop(stdin, stdout, stderr, windowAPI)
	os.Exit(exitCode)
}
