package i3

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/rholder/i3status-title-on-bar/pkg/window"
)

const (
	OK                            int = 0
	BadInputOpenErrorCode         int = 3
	BadInputHeaderErrorCode       int = 4
	BadInputJSONErrorCode         int = 5
	BadCreatedJSONErrorCode       int = 6
	BadEOFErrorCode               int = 7
	MissingStatusProcessErrorCode int = 8
	BadDisplayErrorCode           int = 9
)

func scannerError(out io.Writer, scanner *bufio.Scanner, errorCode int) int {
	if scanner.Err() != nil {
		fmt.Fprintf(out, "ERROR from bufio.Scanner: %s\n", scanner.Err())
	}
	return errorCode
}

func newTitleNode(color string, title string) map[string]string {
	return map[string]string{
		"name":      "window_title",
		"full_text": title,
		"color":     color}
}

func truncateAndPad(value string, fixedWidth int) string {
	safeSubstring := value
	if len(value) > fixedWidth {
		safeSubstring = value[0:fixedWidth]
	}

	template := "%-" + strconv.Itoa(fixedWidth) + "s"
	return fmt.Sprintf(template, safeSubstring)
}

func RunJsonParsingLoop(stdin io.Reader, stdout io.Writer, stderr io.Writer, windowAPI window.WindowAPI,
	color string, appendEnd bool, fixedWidth int) int {

	// read from input
	scanner := bufio.NewScanner(stdin)

	// Skip the first line which contains the version header.
	// {"version":1}
	if !scanner.Scan() {
		// TODO happens way too often, be more resilient to bad scanner starts from stdin
		return scannerError(stderr, scanner, BadInputOpenErrorCode)
	}
	line := strings.TrimSpace(scanner.Text())
	fmt.Fprintf(stdout, "%s\n", line)

	// The second line contains the start of the infinite array.
	// [
	if !scanner.Scan() {
		return scannerError(stderr, scanner, BadInputHeaderErrorCode)
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
			return BadInputJSONErrorCode
		}

		// inject window title node first
		title := windowAPI.ActiveWindowTitle()
		if fixedWidth > 0 {
			title = truncateAndPad(title, fixedWidth)
		}
		titleNode := newTitleNode(color, title)

		// bolt together the JSON
		var allJson []interface{}
		if appendEnd {
			allJson = append(allJson, parsed...)
			allJson = append(allJson, titleNode)
		} else {
			allJson = append(allJson, titleNode)
			allJson = append(allJson, parsed...)
		}

		// parsed = append(titleNode, parsed...) // TODO figure out how to do this cleanly
		parsedJson, err := json.Marshal(allJson)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return BadCreatedJSONErrorCode
		}

		// output i3bar JSON
		fmt.Fprintf(stdout, "%s%s\n", prefix, parsedJson)
	}

	if scanner.Err() != nil {
		return scannerError(stderr, scanner, BadEOFErrorCode)
	} else {
		// we hit EOF normally, everything is fine
		return OK
	}
}
