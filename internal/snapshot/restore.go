package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/a9sk/i3-snapshot/internal/models"
	"go.i3wm.org/i3"
)

var (
	// getTree is a helper to access i3.GetTree from this package
	getTree = func() (i3.Tree, error) {
		return i3.GetTree()
	}
)

// Restore replays a previously saved snapshot by name.
// It:
//  1. loads ~/.config/i3-snapshot/saves/<name>.json
//  2. for each workspace: switches to it, applies layout, then launches commands
//  3. launches all recorded commands concurrently
func Restore(name string) error {
	snap, err := loadSnapshot(name)
	if err != nil {
		return err
	}

	// restore each workspace
	for _, ws := range snap.Workspaces {
		// skip invalid or internal i3 workspaces
		if ws.Name == "" || ws.Name == "root" || strings.HasPrefix(ws.Name, "__i3_") {
			continue
		}

		// switch to the workspace
		cmd := fmt.Sprintf("workspace %s", ws.Name)
		if _, err := i3.RunCommand(cmd); err != nil {
			return fmt.Errorf("switching to workspace %s: %w", ws.Name, err)
		}

		time.Sleep(200 * time.Millisecond)

		// extract workspace children for append_layout
		var layoutRoot *models.LayoutNode
		if ws.Root.Type == "workspace" {
			if len(ws.Root.Nodes) == 0 && len(ws.Root.FloatingNodes) == 0 {
				// empty workspace, skip layout but still launch windows for this workspace
				launchCommands(ws.Windows)
				continue
			}

			layoutRoot = &models.LayoutNode{
				Type:          "con",
				Layout:        ws.Root.Layout,
				Nodes:         ws.Root.Nodes,
				FloatingNodes: ws.Root.FloatingNodes,
				Rect:          ws.Root.Rect,
			}
		} else {
			layoutRoot = &ws.Root
		}

		// apply layout to this workspace
		if err := applyLayout(layoutRoot); err != nil {
			return fmt.Errorf("applying layout to workspace %s: %w", ws.Name, err)
		}

		// wait a bit for layout placeholders to be created
		time.Sleep(200 * time.Millisecond)

		// launch commands for THIS workspace while we're still on it
		// this ensures windows open in the correct workspace
		launchCommands(ws.Windows)

		// wait for windows to appear and get swallowed by placeholders
		// this is important for slow-starting apps like browsers
		waitForWindows(ws.Name, ws.Windows, 10*time.Second)
	}

	return nil
}

// loadSnapshot loads a snapshot JSON by name from the config directory.
func loadSnapshot(name string) (models.Snapshot, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return models.Snapshot{}, fmt.Errorf("resolving config dir: %w", err)
	}
	path := filepath.Join(configDir, "i3-snapshot", "saves", name+".json")

	f, err := os.Open(path)
	if err != nil {
		return models.Snapshot{}, fmt.Errorf("opening snapshot %s: %w", path, err)
	}
	defer f.Close()

	var snap models.Snapshot
	if err := json.NewDecoder(f).Decode(&snap); err != nil {
		return models.Snapshot{}, fmt.Errorf("decoding snapshot %s: %w", path, err)
	}
	return snap, nil
}

// applyLayout converts our LayoutNode to i3's expected format and applies it.
func applyLayout(root *models.LayoutNode) error {
	i3Root := convertToI3Layout(root)

	data, err := json.Marshal(i3Root)
	if err != nil {
		return fmt.Errorf("marshalling layout: %w", err)
	}

	// i3 expects a file for append_layout, so we write to a temp file and clean it up
	tmp, err := os.CreateTemp("", "i3-snapshot-layout-*.json")
	if err != nil {
		return fmt.Errorf("creating temp layout file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp layout file: %w", err)
	}
	tmp.Close()

	reply, err := i3.RunCommand(fmt.Sprintf("append_layout %s", tmp.Name()))
	if err != nil {
		return fmt.Errorf("running append_layout: %w", err)
	}
	_ = reply // TODO: in future we might inspect success/failure per command

	return nil
}

