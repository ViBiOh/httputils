package crud

import (
	native_errors "errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ViBiOh/httputils/pkg/httperror"
	"github.com/ViBiOh/httputils/pkg/httpjson"
	"github.com/ViBiOh/httputils/pkg/pagination"
	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/tools"
)

var (
	// ErrNotFound occurs when item with given ID if not found
	ErrNotFound = native_errors.New(`item not found`)
)

// App stores informations
type App struct {
	path            string
	defaultPage     uint
	defaultPageSize uint
	maxPageSize     uint
	service         ItemService
}

// NewApp creates new App from Flags' config
func NewApp(config map[string]interface{}, service ItemService) *App {
	return &App{
		path:            *config[`path`].(*string),
		defaultPage:     *config[`defaultPage`].(*uint),
		defaultPageSize: *config[`defaultPageSize`].(*uint),
		maxPageSize:     *config[`maxPageSize`].(*uint),
		service:         service,
	}
}

// Flags adds flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`path`:            flag.String(tools.ToCamel(fmt.Sprintf(`%sPath`, prefix)), fmt.Sprintf(`/%s`, prefix), fmt.Sprintf(`[%s] HTTP Path prefix`, prefix)),
		`defaultPage`:     flag.Uint(tools.ToCamel(fmt.Sprintf(`%sDefaultPage`, prefix)), 1, fmt.Sprintf(`[%s] Default page`, prefix)),
		`defaultPageSize`: flag.Uint(tools.ToCamel(fmt.Sprintf(`%sDefaultPageSize`, prefix)), 20, fmt.Sprintf(`[%s] Default page size`, prefix)),
		`maxPageSize`:     flag.Uint(tools.ToCamel(fmt.Sprintf(`%sMaxPageSize`, prefix)), 500, fmt.Sprintf(`[%s] Max page size`, prefix)),
	}
}

func handleError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	if err == ErrNotFound {
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
	page, pageSize, sortKey, sortAsc, err := pagination.ParseParams(r, a.defaultPage, a.defaultPageSize, a.maxPageSize)
	if err != nil {
		httperror.BadRequest(w, err)
		return
	}

	list, total, err := a.service.List(r.Context(), page, pageSize, sortKey, sortAsc, readFilters(r))
	if err != nil {
		httperror.InternalServerError(w, err)
		return
	}

	if len(list) == 0 && total > 0 {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	if err := httpjson.ResponsePaginatedJSON(w, http.StatusOK, page, pageSize, total, list, httpjson.IsPretty(r)); err != nil {
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
	return http.StripPrefix(a.path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
}
