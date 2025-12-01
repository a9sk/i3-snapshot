package models

// Snapshot is the top-level structure written to disk.
// It contains a simplified layout tree and per-window launch information.
type Snapshot struct {
	Name      string      `json:"name"`
	Workspace string      `json:"workspace"`
	Root      LayoutNode  `json:"root"`
	Windows   []WindowRef `json:"windows"`
}

// LayoutNode represents a container in the i3 tree (output, workspace, split, tabbed, etc.).
// This keeps only the fields we care about for reconstructing geometry and hierarchy.
type LayoutNode struct {
	ID           int64        `json:"id"`
	Type         string       `json:"type"`                  // e.g. "root", "output", "workspace", "con", "floating_con"
	Layout       string       `json:"layout,omitempty"`      // e.g. "splith", "splitv", "tabbed", ...
	Name         string       `json:"name,omitempty"`        // workspace name, window title, etc.
	Border       string       `json:"border,omitempty"`      // for completeness
	Rect         Rect         `json:"rect"`                  // container rectangle
	WindowID     int          `json:"window_id,omitempty"`   // X11 window ID, if any
	WindowClass  string       `json:"window_class,omitempty"`
	WindowInst   string       `json:"window_instance,omitempty"`
	WindowTitle  string       `json:"window_title,omitempty"`
	Focused      bool         `json:"focused,omitempty"`
	Nodes        []LayoutNode `json:"nodes,omitempty"`          // tiling children
	FloatingNodes []LayoutNode `json:"floating_nodes,omitempty"` // floating children
}

// Rect is a simple geometry rectangle.
type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// WindowRef ties a window's identifying criteria to its execution details.
type WindowRef struct {
	NodeID   int64  `json:"node_id"`             // ID of the LayoutNode where this window lives
	Class    string `json:"class,omitempty"`     // X11 class
	Instance string `json:"instance,omitempty"`  // X11 instance
	Title    string `json:"title,omitempty"`     // window title

	Command string `json:"command"`              // full command line from /proc/[pid]/cmdline
	Cwd     string `json:"cwd,omitempty"`        // working directory from /proc/[pid]/cwd
}


