package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestJsonParsingLoopEmptyInput(t *testing.T) {
	emptyReader := strings.NewReader("")
	errorCode := runJsonParsingLoop(emptyReader, nil, nil, nil)
	if errorCode != 3 {
		t.Fatal("Expected error from parsing loop")
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
