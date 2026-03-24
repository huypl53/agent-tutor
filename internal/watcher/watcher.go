package watcher

import "context"

// Watcher is the common interface for all watchers (file, terminal, git, etc.).
type Watcher interface {
	Start(ctx context.Context) error
	Stop() error
}