// convertToI3Layout converts our LayoutNode to i3's expected format with swallows.
// It filters out invalid windows (like Cursor, i3bar, etc.) that shouldn't be restored.
func convertToI3Layout(n *models.LayoutNode) models.I3LayoutNode {
	node := models.I3LayoutNode{
		Type:   n.Type,
		Layout: n.Layout,
		Border: n.Border,
		Rect:   n.Rect,
	}

	// only include ID and Name for non-workspace containers to avoid creating workspaces
	// workspace nodes should not have their name/ID in the layout
	if n.Type != "workspace" && n.Type != "root" && n.Type != "output" {
		if n.ID != 0 {
			node.ID = n.ID
		}
		if n.Name != "" {
			node.Name = n.Name
		}
	}

	// only add swallows for valid application windows
	// filter out system windows that i3 will reject
	if n.WindowClass != "" || n.WindowInst != "" || n.WindowTitle != "" {
		// skip invalid window classes (system windows)
		invalidClass := n.WindowClass == "i3bar" ||
			n.WindowClass == "i3status" || n.WindowClass == ""

		if !invalidClass {
			node.Swallows = []models.SwallowCriteria{
				{
					Class:    n.WindowClass,
					Instance: n.WindowInst,
					Title:    n.WindowTitle,
				},
			}
		}
	}

	for i := range n.Nodes {
		node.Nodes = append(node.Nodes, convertToI3Layout(&n.Nodes[i]))
	}
	for i := range n.FloatingNodes {
		node.FloatingNodes = append(node.FloatingNodes, convertToI3Layout(&n.FloatingNodes[i]))
	}

	return node
}

// waitForWindows waits for windows to appear and get swallowed by placeholders.
// It checks the i3 tree periodically to see if windows matching th criteria
// have appeared in the correct workspace. Returns after timeout or when all windows are found.
func waitForWindows(workspaceName string, expectedWindows []models.WindowRef, timeout time.Duration) {
	if len(expectedWindows) == 0 {
		return
	}

	deadline := time.Now().Add(timeout)
	checkInterval := 200 * time.Millisecond

	for time.Now().Before(deadline) {
		tree, err := getTree()
		if err != nil {
			time.Sleep(checkInterval)
			continue
		}

		// find the workspace in the tree
		var workspaceNode *i3.Node
		var findWorkspace func(n *i3.Node)
		findWorkspace = func(n *i3.Node) {
			if n == nil || workspaceNode != nil {
				return
			}
			if n.Type == i3.WorkspaceNode && n.Name == workspaceName {
				workspaceNode = n
				return
			}
			for i := range n.Nodes {
				findWorkspace(n.Nodes[i])
			}
		}
		findWorkspace(tree.Root)

		if workspaceNode == nil {
			time.Sleep(checkInterval)
			continue
		}

		// count how many expected windows we've found in this workspace
		// and also check for windows in other workspaces that should be moved here
		foundCount := 0
		var windowsToMove []*i3.Node

		var checkWindow func(n *i3.Node, inTargetWorkspace bool)
		checkWindow = func(n *i3.Node, inTargetWorkspace bool) {
			if n == nil {
				return
			}
			if n.Window != 0 {
				wp := n.WindowProperties
				// check if this window matches any of our expected windows
				for _, expected := range expectedWindows {
					if expected.Command == "" {
						continue // skip windows without commands
					}
					// match by class and instance (title can change)
					if (expected.Class == "" || wp.Class == expected.Class) &&
						(expected.Instance == "" || wp.Instance == expected.Instance) {
						if inTargetWorkspace {
							foundCount++
						} else {
							// window is in wrong workspace, mark it for moving
							windowsToMove = append(windowsToMove, n)
						}
						break
					}
				}
			}
			for i := range n.Nodes {
				checkWindow(n.Nodes[i], inTargetWorkspace)
			}
			for i := range n.FloatingNodes {
				checkWindow(n.FloatingNodes[i], inTargetWorkspace)
			}
		}

		// check the target workspace
		checkWindow(workspaceNode, true)

		// also check all other workspaces for windows that should be here
		var checkAllWorkspaces func(n *i3.Node)
		checkAllWorkspaces = func(n *i3.Node) {
			if n == nil {
				return
			}
			if n.Type == i3.WorkspaceNode && n.Name != workspaceName {
				checkWindow(n, false)
			}
			for i := range n.Nodes {
				checkAllWorkspaces(n.Nodes[i])
			}
		}
		checkAllWorkspaces(tree.Root)

		// move windows that appeared in wrong workspace
		for _, win := range windowsToMove {
			cmd := fmt.Sprintf("[id=\"%d\"] move workspace %s", win.Window, workspaceName)
			i3.RunCommand(cmd)
		}

		// if we found all windows (or most of them), we're done
		// we use "most" because some windows might not have commands saved
		expectedCount := 0
		for _, w := range expectedWindows {
			if w.Command != "" {
				expectedCount++
			}
		}
		if foundCount >= expectedCount || foundCount >= len(expectedWindows) {
			// give a bit more time for slow apps to fully initialize
			time.Sleep(500 * time.Millisecond)

			removePlaceholders(workspaceName, expectedWindows)
			return
		}

		time.Sleep(checkInterval)
	}

	// timeout reached, try to clean up placeholders anyway
	removePlaceholders(workspaceName, expectedWindows)
}

