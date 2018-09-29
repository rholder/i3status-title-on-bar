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

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

func fetchAtom(c *xgb.Conn, name string) xproto.Atom {
	// Get the atom id (i.e., intern an atom) of "name".
	cookie, err := xproto.InternAtom(c, true, uint16(len(name)), name).Reply()
	if err != nil {
		log.Fatal(err)
	}
	return cookie.Atom
}

func fetchActiveWindowTitle(c *xgb.Conn, window xproto.Window, activeAtom xproto.Atom, nameAtom xproto.Atom) string {
	// Get the actual value of _NET_ACTIVE_WINDOW.
	reply, err := xproto.GetProperty(c, false, window, activeAtom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		// no title on error
		return ""
	}
	windowId := xproto.Window(xgb.Get32(reply.Value))

	// Now get the value of _NET_WM_NAME for the active window.
	reply, err = xproto.GetProperty(c, false, windowId, nameAtom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		// no title on error
		return ""
	}
	return string(reply.Value)
}

func patchWifiBug(allJson []interface{}) {
	for _, rawEntry := range allJson {
		switch entry := rawEntry.(type) {
		case map[string]interface{}:
			if entry["name"] == "wireless" {
				full_text := entry["full_text"].(string)
				percent := full_text[len(full_text)-4:]
				if strings.HasPrefix(percent, "0") {
					// W: ONYX 067%"
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

func runTitleChangeDetectionLoop(titleChangeEvents chan<- string, stderr io.Writer) {
	X, err := xgb.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	// Get the window id of the root window.
	setup := xproto.Setup(X)
	root := setup.DefaultScreen(X).Root

	// Get the atom id (i.e., intern an atom) of "_NET_ACTIVE_WINDOW".
	activeAtom := fetchAtom(X, "_NET_ACTIVE_WINDOW")

	// Get the atom id (i.e., intern an atom) of all the possible window name properties
	nameAtom := fetchAtom(X, "_NET_WM_NAME")
	otherNameAtom := fetchAtom(X, "WM_NAME")
	legacyNameAtom := fetchAtom(X, "_WM_NAME")

	atoms := make(map[xproto.Atom]string)
	atoms[activeAtom] = "_NET_ACTIVE_WINDOW"
	atoms[nameAtom] = "_NET_WM_NAME"
	atoms[otherNameAtom] = "WM_NAME"
	atoms[legacyNameAtom] = "_WM_NAME"

	// subscribe to events from the root window
	xproto.ChangeWindowAttributes(X, root,
		xproto.CwEventMask,
		[]uint32{ // values must be in the order defined by the protocol
			xproto.EventMaskStructureNotify |
				xproto.EventMaskPropertyChange})

	// Start the main event loop.
	for {
		// WaitForEvent either returns an event or an error and never both.
		// If both are nil, then something went wrong and the loop should be
		// halted.
		//
		// An error can only be seen here as a response to an unchecked
		// request.
		ev, xerr := X.WaitForEvent()
		if ev == nil && xerr == nil {
			fmt.Fprintln(stderr, "Both event and error are nil. Exiting...")
			return
		}

		if ev != nil {
			switch v := ev.(type) {
			case xproto.PropertyNotifyEvent:
				switch v.Atom {
				case nameAtom, otherNameAtom, legacyNameAtom:
					titleChangeEvents <- "changed"
					//fmt.Printf("Title change detected: %s\n", atoms[v.Atom])
				case activeAtom:
					// subscribe to events of all windows as they are activated
					titleChangeEvents <- "changed"
					reply, err := xproto.GetProperty(X, false, root, activeAtom,
						xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
					if err != nil {
						fmt.Fprintln(stderr, err)
						return
					}
					windowId := xproto.Window(xgb.Get32(reply.Value))
					xproto.ChangeWindowAttributes(X, windowId,
						xproto.CwEventMask,
						[]uint32{ // values must be in the order defined by the protocol
							xproto.EventMaskStructureNotify |
								xproto.EventMaskPropertyChange})
				default:
					// ignore everything else
					//fmt.Printf("Not title: %d\n", v.Atom)
				}
			}
		}

		if xerr != nil {
			fmt.Fprintln(stderr, xerr)
		}
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

func runJsonParsingLoop(stdin io.Reader, stdout io.Writer, stderr io.Writer) int {

	X, err := xgb.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	// Get the window id of the root window.
	setup := xproto.Setup(X)
	root := setup.DefaultScreen(X).Root

	// Get the atom id (i.e., intern an atom) of "_NET_ACTIVE_WINDOW".
	activeAtom := fetchAtom(X, "_NET_ACTIVE_WINDOW")

	// Get the atom id (i.e., intern an atom) of "_NET_WM_NAME".
	nameAtom := fetchAtom(X, "_NET_WM_NAME")

	// read from input
	scanner := bufio.NewScanner(stdin)

	// Skip the first line which contains the version header.
	// {"version":1}
	if !scanner.Scan() {
		// TODO happens way too often, be more resilient to bad scanner starts from stdin
		return 3
	}
	line := strings.TrimSpace(scanner.Text())
	fmt.Fprintf(stdout, "%s\n", line)

	// The second line contains the start of the infinite array.
	// [
	if !scanner.Scan() {
		return 4
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
		title := fetchActiveWindowTitle(X, root, activeAtom, nameAtom)
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
	return 7
}

func main() {
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr

	titleChangeEvents := make(chan string, 100)
	go runTitleChangeDetectionLoop(titleChangeEvents, stderr)
	go runSignalLoop(titleChangeEvents)

	exitCode := runJsonParsingLoop(stdin, stdout, stderr)
	os.Exit(exitCode)
}
