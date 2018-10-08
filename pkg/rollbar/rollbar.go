package rollbar

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"

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

var configured = false

// App stores informations
type App struct{}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	token := strings.TrimSpace(*config[`token`])

	if token == `` {
		LogWarning(`no token provided`)
		return &App{}
	}

	rollbar.SetToken(token)
	rollbar.SetEnvironment(strings.TrimSpace(*config[`env`]))
	rollbar.SetServerRoot(strings.TrimSpace(*config[`root`]))

	log.Print(fmt.Sprintf(`configuration for %s`, rollbar.Environment()))
	rollbar.Info(fmt.Sprintf(`%s start`, rollbar.Environment()))

	configured = true

	return &App{}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`token`: flag.String(tools.ToCamel(fmt.Sprintf(`%sToken`, prefix)), ``, `[rollbar] Token`),
		`env`:   flag.String(tools.ToCamel(fmt.Sprintf(`%sEnv`, prefix)), `prod`, `[rollbar] Environment`),
		`root`:  flag.String(tools.ToCamel(fmt.Sprintf(`%sServerRoot`, prefix)), ``, `[rollbar] Server Root`),
	}
}

// GetCaller of function (use start 1)
func GetCaller(start, depth int) string {
	pc := make([]uintptr, depth)
	n := runtime.Callers(start, pc)
	if n == 0 {
		return ``
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)
	stacktraces := make([]string, 0)

	for {
		frame, more := frames.Next()
		if strings.Contains(frame.File, `runtime/`) {
			break
		}

		stacktraces = append(stacktraces, fmt.Sprintf(`%s:%d`, frame.Function, frame.Line))
		if !more {
			break
		}
	}

	if len(stacktraces) == 0 {
		return ``
	}

	if len(stacktraces) == 1 {
		return fmt.Sprintf("\nfrom %s", stacktraces[0])
	}

	return fmt.Sprintf("\nfrom\n\t- %s", strings.Join(stacktraces, "\n\t- "))
}

// Warning send warning message to rollbar
func Warning(interfaces ...interface{}) {
	if configured {
		rollbar.Warning(interfaces...)
	}
}

// Error send error message to rollbar
func Error(interfaces ...interface{}) {
	if configured {
		rollbar.Error(interfaces...)
	}
}

// LogWarning send warning to rollbar and to standard log
func LogWarning(format string, a ...interface{}) {
	content := fmt.Sprintf(format, a...)

	log.Print(content, GetCaller(3, 1))
	Warning(content)
}

// LogError send error to rollbar and to standard log
func LogError(format string, a ...interface{}) {
	err := fmt.Errorf(format, a...)

	log.Print(err, GetCaller(3, 5))
	Error(err)
}

// Deploy send deploy informations to rollbar
func Deploy(ctx context.Context, token, environment, revision, user string) error {
	payload, err := request.PostForm(ctx, deployEndpoint, url.Values{
		`access_token`:   {token},
		`environment`:    {environment},
		`revision`:       {revision},
		`local_username`: {user},
	}, nil)

	if err != nil {
		return fmt.Errorf(`error while posting form: %v. %s`, err, payload)
	}
	return nil
}

func (a App) check() bool {
	return configured
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
