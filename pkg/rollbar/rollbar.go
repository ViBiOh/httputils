package rollbar

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/model"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/tools"
	rollbar "github.com/rollbar/rollbar-go"
)

const (
	deployEndpoint = `https://api.rollbar.com/api/1/deploy/`
)

var _ model.Middleware = &App{}
var _ model.Flusher = &App{}
var _ logger.LogReporter = &App{}

// App stores informations
type App struct {
	active bool
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	token := strings.TrimSpace(*config[`token`])

	if token == `` {
		logger.Warn(`no token provided`)
		return &App{
			active: false,
		}
	}

	rollbar.SetToken(token)
	rollbar.SetEnvironment(strings.TrimSpace(*config[`env`]))
	rollbar.SetServerRoot(strings.TrimSpace(*config[`root`]))

	logger.Info(`Configuration for %s`, rollbar.Environment())

	app := &App{
		active: true,
	}
	logger.AddReporter(app)

	return app
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`token`: flag.String(tools.ToCamel(fmt.Sprintf(`%sToken`, prefix)), ``, `[rollbar] Token`),
		`env`:   flag.String(tools.ToCamel(fmt.Sprintf(`%sEnv`, prefix)), `prod`, `[rollbar] Environment`),
		`root`:  flag.String(tools.ToCamel(fmt.Sprintf(`%sServerRoot`, prefix)), ``, `[rollbar] Server Root`),
	}
}

func (a App) check() bool {
	return a.active
}

// Info send info message to rollbar
func (a App) Info(content string) {
	if a.check() {
		rollbar.Info(content)
	}
}

// Warn send warning message to rollbar
func (a App) Warn(content string) {
	if a.check() {
		rollbar.Warning(content)
	}
}

// Error send error message to rollbar
func (a App) Error(content string) {
	if a.check() {
		rollbar.Error(content)
	}
}

// Handler for net/http
func (a App) Handler(next http.Handler) http.Handler {
	if !a.check() {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rollbar.Wrap(func() {
			next.ServeHTTP(w, r)
		})
	})
}

// Flush wait for empty queues of message
func (a App) Flush() {
	rollbar.Wait()
}

// Deploy send deploy informations to rollbar
func Deploy(ctx context.Context, token, environment, revision, user string) error {
	_, err := request.PostForm(ctx, deployEndpoint, url.Values{
		`access_token`:   {token},
		`environment`:    {environment},
		`revision`:       {revision},
		`local_username`: {user},
	}, nil)

	if err != nil {
		return err
	}
	return nil
}
