package crud

import (
	"errors"
	"flag"
	"net/http"
	"net/url"

	"github.com/ViBiOh/httputils/v2/pkg/httperror"
	"github.com/ViBiOh/httputils/v2/pkg/httpjson"
	"github.com/ViBiOh/httputils/v2/pkg/query"
	"github.com/ViBiOh/httputils/v2/pkg/request"
	"github.com/ViBiOh/httputils/v2/pkg/tools"
)

var (
	// ErrNotFound occurs when item with given ID if not found
	ErrNotFound = errors.New("item not found")

	// ErrInvalid occurs when invalid action is requested
	ErrInvalid = errors.New("invalid")
)

// Config of package
type Config struct {
	defaultPage     *uint
	defaultPageSize *uint
	maxPageSize     *uint
}

// App of package
type App struct {
	defaultPage     uint
	defaultPageSize uint
	maxPageSize     uint
	service         ItemService
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		defaultPage:     tools.NewFlag(prefix, "crud").Name("DefaultPage").Default(1).Label("Default page").ToUint(fs),
		defaultPageSize: tools.NewFlag(prefix, "crud").Name("DefaultPageSize").Default(20).Label("Default page size").ToUint(fs),
		maxPageSize:     tools.NewFlag(prefix, "crud").Name("MaxPageSize").Default(500).Label("Max page size").ToUint(fs),
	}
}

// New creates new App from Config
func New(config Config, service ItemService) *App {
	return &App{
		defaultPage:     *config.defaultPage,
		defaultPageSize: *config.defaultPageSize,
		maxPageSize:     *config.maxPageSize,
		service:         service,
	}
}

func handleError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrInvalid) {
		httperror.BadRequest(w, err)
	} else if errors.Is(err, ErrNotFound) {
		httperror.NotFound(w)
	} else {
		httperror.InternalServerError(w, err)
	}

	return true
}

func (a App) readPayload(r *http.Request) (Item, error) {
	bodyBytes, err := request.ReadBodyRequest(r)
	if err != nil {
		return nil, err
	}

	return a.service.Unmarsall(bodyBytes)
}

func readFilters(r *http.Request) map[string][]string {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil
	}

	return params
}

func (a App) list(w http.ResponseWriter, r *http.Request) {
	params, err := query.ParsePagination(r, a.defaultPage, a.defaultPageSize, a.maxPageSize)
	if err != nil {
		httperror.BadRequest(w, err)
		return
	}

	list, total, err := a.service.List(r.Context(), params.Page, params.PageSize, params.Sort, params.Desc, readFilters(r))
	if handleError(w, err) {
		return
	}

	if len(list) == 0 && total > 0 {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	if err := httpjson.ResponsePaginatedJSON(w, http.StatusOK, params.Page, params.PageSize, total, list, httpjson.IsPretty(r)); err != nil {
		httperror.InternalServerError(w, err)
		return
	}
}

func (a App) get(w http.ResponseWriter, r *http.Request, id string) {
	obj, err := a.service.Get(r.Context(), id)
	if handleError(w, err) {
		return
	}

	if err := httpjson.ResponseJSON(w, http.StatusOK, obj, httpjson.IsPretty(r)); err != nil {
		httperror.InternalServerError(w, err)
		return
	}
}

func (a App) create(w http.ResponseWriter, r *http.Request) {
	obj, err := a.readPayload(r)

	if err != nil {
		httperror.BadRequest(w, err)
		return
	}

	obj.SetID("")

	obj, err = a.service.Create(r.Context(), obj)
	if handleError(w, err) {
		return
	}

	if err := httpjson.ResponseJSON(w, http.StatusCreated, obj, httpjson.IsPretty(r)); err != nil {
		httperror.InternalServerError(w, err)
		return
	}
}

func (a App) update(w http.ResponseWriter, r *http.Request, id string) {
	obj, err := a.readPayload(r)

	if err != nil {
		httperror.BadRequest(w, err)
		return
	}

	_, err = a.service.Get(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	obj.SetID(id)

	obj, err = a.service.Update(r.Context(), obj)
	if handleError(w, err) {
		return
	}

	if err := httpjson.ResponseJSON(w, http.StatusOK, obj, httpjson.IsPretty(r)); err != nil {
		httperror.InternalServerError(w, err)
		return
	}
}

func (a App) delete(w http.ResponseWriter, r *http.Request, id string) {
	obj, err := a.service.Get(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	err = a.service.Delete(r.Context(), obj)
	if handleError(w, err) {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Handler for CRUD requests. Should be use with net/http
func (a App) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isRoot := tools.IsRoot(r)

		switch r.Method {
		case http.MethodGet:
			if isRoot {
				a.list(w, r)
			} else {
				a.get(w, r, tools.GetID(r))
			}

		case http.MethodPost:
			if isRoot {
				a.create(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}

		case http.MethodPut:
			if !isRoot {
				a.update(w, r, tools.GetID(r))
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}

		case http.MethodDelete:
			if !isRoot {
				a.delete(w, r, tools.GetID(r))
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}

		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
