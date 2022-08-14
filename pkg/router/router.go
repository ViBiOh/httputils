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
	wildcardPrefix = '*'
)

var contextKey ctxKey

func getURLPart(url string) (string, string) {
	urlPart := url
	if index := strings.Index(url, pathSeparator); index > 0 {
		urlPart = url[:index]
		url = url[index+1:]
	}

	return url, urlPart
}

// Router with path management
type Router struct {
	defaultHandler http.Handler
	root           node
}

// NewRouter creates a new empty Router
func NewRouter() Router {
	return Router{
		defaultHandler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
		root: node{
			value: make(map[string]http.Handler),
		},
	}
}

// DefaultHandler sets default handler when no route is not found
func (r Router) DefaultHandler(handler http.Handler) Router {
	r.defaultHandler = handler
	return r
}

// Get configure route for GET method
func (r Router) Get(pattern string, handler http.Handler) Router {
	return r.AddRoute(http.MethodGet, pattern, handler)
}

// Head configure route for HEAD method
func (r Router) Head(pattern string, handler http.Handler) Router {
	return r.AddRoute(http.MethodHead, pattern, handler)
}

// Post configure route for POST method
func (r Router) Post(pattern string, handler http.Handler) Router {
	return r.AddRoute(http.MethodPost, pattern, handler)
}

// Put configure route for PUT method
func (r Router) Put(pattern string, handler http.Handler) Router {
	return r.AddRoute(http.MethodPut, pattern, handler)
}

// Patch configure route for PATCH method
func (r Router) Patch(pattern string, handler http.Handler) Router {
	return r.AddRoute(http.MethodPatch, pattern, handler)
}

// Delete configure route for DELETE method
func (r Router) Delete(pattern string, handler http.Handler) Router {
	return r.AddRoute(http.MethodDelete, pattern, handler)
}

// Connect configure route for CONNECT method
func (r Router) Connect(pattern string, handler http.Handler) Router {
	return r.AddRoute(http.MethodConnect, pattern, handler)
}

// Options configure route for OPTIONS method
func (r Router) Options(pattern string, handler http.Handler) Router {
	return r.AddRoute(http.MethodOptions, pattern, handler)
}

// Trace configure route for TRACE method
func (r Router) Trace(pattern string, handler http.Handler) Router {
	return r.AddRoute(http.MethodTrace, pattern, handler)
}

// Any configure route for any method in the specs
func (r Router) Any(pattern string, handler http.Handler) Router {
	return r.
		Get(pattern, handler).
		Head(pattern, handler).
		Post(pattern, handler).
		Put(pattern, handler).
		Patch(pattern, handler).
		Delete(pattern, handler).
		Connect(pattern, handler).
		Options(pattern, handler).
		Trace(pattern, handler)
}

// AddRoute for given method and pattern. Pattern must startss with a slash, should not contain trailing slash.
// Path variable must be prefixed with ':', next to the slash separator
// Glob variable must be prefixed with '*', next to the slash separator, at the end of the pattern
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

	var err error
	r.root, err = r.root.insert(method, pattern[1:], false, handler)

	if err != nil {
		panic(err)
	}

	return r
}

func sanitizeURL(req *http.Request) string {
	url := req.URL.Path

	if strings.HasPrefix(url, pathSeparator) {
		url = url[1:]
	}
	if strings.HasSuffix(url, pathSeparator) {
		url = url[:len(url)-1]
	}

	return url
}

// Handler for request. Should be use with net/http
func (r Router) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handler, hasVariable := r.root.find(req.Method, sanitizeURL(req))
		if handler == nil {
			r.defaultHandler.ServeHTTP(w, req)
			return
		}

		if hasVariable {
			req = req.WithContext(context.WithValue(req.Context(), contextKey, &r.root))
		}
		handler.ServeHTTP(w, req)
	})
}

// GetParams of a request
func GetParams(r *http.Request) map[string]string {
	switch value := r.Context().Value(contextKey).(type) {
	case *node:
		params := make(map[string]string)
		value.extractVariable(r.Method, sanitizeURL(r), params)
		return params
	default:
		return nil
	}
}
