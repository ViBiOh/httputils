package redistest

import (
	"context"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/redis"
)

var _ redis.App = &App{}

// App mocks
type App struct {
	pingErr error

	storeErr error

	loadContent string
	loadErr     error

	deleteErr error

	exclusiveErr error
}

// New create new instances
func New() *App {
	return &App{}
}

// SetPing mocks
func (a *App) SetPing(err error) *App {
	a.pingErr = err

	return a
}

// SetStore mocks
func (a *App) SetStore(err error) *App {
	a.storeErr = err

	return a
}

// SetLoad mocks
func (a *App) SetLoad(content string, err error) *App {
	a.loadContent = content
	a.loadErr = err

	return a
}

// SetDelete mocks
func (a *App) SetDelete(err error) *App {
	a.deleteErr = err

	return a
}

// SetExclusive mocks
func (a *App) SetExclusive(err error) *App {
	a.exclusiveErr = err

	return a
}

// Ping mocks
func (a *App) Ping() error {
	return a.pingErr
}

// Store mocks
func (a *App) Store(context.Context, string, string, time.Duration) error {
	return a.storeErr
}

// Load mocks
func (a *App) Load(context.Context, string) (string, error) {
	return a.loadContent, a.loadErr
}

// Delete mocks
func (a *App) Delete(context.Context, string) error {
	return a.deleteErr
}

// Exclusive mocks
func (a *App) Exclusive(ctx context.Context, _ string, _ time.Duration, action func(context.Context) error) error {
	return action(ctx)
}
