package i3

import (
	"encoding/json"
	"fmt"

	"go.i3wm.org/i3"
)

// Connect verifies that we can talk to the i3 IPC socket.
func Connect() {
	_ = GetTree()
}

// PrintTree prints the current layout tree of the focused workspace in i3.
func PrintTree() {
	tree := GetTree()

	out, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		fmt.Printf("failed to marshal i3 tree: %v\n", err)
		return
	}

	fmt.Println(string(out))
}

// GetTree retrieves the current layout tree from i3 and is exported for internal packages.
func GetTree() i3.Tree {
	tree, err := i3.GetTree()
	if err != nil {
		fmt.Printf("failed to get i3 tree: %v\n", err)
		return i3.Tree{}
	}
	return tree
}
