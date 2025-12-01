package proc

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
)

// GetPIDFromWindowID attempts to resolve the PID for a given X11 window ID (XID)
// using the _NET_WM_PID property. It returns an error if the PID cannot be
// determined (e.g. property missing, skill issues, permissions, or X11 issues).
//
// This is best-effort: callers should treat errors as "no PID available".
func GetPIDFromWindowID(xid uint32) (int, error) {
	if xid == 0 {
		return 0, fmt.Errorf("invalid window id: 0")
	}

	xu, err := xgbutil.NewConn()
	if err != nil {
		return 0, fmt.Errorf("connecting to X11: %w", err)
	}
	defer xu.Conn().Close()

	win := xproto.Window(xid)

	// _NET_WM_PID is a standard EWMH property (CARDINAL, 32-bit)
	atom, err := xproto.InternAtom(xu.Conn(), true, uint16(len("_NET_WM_PID")), "_NET_WM_PID").Reply()
	if err != nil {
		return 0, fmt.Errorf("interning _NET_WM_PID atom: %w", err)
	}

	prop, err := xproto.GetProperty(xu.Conn(), false, win, atom.Atom, xproto.AtomCardinal, 0, 1).Reply()
	if err != nil {
		return 0, fmt.Errorf("reading _NET_WM_PID property: %w", err)
	}

	if prop == nil || prop.ValueLen == 0 {
		return 0, fmt.Errorf("_NET_WM_PID property empty for window 0x%x", xid)
	}

	// Value is a 32-bit CARDINAL; interpret first 4 bytes as little-endian uint32, then pray
	if len(prop.Value) < 4 {
		return 0, fmt.Errorf("_NET_WM_PID value too short for window 0x%x", xid)
	}
	pid := uint32(prop.Value[0]) |
		uint32(prop.Value[1])<<8 |
		uint32(prop.Value[2])<<16 |
		uint32(prop.Value[3])<<24

	return int(pid), nil
}
