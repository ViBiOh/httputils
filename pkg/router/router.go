package router

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey struct{}

const (
	pathSeparator  = "/"
	variablePrefix = ':'
)

var contextKey ctxKey

type route struct {
	handler     http.Handler
	parts       []string
	hasVariable bool
}

// Router with path management
type Router struct {
	routes map[string][][]route // one entry for each method, and one array for each size of slash, and finally an array for possibilities
}

// NewRouter creates a new empty Router
func NewRouter() Router {
	return Router{
		routes: make(map[string][][]route, 0),
	}
}

// AddRoute for given method and pattern. Pattern must starts with a slash, should not contains trailing slash and path variable must be prefixed with ':'
func (r Router) AddRoute(method, pattern string, handler http.Handler) Router {
	if len(method) == 0 {
		panic("method is required")
	}

	if handler == nil {
		panic("handler is required")
	}

	if !strings.HasPrefix(pattern, pathSeparator) {
		panic("pattern has to start with a slash")
	}

	if len(pattern) != 1 && strings.HasSuffix(pattern, pathSeparator) {
		panic("pattern can't end with a slash")
	}

	if r.routes == nil {
		r.routes = make(map[string][][]route)
	}

	parts := strings.Split(pattern[1:], pathSeparator)
	index := len(parts) - 1

	if len(r.routes[method]) < len(parts) {
		newRoutes := make([][]route, len(parts))
		copy(newRoutes, r.routes[method])
		r.routes[method] = newRoutes
	}

	var hasVariable bool
	for _, part := range parts {
		if len(part) > 0 && part[0] == variablePrefix {
			hasVariable = true
			break
		}
	}

	r.routes[method][index] = append(r.routes[method][index], route{
		parts:       parts,
		handler:     handler,
		hasVariable: hasVariable,
	})

	return r
}

// Handler for request. Should be use with net/http
func (r Router) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		routes, ok := r.routes[req.Method]
		if !ok {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		url := req.URL.Path
		if strings.HasPrefix(url, pathSeparator) {
			url = url[1:]
		}
		if strings.HasSuffix(url, pathSeparator) {
			url = url[:len(url)-1]
		}

		size := strings.Count(url, pathSeparator)
		if len(routes) < size {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var index int

		for _, route := range routes[size] {
			for _, part := range route.parts {
				urlPart := url
				if index = strings.Index(url, pathSeparator); index > 0 {
					urlPart = url[:index]
					url = url[index+1:]
				}

				if part == urlPart || len(part) > 0 && part[0] == variablePrefix {
					continue
				}

				goto checkAnotherPart
			}

			if route.hasVariable {
				route.handler.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), contextKey, route)))
			}
			route.handler.ServeHTTP(w, req)
		checkAnotherPart:
		}

		w.WriteHeader(http.StatusNotFound)
	})
}