// removePlaceholders removes placeholder windows that weren't swallowed by real windows.
// Placeholders are containers with no actual window (Window == 0) that are waiting to be swallowed.
// We identify them by checking if they have no window and no children with windows.
func removePlaceholders(workspaceName string, expectedWindows []models.WindowRef) {
	tree, err := getTree()
	if err != nil {
		return
	}

	var workspaceNode *i3.Node
	var findWorkspace func(n *i3.Node)
	findWorkspace = func(n *i3.Node) {
		if n == nil || workspaceNode != nil {
			return
		}
		if n.Type == i3.WorkspaceNode && n.Name == workspaceName {
			workspaceNode = n
			return
		}
		for i := range n.Nodes {
			findWorkspace(n.Nodes[i])
		}
	}
	findWorkspace(tree.Root)

	if workspaceNode == nil {
		return
	}

	// first, collect all real windows that have appeared in this workspace
	realWindows := make(map[string]bool)
	var collectRealWindows func(n *i3.Node)
	collectRealWindows = func(n *i3.Node) {
		if n == nil {
			return
		}
		if n.Window != 0 {
			wp := n.WindowProperties
			// create a key from class+instance to identify this window
			key := fmt.Sprintf("%s|%s", wp.Class, wp.Instance)
			realWindows[key] = true
		}
		for i := range n.Nodes {
			collectRealWindows(n.Nodes[i])
		}
		for i := range n.FloatingNodes {
			collectRealWindows(n.FloatingNodes[i])
		}
	}
	collectRealWindows(workspaceNode)

	// now find placeholders: containers with Window == 0 that have no children with windows
	// and match expected windows that have already appeared (so the placeholder is no longer needed)
	var placeholders []*i3.Node
	var findPlaceholders func(n *i3.Node) bool
	findPlaceholders = func(n *i3.Node) bool {
		if n == nil {
			return false
		}

		// check if this node or any descendant has a real window
		hasRealWindow := false
		if n.Window != 0 {
			hasRealWindow = true
		}

		// check children recursively
		for i := range n.Nodes {
			if findPlaceholders(n.Nodes[i]) {
				hasRealWindow = true
			}
		}
		for i := range n.FloatingNodes {
			if findPlaceholders(n.FloatingNodes[i]) {
				hasRealWindow = true
			}
		}

		// if this is a leaf container (no children) with no window, it's likely a placeholder
		// layout containers have children, placeholders are leaf nodes waiting to be swallowed
		isLeaf := len(n.Nodes) == 0 && len(n.FloatingNodes) == 0
		if !hasRealWindow && n.Window == 0 && n.Type == i3.Con && isLeaf {
			// check if we have any expected windows for this workspace
			// if we do, and this is a leaf with no window, it's almost certainly a placeholder
			hasExpectedWindows := false
			for _, expected := range expectedWindows {
				if expected.Command != "" {
					hasExpectedWindows = true
					break
				}
			}

			if hasExpectedWindows {
				// any leaf container with Window == 0 in a workspace with expected windows
				// is a placeholder (layout containers have children, real windows have Window != 0)
				placeholders = append(placeholders, n)
				return false // don't recurse further, we're removing this
			}
		}

		return hasRealWindow
	}
	findPlaceholders(workspaceNode)

	// remove placeholders (in reverse order to avoid ID changes)
	for i := len(placeholders) - 1; i >= 0; i-- {
		placeholder := placeholders[i]
		cmd := fmt.Sprintf("[con_id=\"%d\"] kill", placeholder.ID)
		i3.RunCommand(cmd)
		// small delay to let i3 process the kill
	}
}

// launchCommands starts each window's command in its recorded working directory.
// Launch errors are logged to stderr but do not abort the whole restore.
func launchCommands(windows []models.WindowRef) {
	var wg sync.WaitGroup

	for _, w := range windows {
		cmdline := w.Command
		if cmdline == "" {
			continue
		}

		wg.Add(1)
		go func(w models.WindowRef) {
			defer wg.Done()

			// naive splitting on spaces for now; TODO:in future we may want a shell parser
			args := splitCommandLine(w.Command)
			if len(args) == 0 {
				return
			}

			cmd := exec.Command(args[0], args[1:]...)
			if w.Cwd != "" {
				cmd.Dir = w.Cwd
			}

			// detach: we don't need stdout/stderr and don't wait for completion
			_ = cmd.Start()
		}(w)
	}

	// allow goroutines to be scheduled; we don't strictly need to wait,
	// but waiting briefly avoids exiting before Start() is called
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
}

// splitCommandLine is a simple, conservative splitter: it splits on spaces and
// ignores quoting/escaping. This is good enough for many GUI apps (e.g. "alacritty",
// "firefox", "code /path"), but can be improved in future versions.
//
//	TODO: improve this
func splitCommandLine(cmd string) []string {
	var out []string
	cur := ""
	for _, r := range cmd {
		if r == ' ' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
