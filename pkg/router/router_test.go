package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	cases := map[string]struct {
		router  Router
		request *http.Request
		want    int
	}{
		"not allowed": {
			NewRouter().AddRoute(http.MethodPost, "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})),
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusMethodNotAllowed,
		},
		"root": {
			NewRouter().AddRoute(http.MethodGet, "/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})),
			httptest.NewRequest(http.MethodGet, "/", nil),
			http.StatusNoContent,
		},
		"simple": {
			NewRouter().AddRoute(http.MethodGet, "/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})),
			httptest.NewRequest(http.MethodGet, "/hello", nil),
			http.StatusNoContent,
		},
		"api pattern": {
			NewRouter().AddRoute(http.MethodGet, "/api/users/:userId/items/:itemId", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})),
			httptest.NewRequest(http.MethodGet, "/api/users/1/items/2", nil),
			http.StatusNoContent,
		},
		"trailing slash pattern": {
			NewRouter().AddRoute(http.MethodGet, "/api/users/:userId/items/:itemId", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})),
			httptest.NewRequest(http.MethodGet, "/api/users/1/items/2/", nil),
			http.StatusNoContent,
		},
		"no match": {
			NewRouter().AddRoute(http.MethodGet, "/api/users/:userId/items/:itemId", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})),
			httptest.NewRequest(http.MethodGet, "/api/users/1/items/", nil),
			http.StatusNotFound,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			tc.router.Handler().ServeHTTP(writer, tc.request)

			if got := writer.Code; got != tc.want {
				t.Errorf("Handler = HTTP/%d, want HTTP/%d", got, tc.want)
			}
		})
	}
}

func BenchmarkHandlerNoVariable(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()

	router := NewRouter().
		AddRoute(http.MethodGet, "/api/users", handler).
		AddRoute(http.MethodGet, "/api/users/:userId/items", handler).
		AddRoute(http.MethodPost, "/api/users/:userId/items", handler).
		AddRoute(http.MethodGet, "/api/users/:userId/items/:itemId", handler).
		AddRoute(http.MethodPut, "/api/users/:userId/items/:itemId", handler).
		AddRoute(http.MethodDelete, "/api/users/:userId/items/:itemId", handler).
		Handler()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, r)
	}
}

func BenchmarkHandler(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r := httptest.NewRequest(http.MethodGet, "/api/users/1/items/1", nil)
	w := httptest.NewRecorder()

	router := NewRouter().
		AddRoute(http.MethodGet, "/api/users/:userId/items", handler).
		AddRoute(http.MethodPost, "/api/users/:userId/items", handler).
		AddRoute(http.MethodGet, "/api/users/:userId/items/:itemId", handler).
		AddRoute(http.MethodPut, "/api/users/:userId/items/:itemId", handler).
		AddRoute(http.MethodDelete, "/api/users/:userId/items/:itemId", handler).
		Handler()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, r)
	}
}

func BenchmarkMuxHandler(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPut:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	r := httptest.NewRequest(http.MethodGet, "/api/users/1/items/1", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.Handle("/api/users/1/items", handler)
	mux.Handle("/api/users/1/items/1", handler)

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, r)
	}
}

func BenchmarkStriPrefixHandler(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPut:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	r := httptest.NewRequest(http.MethodGet, "/api/users/1/items/1", nil)
	w := httptest.NewRecorder()

	striped := http.StripPrefix("/api/users", handler)

	for i := 0; i < b.N; i++ {
		striped.ServeHTTP(w, r)
	}
}

func BenchmarkNoopHandler(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r := httptest.NewRequest(http.MethodGet, "/api/users/1/items/1", nil)
	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, r)
	}
}
