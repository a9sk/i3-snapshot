package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	i3internal "github.com/a9sk/i3-snapshot/internal/i3"
	"github.com/a9sk/i3-snapshot/internal/models"
	"github.com/a9sk/i3-snapshot/internal/proc"
	"go.i3wm.org/i3"
)

// Save captures the current workspace layout and associated commands into a JSON file.
func Save(name string) error {
	tree := getWorkspaceTree()
	snap := buildSnapshot(name, tree)

	// resolve output path: ~/.config/i3-snapshot/saves/<name>.json
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("resolving config dir: %w", err)
	}
	saveDir := filepath.Join(configDir, "i3-snapshot", "saves")
	if err := os.MkdirAll(saveDir, 0o755); err != nil {
		return fmt.Errorf("creating save dir %s: %w", saveDir, err)
	}

	outPath := filepath.Join(saveDir, name+".json")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating snapshot file %s: %w", outPath, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return fmt.Errorf("encoding snapshot: %w", err)
	}

	return nil
}

// getWorkspaceTree pulls the current i3 tree and returns the focused workspace node.
func getWorkspaceTree() *i3.Node {
	tree := i3internal.GetTree()

	var focused *i3.Node
	var walk func(n *i3.Node)
	walk = func(n *i3.Node) {
		if n == nil || focused != nil {
			return
		}
		if n.Type == i3.WorkspaceNode && n.Focused {
			focused = n
			return
		}
		for i := range n.Nodes {
			walk(n.Nodes[i])
		}
		for i := range n.FloatingNodes {
			walk(n.FloatingNodes[i])
		}
	}
	root := tree.Root
	walk(root)
	if focused == nil {
		// fallback: ret only root if we cant find a focused workspace
		return root
	}
	return focused
}

// buildSnapshot converts the i3 workspace node + /proc data into our Snapshot model.
func buildSnapshot(name string, root *i3.Node) models.Snapshot {
	snap := models.Snapshot{
		Name:      name,
		Workspace: root.Name,
	}

	var windows []models.WindowRef
	snap.Root, windows = convertNode(root)
	snap.Windows = windows
	return snap
}

// convertNode walks an i3.Node tree and returns the LayoutNode plus a flat list of WindowRefs.
func convertNode(n *i3.Node) (models.LayoutNode, []models.WindowRef) {
	node := models.LayoutNode{
		ID:       int64(n.ID),
		Type:     string(n.Type),
		Layout:   string(n.Layout),
		Name:     n.Name,
		Border:   string(n.Border),
		Rect:     models.Rect{X: int(n.Rect.X), Y: int(n.Rect.Y), Width: int(n.Rect.Width), Height: int(n.Rect.Height)},
		WindowID: int(n.Window),
		Focused:  n.Focused,
	}

	var allWindows []models.WindowRef

	// capture properties/command if it is a window
	if n.Window != 0 {
		wp := n.WindowProperties
		node.WindowClass = wp.Class
		node.WindowInst = wp.Instance
		node.WindowTitle = wp.Title

		// resolve PID via X11 (_NET_WM_PID) using the X11 window id from n.Window
		// errors are treated as "no PID available" so snapshots remain usable
		cmd := ""
		cwd := ""
		if pid, err := proc.GetPIDFromWindowID(uint32(n.Window)); err == nil && pid > 0 {
			if c, e := proc.GetCommandFromPID(pid); e == nil {
				cmd = c
			}
			if d, e := proc.GetCWDFromPID(pid); e == nil {
				cwd = d
			}
		}

		w := models.WindowRef{
			NodeID:   int64(n.ID),
			Class:    wp.Class,
			Instance: wp.Instance,
			Title:    wp.Title,
			Command:  cmd,
			Cwd:      cwd,
		}
		allWindows = append(allWindows, w)
	}

	// recurse into tiling and floating children
	for i := range n.Nodes {
		childNode, childWindows := convertNode(n.Nodes[i])
		node.Nodes = append(node.Nodes, childNode)
		allWindows = append(allWindows, childWindows...)
	}
	for i := range n.FloatingNodes {
		childNode, childWindows := convertNode(n.FloatingNodes[i])
		node.FloatingNodes = append(node.FloatingNodes, childNode)
		allWindows = append(allWindows, childWindows...)
	}

	return node, allWindows
}
