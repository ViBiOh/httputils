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

// App stores informations
type App struct {
	token string
}

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

	return &App{
		token: token,
	}
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
	return a.token != ``
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
