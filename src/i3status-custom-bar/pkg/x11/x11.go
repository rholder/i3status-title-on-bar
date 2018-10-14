package x11

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type X11 struct {
	XConnection *xgb.Conn
	RootWindow xproto.Window
	ActiveWindowAtom xproto.Atom

	// this is the canonical window title atom, always the real title
	WindowNameAtom xproto.Atom

	// common window title atoms that also indicate the title has been updated
	WindowName2Atom xproto.Atom
	WindowName3Atom xproto.Atom
}

func New() (X11) {
	xConnection, err := xgb.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	return X11 {
		XConnection: xConnection,
		RootWindow: xproto.Setup(xConnection).DefaultScreen(xConnection).Root,
		ActiveWindowAtom: fetchAtom(xConnection, "_NET_ACTIVE_WINDOW"),
		WindowNameAtom: fetchAtom(xConnection, "_NET_WM_NAME"),
		WindowName2Atom: fetchAtom(xConnection, "WM_NAME"),
		WindowName3Atom: fetchAtom(xConnection, "_WM_NAME"),
	}
}

func (x11 X11) ActiveWindowTitle() (string) {
    // Get the actual value of _NET_ACTIVE_WINDOW.
	reply, err := xproto.GetProperty(x11.XConnection, false, x11.RootWindow, x11.ActiveWindowAtom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		// no title on error
		return ""
	}
	windowId := xproto.Window(xgb.Get32(reply.Value))

	// Now get the value of _NET_WM_NAME for the active window.
	reply, err = xproto.GetProperty(x11.XConnection, false, windowId, x11.WindowNameAtom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		// no title on error
		return ""
	}
	return string(reply.Value)
}

func (x11 X11) BeginTitleChangeDetection(stderr io.Writer, onChange func()) (error) {
	// subscribe to events from the root window
	xproto.ChangeWindowAttributes(x11.XConnection, x11.RootWindow,
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
		ev, xerr := x11.XConnection.WaitForEvent()
		if ev == nil && xerr == nil {
			fmt.Fprintln(stderr, "Both event and error are nil. Exiting...")
			return errors.New("Both event and error are nil. Exiting...")
		}

		if ev != nil {
			switch v := ev.(type) {
			case xproto.PropertyNotifyEvent:
				switch v.Atom {
				case x11.WindowNameAtom, x11.WindowName2Atom, x11.WindowName3Atom:
					onChange()
				case x11.ActiveWindowAtom:
					// subscribe to events of all windows as they are activated
					onChange()
					reply, err := xproto.GetProperty(x11.XConnection, false, x11.RootWindow, x11.ActiveWindowAtom,
						xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
					if err != nil {
						fmt.Fprintln(stderr, err)
						return err
					}
					windowId := xproto.Window(xgb.Get32(reply.Value))
					xproto.ChangeWindowAttributes(x11.XConnection, windowId,
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

func fetchAtom(c *xgb.Conn, name string) xproto.Atom {
	// Get the atom id (i.e., intern an atom) of "name".
	cookie, err := xproto.InternAtom(c, true, uint16(len(name)), name).Reply()
	if err != nil {
		log.Fatal(err)
	}
	return cookie.Atom
}