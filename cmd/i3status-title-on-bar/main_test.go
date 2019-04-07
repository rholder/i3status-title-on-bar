package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type TestWindowAPI struct {}

func (testWindowAPI TestWindowAPI) ActiveWindowTitle() string {
	return "foo"
}

func (testWindowAPI TestWindowAPI) BeginTitleChangeDetection(stderr io.Writer, onChange func()) error {
	return nil
}

func TestJsonParsingLoopEmptyInput(t *testing.T) {
	lines := strings.NewReader("")
	errorCode := runJsonParsingLoop(lines, nil, nil, nil)
	if errorCode != 3 {
		t.Fatal("Expected error from parsing loop")
	}
}

func TestJsonParsingLoopNewlineInput(t *testing.T) {
	lines := strings.NewReader("\n")
	stdout := os.Stdout
	stderr := os.Stderr
	errorCode := runJsonParsingLoop(lines, stdout, stderr, nil)
	if errorCode != 4 {
		t.Fatal("Expected error from parsing loop")
	}
}

func TestJsonParsingLoopHappyBlankInput(t *testing.T) {
	lines := strings.NewReader("\n\n")
	stdout := os.Stdout
	stderr := os.Stderr
	errorCode := runJsonParsingLoop(lines, stdout, stderr, nil)
	if errorCode != 0 {
		t.Fatal("Expected no error from parsing loop")
	}
}

func TestJsonParsingLoopBadJSONInput(t *testing.T) {
	lines := strings.NewReader("\n\nPOTATO")
	stdout := os.Stdout
	stderr := os.Stderr
	errorCode := runJsonParsingLoop(lines, stdout, stderr, nil)
	if errorCode != 5 {
		t.Fatal("Expected error from parsing loop")
	}
}

func TestJsonParsingLoopGoodJSONInput(t *testing.T) {
	input := `[{"name":"wireless","instance":"wlp1s0","color":"#00FF00","markup":"none","full_text":"W: SOME_WIFI_SSID 067%"}]`
	emptyReader := strings.NewReader("\n\n" + input)
	stdout := os.Stdout
	stderr := os.Stderr
	windowAPI := TestWindowAPI{}
	errorCode := runJsonParsingLoop(emptyReader, stdout, stderr, windowAPI)
	if errorCode != 0 {
		t.Fatal("Expected no error from parsing loop")
	}
}

func TestPatchWifiBug(t *testing.T) {
	line := `[{"name":"disk_info","instance":"/","markup":"none","full_text":"18.4 GiB"},{"name":"wireless","instance":"wlp1s0","color":"#00FF00","markup":"none","full_text":"W: SOME_WIFI_SSID 067%"}]`
	var jsonPatched, jsonUnpatched []interface{}

	// patched version
	json.Unmarshal([]byte(line), &jsonPatched)
	patchWifiBug(jsonPatched)
	var patched, _ = json.Marshal(jsonPatched)

	// unpatched version
	json.Unmarshal([]byte(line), &jsonUnpatched)
	var unpatched, _ = json.Marshal(jsonUnpatched)

	if bytes.Equal(patched, unpatched) {
		t.Fatal("Expected a difference between patched and unpatched wifi bug fix")
	}
}

func TestSampleLoopSingleEvent(t *testing.T) {
	titleChangeEvents := make(chan interface{}, 100)
	stopSamples := make(chan interface{}, 1)

	titleChangeEvents <- "changed"
	count := 0
	runSampleLoop(stopSamples, titleChangeEvents, func() {
		stopSamples <- "stop"
		count++
	})

	if count != 1 {
		t.Fatal("Expected only 1 stop event")
	}
}

func TestSampleLoopMultipleEvents(t *testing.T) {
	titleChangeEvents := make(chan interface{}, 100)
	stopSamples := make(chan interface{}, 1)

	titleChangeEvents <- "changed"
	titleChangeEvents <- "changed"
	titleChangeEvents <- "changed"
	titleChangeEvents <- "changed"

	count := 0
	runSampleLoop(stopSamples, titleChangeEvents, func() {
		stopSamples <- "stop"
		count++
	})

	if count != 1 {
		t.Fatal("Expected only 1 stop event")
	}
}

func TestFindPidsByProcessName(t *testing.T) {
	cmd := exec.Command("sleep", "2")
	cmd.Start()
	currentPid := cmd.Process.Pid
	pids, err := findPidsByProcessName("sleep")
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