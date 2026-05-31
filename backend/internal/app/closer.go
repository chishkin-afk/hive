package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type dep struct {
	name string
	fc   func(ctx context.Context) error
}

type closer struct {
	mu   sync.Mutex
	deps []dep
}

var globalCloser closer

func init() {
	globalCloser = closer{
		mu:   sync.Mutex{},
		deps: make([]dep, 0),
	}
}

func Add(dp dep) {
	globalCloser.mu.Lock()
	defer globalCloser.mu.Unlock()

	globalCloser.deps = append(globalCloser.deps, dp)
}

func Close(ctx context.Context) error {
	var errs error
	for _, dp := range globalCloser.deps {
		if err := dp.fc(ctx); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to close %s: %w", dp.name, err))
		}
	}

	return errs
}
