package rollbar

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/pkg/tools"
	rollbar "github.com/rollbar/rollbar-go"
)

var configured = false

// App stores informations
type App struct{}

// NewApp creates new App from Flags' config
func NewApp(config map[string]*string) *App {
	token := strings.TrimSpace(*config[`token`])

	if token == `` {
		log.Print(`[rollbar] No token provided`)
		return &App{}
	}

	rollbar.SetToken(token)
	rollbar.SetEnvironment(strings.TrimSpace(*config[`env`]))
	rollbar.SetServerRoot(strings.TrimSpace(*config[`root`]))

	log.Print(fmt.Sprintf(`[rollbar] Configuration for %s`, rollbar.Environment()))
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
