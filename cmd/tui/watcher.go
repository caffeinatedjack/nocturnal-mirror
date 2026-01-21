package tui

import (
	"time"

	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// Watcher monitors the spec directory for file changes.
type Watcher struct {
	watcher    *fsnotify.Watcher
	specPath   string
	debounceMs time.Duration
	lastChange time.Time
}

// NewWatcher creates a new file system watcher.
func NewWatcher(specPath string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		watcher:    fsWatcher,
		specPath:   specPath,
		debounceMs: 500 * time.Millisecond,
	}, nil
}

// Start returns a tea.Cmd that begins watching for file changes.
func (w *Watcher) Start() bubbletea.Cmd {
	// Add directories to watch
	dirs := []string{
		"proposal",
		"rule",
		"maintenance",
		"section",
		"third",
	}

	for _, dir := range dirs {
		path := w.specPath + "/" + dir
		if err := w.watcher.Add(path); err != nil {
			// Directory might not exist, that's OK
			continue
		}
	}

	// Watch state file and config file
	if err := w.watcher.Add(w.specPath + "/.nocturnal.json"); err == nil {
		// ignore error
	}
	if err := w.watcher.Add(w.specPath + "/nocturnal.yaml"); err == nil {
		// ignore error
	}

	return func() bubbletea.Msg {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}
			// Filter for write and create events
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				// Debounce
				now := time.Now()
				if now.Sub(w.lastChange) > w.debounceMs {
					w.lastChange = now
					return RefreshMsg{}
				}
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			return ErrorMsg{Err: err}
		}
		return nil
	}
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	if w.watcher != nil {
		return w.watcher.Close()
	}
	return nil
}
