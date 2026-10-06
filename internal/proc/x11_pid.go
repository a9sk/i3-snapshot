package proc

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
)

// X11Resolver resolves X11 window IDs to process IDs using a single X11
// connection. Creating one resolver per save run avoids opening a connection
// for every window.
type X11Resolver struct {
	xu   *xgbutil.XUtil
	atom xproto.Atom
}

// NewX11Resolver opens a new X11 connection and internes the _NET_WM_PID atom.
func NewX11Resolver() (*X11Resolver, error) {
	xu, err := xgbutil.NewConn()
	if err != nil {
		return nil, fmt.Errorf("connecting to X11: %w", err)
	}

	atom, err := xproto.InternAtom(xu.Conn(), true, uint16(len("_NET_WM_PID")), "_NET_WM_PID").Reply()
	if err != nil {
		xu.Conn().Close()
		return nil, fmt.Errorf("interning _NET_WM_PID atom: %w", err)
	}

	return &X11Resolver{xu: xu, atom: atom.Atom}, nil
}

// PID resolves the PID for a single X11 window ID using the shared connection.
func (r *X11Resolver) PID(xid uint32) (int, error) {
	if xid == 0 {
		return 0, fmt.Errorf("invalid window id: 0")
	}

	win := xproto.Window(xid)
	prop, err := xproto.GetProperty(r.xu.Conn(), false, win, r.atom, xproto.AtomCardinal, 0, 1).Reply()
	if err != nil {
		return 0, fmt.Errorf("reading _NET_WM_PID property: %w", err)
	}
	if prop == nil || prop.ValueLen == 0 {
		return 0, fmt.Errorf("_NET_WM_PID property empty for window 0x%x", xid)
	}
	if len(prop.Value) < 4 {
		return 0, fmt.Errorf("_NET_WM_PID value too short for window 0x%x", xid)
	}

	pid := uint32(prop.Value[0]) |
		uint32(prop.Value[1])<<8 |
		uint32(prop.Value[2])<<16 |
		uint32(prop.Value[3])<<24
	return int(pid), nil
}

// Close closes the underlying X11 connection.
func (r *X11Resolver) Close() {
	if r != nil && r.xu != nil {
		r.xu.Conn().Close()
	}
}

// GetPIDFromWindowID is a convenience wrapper that opens a fresh X11
// connection, resolves a single window ID, and closes the connection.
func GetPIDFromWindowID(xid uint32) (int, error) {
	r, err := NewX11Resolver()
	if err != nil {
		return 0, err
	}
	defer r.Close()
	return r.PID(xid)
}
