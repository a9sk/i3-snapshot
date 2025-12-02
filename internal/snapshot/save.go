package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	i3internal "github.com/a9sk/i3-snapshot/internal/i3"
	"github.com/a9sk/i3-snapshot/internal/models"
	"github.com/a9sk/i3-snapshot/internal/proc"
	"go.i3wm.org/i3"
)

// Save captures all workspace layouts and associated commands into a JSON file.
func Save(name string) error {
	tree := i3internal.GetTree()
	if tree.Root == nil {
		return fmt.Errorf("i3 tree root is nil")
	}

	// collect all workspaces from the tree
	workspaces := getAllWorkspaces(tree.Root)
	if len(workspaces) == 0 {
		return fmt.Errorf("no workspaces found in i3 tree")
	}

	snap := buildSnapshot(name, workspaces)

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
// Returns an error if no focused workspace is found.
// Note: The focused node is usually a window (leaf), not the workspace container.
// We track the current workspace as we walk and return it when we find a focused node.
func getWorkspaceTree() (*i3.Node, error) {
	tree := i3internal.GetTree()

	if tree.Root == nil {
		return nil, fmt.Errorf("i3 tree root is nil")
	}

	var focusedWorkspace *i3.Node

	var walk func(n *i3.Node, currentWS *i3.Node)
	walk = func(n *i3.Node, currentWS *i3.Node) {
		if n == nil || focusedWorkspace != nil {
			return
		}

		if n.Type == i3.WorkspaceNode {
			currentWS = n
		}

		// check if THIS node is the one with focus
		if n.Focused {
			// f the workspace itself is focused (empty), currentWS is n
			// if a winow inside is focused, currentWS is the parent workspace
			focusedWorkspace = currentWS
			return
		}

		// recurse into children, passing down the current workspace
		for i := range n.Nodes {
			walk(n.Nodes[i], currentWS)
		}
		for i := range n.FloatingNodes {
			walk(n.FloatingNodes[i], currentWS)
		}
	}

	walk(tree.Root, nil)

	if focusedWorkspace == nil {
		return nil, fmt.Errorf("no focused workspace found in i3 tree")
	}
	return focusedWorkspace, nil
}

// getAllWorkspaces collects all workspace nodes from the i3 tree.
// Filters out internal i3 workspaces (like __i3_scratch) that cannot be switched to.
func getAllWorkspaces(root *i3.Node) []*i3.Node {
	var workspaces []*i3.Node
	var walk func(n *i3.Node)
	walk = func(n *i3.Node) {
		if n == nil {
			return
		}
		if n.Type == i3.WorkspaceNode {
			// skip internal i3 workspaces (they start with __i3_)
			if !strings.HasPrefix(n.Name, "__i3_") {
				workspaces = append(workspaces, n)
			}
		}
		for i := range n.Nodes {
			walk(n.Nodes[i])
		}
		for i := range n.FloatingNodes {
			walk(n.FloatingNodes[i])
		}
	}
	walk(root)
	return workspaces
}

// buildSnapshot converts multiple i3 workspace nodes + /proc data into the Snapshot model.
func buildSnapshot(name string, workspaces []*i3.Node) models.Snapshot {
	snap := models.Snapshot{
		Name: name,
	}

	for _, ws := range workspaces {
		var windows []models.WindowRef
		root, windows := convertNode(ws)
		snap.Workspaces = append(snap.Workspaces, models.WorkspaceSnapshot{
			Name:    ws.Name,
			Root:    root,
			Windows: windows,
		})
	}

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
	// skip system windows that shouldn't be restored (i3bar, cursor, etc.)
	if n.Window != 0 {
		wp := n.WindowProperties
		// filter out system windows that shouldn't be restored
		skipWindow := wp.Class == "i3bar" || wp.Instance == "i3bar" || wp.Class == "i3status" ||
			wp.Class == "" || wp.Instance == "" // windows without class/instance are likely invalid

		if skipWindow {
			// skip this window, but still recurse into children
		} else {
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
