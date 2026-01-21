package tui

import (
	"fmt"
	"os"
	"os/exec"

	bubbletea "github.com/charmbracelet/bubbletea"
)

// Editor handles opening external editors for file editing.
type Editor struct {
	editor string
	path   string
}

// NewEditor creates a new editor instance for the given file.
func NewEditor(path string) *Editor {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Try common editors
		editors := []string{"vim", "nvim", "vi", "nano", "code", "code --wait"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		editor = "vim" // Final fallback
	}

	return &Editor{
		editor: editor,
		path:   path,
	}
}

// RunCmd returns a bubbletea.Cmd that opens the editor.
func (e *Editor) RunCmd() bubbletea.Cmd {
	c := exec.Command(e.editor, e.path)
	return bubbletea.ExecProcess(c, func(err error) bubbletea.Msg {
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("editor error: %w", err)}
		}
		return EditorDoneMsg{Path: e.path}
	})
}

// OpenEditor is a convenience function to open a file in an editor.
func OpenEditor(path string) bubbletea.Cmd {
	return NewEditor(path).RunCmd()
}
