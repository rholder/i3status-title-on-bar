package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

type TestWindowAPI struct{}

func (testWindowAPI TestWindowAPI) ActiveWindowTitle() string {
	return "foo"
}

func (testWindowAPI TestWindowAPI) BeginTitleChangeDetection(stderr io.Writer, onChange func()) error {
	return nil
}

func TestJsonParsingLoopEmptyInput(t *testing.T) {
	lines := strings.NewReader("")
	errorCode := runJsonParsingLoop(lines, nil, nil, nil, "#00FF00")
	if errorCode != 3 {
		t.Fatal("Expected error from parsing loop")
	}
}

func TestJsonParsingLoopNewlineInput(t *testing.T) {
	lines := strings.NewReader("\n")
	stdout := os.Stdout
	stderr := os.Stderr
	errorCode := runJsonParsingLoop(lines, stdout, stderr, nil, "#00FF00")
	if errorCode != 4 {
		t.Fatal("Expected error from parsing loop")
	}
}

func TestJsonParsingLoopHappyBlankInput(t *testing.T) {
	lines := strings.NewReader("\n\n")
	stdout := os.Stdout
	stderr := os.Stderr
	errorCode := runJsonParsingLoop(lines, stdout, stderr, nil, "#00FF00")
	if errorCode != 0 {
		t.Fatal("Expected no error from parsing loop")
	}
}

func TestJsonParsingLoopBadJSONInput(t *testing.T) {
	lines := strings.NewReader("\n\nPOTATO")
	stdout := os.Stdout
	stderr := os.Stderr
	errorCode := runJsonParsingLoop(lines, stdout, stderr, nil, "#00FF00")
	if errorCode != 5 {
		t.Fatal("Expected error from parsing loop")
	}
}

func TestJsonParsingLoopGoodJSONInput(t *testing.T) {
	input := "\n\n" +
		`[{"name":"wireless","instance":"wlp1s0","color":"#00FF00","markup":"none","full_text":"W: SOME_WIFI_SSID 067%"}]` +
		"\n" +
		`,[{"name":"wireless","instance":"wlp1s0","color":"#00FF00","markup":"none","full_text":"W: SOME_WIFI_SSID 064%"}]`
	lines := strings.NewReader(input)
	stdout := os.Stdout
	stderr := os.Stderr
	windowAPI := TestWindowAPI{}
	errorCode := runJsonParsingLoop(lines, stdout, stderr, windowAPI, "#00FF00")
	if errorCode != 0 {
		t.Fatal("Expected no error from parsing loop")
	}
}
