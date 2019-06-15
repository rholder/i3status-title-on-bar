package window

import (
	"errors"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type X11 struct {
	XConnection      *xgb.Conn
	RootWindow       xproto.Window
	ActiveWindowAtom xproto.Atom

	// this is the canonical window title atom, always the real title
	WindowNameAtom xproto.Atom

	// common window title atoms that also indicate the title has been updated
	WindowName2Atom xproto.Atom
	WindowName3Atom xproto.Atom
}

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

// Get the currently active window.
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

// Get the title of the given window.
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

func (x11 X11) subscribeToWindowChangeEvents(window xproto.Window) {
	// subscribe this XConnection to changes in window attributes, like the title attribute
	xproto.ChangeWindowAttributes(x11.XConnection, window,
		xproto.CwEventMask,
		[]uint32{ // values must be in the order defined by the protocol
			xproto.EventMaskStructureNotify |
				xproto.EventMaskPropertyChange})
}

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

func (x11 X11) BeginTitleChangeDetection(onChange func(), onError func(error)) error {
	// subscribe to events from the root window
	x11.subscribeToWindowChangeEvents(x11.RootWindow)

	// Start the main event loop.
	// TODO refactor this to remove the infinite loop
	for {
		// WaitForEvent either returns an event or an error and never both.
		// If both are nil, then something went wrong and the loop should be
		// halted.
		//
		// An error can only be seen here as a response to an unchecked
		// request.
		ev, xerr := x11.XConnection.WaitForEvent()
		if ev == nil && xerr == nil {
			err := errors.New("Both event and error are nil. Exiting...")
			onError(err)
			return err
		}

		if ev != nil {
			switch v := ev.(type) {
			case xproto.PropertyNotifyEvent:
				switch v.Atom {
				case x11.WindowNameAtom, x11.WindowName2Atom, x11.WindowName3Atom:
					onChange()
				case x11.ActiveWindowAtom:
					onChange()

					// subscribe to events of all windows as they are activated
					activeWindow, err := x11.activeWindow()
					if err != nil {
						onError(err)
					} else {
						x11.subscribeToWindowChangeEvents(*activeWindow)
					}
				default:
					// ignore everything else
				}
			}
		}

		if xerr != nil {
			onError(xerr)
		}
	}
}

func fetchAtom(c *xgb.Conn, name string) (*xproto.Atom, error) {
	// Get the atom id (i.e., intern an atom) of "name".
	cookie, err := xproto.InternAtom(c, true, uint16(len(name)), name).Reply()
	if err != nil {
		return nil, err
	}
	return &cookie.Atom, nil
}
