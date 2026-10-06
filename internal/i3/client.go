package i3

import (
	"encoding/json"
	"fmt"

	"go.i3wm.org/i3"
)

// PrintTree prints the current layout tree of the focused workspace in i3.
func PrintTree() error {
	tree, err := GetTree()
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal i3 tree: %w", err)
	}

	fmt.Println(string(out))
	return nil
}

// GetTree retrieves the current layout tree from i3 and is exported for internal packages.
func GetTree() (i3.Tree, error) {
	tree, err := i3.GetTree()
	if err != nil {
		return i3.Tree{}, fmt.Errorf("failed to get i3 tree: %w", err)
	}
	return tree, nil
}
