# i3-snapshot

> NOTE: this project is still HEAVILY under development, usage is limited to myself

A "save state" button for the i3 window manager. 

**_i3-snapshot_**'s primary purpose is to solve the long-standing difficulty of saving and restoring workspace layouts. Unlike the built-in "i3-save-tree" tool, which produces incomplete JSON requiring manual editing, **_i3-snapshot_** aims to be a "zero-config" solution.

## Installation

### From source

```bash
git clone https://github.com/a9sk/i3-snapshot.git
cd i3-snapshot
make build
sudo cp i3-snapshot /usr/local/bin/
```

### Requirements

- i3wm
- Go 1.21+ (if building from source)

## Usage

### Save a snapshot

```bash
i3-snapshot save [name]
```

This saves all workspaces to `~/.config/i3-snapshot/saves/[name].json`.

### Restore a snapshot

```bash
i3-snapshot restore [name]
```

This will:
1. Switch to each saved workspace
2. Apply the saved layout
3. Launch all applications
4. Wait for windows to appear and get swallowed by placeholders (in future versions)
5. Clean up unused placeholders (also in future versions)

### Other commands

```bash
i3-snapshot tree        # print the current i3 tree (debug)
i3-snapshot pid <pid>   # show command for a PID (debug)
i3-snapshot version     # show version information
```

## How it works

1. **Save**: Connects to i3 IPC, walks the tree, and for each window:
   - Records window properties (class, instance, title)
   - Uses X11 `_NET_WM_PID` to get the process ID
   - Reads `/proc/[PID]/cmdline` and `/proc/[PID]/cwd` for execution details

2. **Restore**: 
   - Reads the snapshot JSON
   - For each workspace: switches to it, applies layout via `append_layout`
   - Launches commands and waits for windows to appear
   - Automatically corrects windows that appear in wrong workspaces
   - Removes placeholder windows that weren't swallowed (in future versions)

## Limitations

- Command line parsing is naive (splits on spaces, no quoting support)
- Some windows may not have `_NET_WM_PID` set (will have empty command/cwd)
- Terminals only record the terminal process, not what runs inside (e.g., neovim sessions)
- Placeholder cleanup is in development and does not work right now (manual closing needed)

## References:
- https://pkg.go.dev/go.i3wm.org/i3
- https://cobra.dev/docs/
- https://github.com/BurntSushi/xgbutil
- https://github.com/BurntSushi/xgb

## License

MIT License - see [LICENSE](LICENSE) file.