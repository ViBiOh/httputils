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
func NewApp(config map[string]interface{}) *App {
	token := strings.TrimSpace(*(config[`token`].(*string)))

	if token == `` {
		log.Print(`[rollbar] No token provided`)
		return &App{}
	}

	rollbar.SetToken(token)
	rollbar.SetEnvironment(strings.TrimSpace(*(config[`env`].(*string))))
	rollbar.SetServerRoot(strings.TrimSpace(*(config[`root`].(*string))))

	log.Print(fmt.Sprintf(`[rollbar] Configuration for %s`, rollbar.Environment()))
	if *(config[`welcome`].(*bool)) {
		rollbar.Info(`App started`)
	}

	return &App{
		token: token,
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`token`:   flag.String(tools.ToCamel(fmt.Sprintf(`%sToken`, prefix)), ``, `[rollbar] Token`),
		`env`:     flag.String(tools.ToCamel(fmt.Sprintf(`%sEnv`, prefix)), `prod`, `[rollbar] Environment`),
		`root`:    flag.String(tools.ToCamel(fmt.Sprintf(`%sServerRoot`, prefix)), ``, `[rollbar] Server Root`),
		`welcome`: flag.Bool(tools.ToCamel(fmt.Sprintf(`%sWelcome`, prefix)), false, `[rollbar] Send welcome message on start`),
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

// Critical send critical error to rollbar
func (a App) Critical(interfaces ...interface{}) {
	rollbar.Critical(interfaces)
}

// Flush wait for empty queues of message
func (a App) Flush() {
	rollbar.Wait()
}
