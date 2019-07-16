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

package window

import (
	"errors"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

// X11 creates and manages the currently active X11 connection.
type X11 struct {
	// This is the underlying X11 connection.
	XConnection *xgb.Conn

	// This is the root window.
	RootWindow xproto.Window

	// The value of this atom should be the currently active window identifier.
	// NOTE: Sometimes there is no currently active window.
	ActiveWindowAtom xproto.Atom

	// This is the canonical window title atom. When the value for it is
	// retrieved, it should always return the current real window title.
	WindowNameAtom xproto.Atom

	// This is a common window title atom. Any changes that occur for it may
	// indicate the title has been updated.
	WindowName2Atom xproto.Atom

	// This is another common window title atom. Any changes that occur for it
	// may indicate the title has been updated.
	WindowName3Atom xproto.Atom
}

// NewX11 starts up a new connection to an X11 display server, interning all
// necessary atoms up front and setting up the root window.
func NewX11() (*X11, error) {
	xConnection, err := xgb.NewConn()
	if err != nil {
		return nil, err
	}

	rootWindow := xproto.Setup(xConnection).DefaultScreen(xConnection).Root

	activeWindowAtom, err := fetchAtom(xConnection, "_NET_ACTIVE_WINDOW")
	if err != nil {
		return nil, err
	}

	windowNameAtom, err := fetchAtom(xConnection, "_NET_WM_NAME")
	if err != nil {
		return nil, err
	}

	windowName2Atom, err := fetchAtom(xConnection, "WM_NAME")
	if err != nil {
		return nil, err
	}

	windowName3Atom, err := fetchAtom(xConnection, "_WM_NAME")
	if err != nil {
		return nil, err
	}

	return &X11{
		XConnection:      xConnection,
		RootWindow:       rootWindow,
		ActiveWindowAtom: *activeWindowAtom,
		WindowNameAtom:   *windowNameAtom,
		WindowName2Atom:  *windowName2Atom,
		WindowName3Atom:  *windowName3Atom,
	}, nil
}

// Get the currently active xproto.Window.
func (x11 X11) activeWindow() (*xproto.Window, error) {
	// Get the actual value of _NET_ACTIVE_WINDOW.
	reply, err := xproto.GetProperty(x11.XConnection, false, x11.RootWindow, x11.ActiveWindowAtom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return nil, err
	}
	window := xproto.Window(xgb.Get32(reply.Value))
	return &window, nil
}

// Get the title attribute as a string of the given xproto.Window.
func (x11 X11) windowTitleProperty(window xproto.Window) (*string, error) {
	// Now get the value of _NET_WM_NAME for the active window.
	reply, err := xproto.GetProperty(x11.XConnection, false, window, x11.WindowNameAtom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return nil, err
	}
	title := string(reply.Value)
	return &title, nil
}

// Subscribe the current XConnection to change events in window attributes (like
// the title attribute) for the given xproto.Window.
func (x11 X11) subscribeToWindowChangeEvents(window xproto.Window) {
	xproto.ChangeWindowAttributes(x11.XConnection, window,
		xproto.CwEventMask,
		[]uint32{ // values must be in the order defined by the protocol
			xproto.EventMaskStructureNotify |
				xproto.EventMaskPropertyChange})
}

// ActiveWindowTitle returns the currently active window title or an empty
// string if one is not available.
func (x11 X11) ActiveWindowTitle() string {
	activeWindow, err := x11.activeWindow()
	if err != nil {
		// no title on error
		return ""
	}

	windowTitle, err := x11.windowTitleProperty(*activeWindow)
	if err != nil {
		// no title on error
		return ""
	}
	return *windowTitle
}

// DetectWindowTitleChanges blocks and starts detecting changes in window
// titles. When a change is detected, the onChange function is called and when a
// non-fatal error occurs the onError function is called for that error.
func (x11 X11) DetectWindowTitleChanges(onChange func(), onError func(error)) error {
	// Subscribe to events from the root window.
	x11.subscribeToWindowChangeEvents(x11.RootWindow)

	// TODO Refactor this infinite loop when xgb supports a clean shut down.

	// Start the main event loop.
	for {
		// WaitForEvent either returns an event or an error and never both.
		// If both are nil, then something went wrong and the loop should be
		// halted.
		//
		// An error can only be seen here as a response to an unchecked
		// request.
		ev, xerr := x11.XConnection.WaitForEvent()
		if ev == nil && xerr == nil {
			err := errors.New("Both event and error are nil from XConnection, exiting X11 event loop")
			onError(err)
			return err
		}

		// Filter this event down to only what we care about.
		if ev != nil {
			switch v := ev.(type) {
			case xproto.PropertyNotifyEvent:
				switch v.Atom {
				case x11.WindowNameAtom, x11.WindowName2Atom, x11.WindowName3Atom:
					onChange()
				case x11.ActiveWindowAtom:
					onChange()

					// Subscribe to events of all windows as they are activated.
					// This is the trick to get complex windows that change
					// their titles as tabs are activated to be detected.
					activeWindow, err := x11.activeWindow()
					if err != nil {
						onError(err)
					} else {
						x11.subscribeToWindowChangeEvents(*activeWindow)
					}
				default:
					// Ignore everything else.
				}
			}
		}

		// An error from the X11 event loop is not fatal.
		if xerr != nil {
			onError(xerr)
		}
	}
}

// Get the atom id (i.e., intern an atom) of the given name.
func fetchAtom(c *xgb.Conn, name string) (*xproto.Atom, error) {
	cookie, err := xproto.InternAtom(c, true, uint16(len(name)), name).Reply()
	if err != nil {
		return nil, err
	}
	return &cookie.Atom, nil
}
