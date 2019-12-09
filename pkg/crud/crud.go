package crud

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/httpjson"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/query"
	"github.com/ViBiOh/httputils/v3/pkg/request"
)

var (
	// ErrServiceIsRequired occurs when underlying service is not provided
	ErrServiceIsRequired = errors.New("service is required")

	// ErrNotFound occurs when item is not found
	ErrNotFound = errors.New("not found")

	// ErrInvalid occurs when invalid action is requested
	ErrInvalid = errors.New("invalid")

	// ErrUnauthorized occurs when authentication not provided
	ErrUnauthorized = errors.New("authentication required")

	// ErrForbidden occurs when action is forbidden
	ErrForbidden = errors.New("forbidden")

	// ErrInternal occurs when unhandled behavior occurs
	ErrInternal = errors.New("internal server error")

	reservedQueryParams = []string{"page", "pageSize", "sort", "desc"}
)

// App of package
type App interface {
	Handler() http.Handler
}

// Config of package
type Config struct {
	defaultPage     *uint
	defaultPageSize *uint
	maxPageSize     *uint
}

type app struct {
	defaultPage     uint
	defaultPageSize uint
	maxPageSize     uint
	service         Service
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		defaultPage:     flags.New(prefix, "crud").Name("DefaultPage").Default(1).Label("Default page").ToUint(fs),
		defaultPageSize: flags.New(prefix, "crud").Name("DefaultPageSize").Default(20).Label("Default page size").ToUint(fs),
		maxPageSize:     flags.New(prefix, "crud").Name("MaxPageSize").Default(100).Label("Max page size").ToUint(fs),
	}
}

// New creates new App from Config
func New(config Config, service Service) (App, error) {
	if service == nil {
		return nil, ErrServiceIsRequired
	}

	return &app{
		defaultPage:     *config.defaultPage,
		defaultPageSize: *config.defaultPageSize,
		maxPageSize:     *config.maxPageSize,

		service: service,
	}, nil
}

// Handler for CRUD requests. Should be use with net/http
func (a app) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isRoot := query.IsRoot(r)

		switch r.Method {
		case http.MethodGet:
			if isRoot {
				a.list(w, r)
			} else if id, err := query.GetUintID(r); err != nil {
				httperror.BadRequest(w, err)
			} else {
				a.get(w, r, id)
			}

		case http.MethodPost:
			if isRoot {
				a.create(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}

		case http.MethodPut:
			if isRoot {
				w.WriteHeader(http.StatusMethodNotAllowed)
			} else if id, err := query.GetUintID(r); err != nil {
				httperror.BadRequest(w, err)
			} else {
				a.update(w, r, id)
			}

		case http.MethodDelete:
			if isRoot {
				w.WriteHeader(http.StatusMethodNotAllowed)
			} else if id, err := query.GetUintID(r); err != nil {
				httperror.BadRequest(w, err)
			} else {
				a.delete(w, r, id)
			}

		case http.MethodOptions:
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func handleError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrInvalid) {
		httperror.BadRequest(w, err)
	} else if errors.Is(err, ErrUnauthorized) {
		httperror.Unauthorized(w, err)
	} else if errors.Is(err, ErrForbidden) {
		httperror.Forbidden(w)
	} else if errors.Is(err, ErrNotFound) {
		httperror.NotFound(w)
	} else {
		logger.Error(err.Error())
		httperror.InternalServerError(w, ErrInternal)
	}

	return true
}

func writeErrors(w http.ResponseWriter, errors []error) {
	output := strings.Builder{}
	output.WriteString("invalid payload:")

	for _, err := range errors {
		output.WriteString("\n\t")
		output.WriteString(err.Error())
	}

	httperror.BadRequest(w, fmt.Errorf(output.String()))
}

func readFilters(r *http.Request) map[string][]string {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		logger.Warn(err.Error())
		return nil
	}

	for _, reservedParam := range reservedQueryParams {
		delete(params, reservedParam)
	}

	return params
}

func (a app) readPayload(r *http.Request) (Item, error) {
	bodyBytes, err := request.ReadBodyRequest(r)
	if err != nil {
		return nil, fmt.Errorf("body read error: %w", err)
	}

	item, err := a.service.Unmarsall(bodyBytes)
	if err != nil {
		return item, fmt.Errorf("unmarshall error: %w", err)
	}

	return item, nil
}

func (a app) list(w http.ResponseWriter, r *http.Request) {
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

	httpjson.ResponsePaginatedJSON(w, http.StatusOK, params.Page, params.PageSize, total, list, httpjson.IsPretty(r))
}

func (a app) get(w http.ResponseWriter, r *http.Request, id uint64) {
	obj, err := a.service.Get(r.Context(), id)
	if handleError(w, err) {
		return
	}

	httpjson.ResponseJSON(w, http.StatusOK, obj, httpjson.IsPretty(r))
}

func (a app) create(w http.ResponseWriter, r *http.Request) {
	obj, err := a.readPayload(r)
	if err != nil {
		httperror.BadRequest(w, err)
		return
	}

	id := uint64(0)
	obj.SetID(id)

	if errors := a.service.Check(nil, obj); len(errors) != 0 {
		writeErrors(w, errors)
		return
	}

	obj, id, err = a.service.Create(r.Context(), obj)
	if handleError(w, err) {
		return
	}
	obj.SetID(id)

	httpjson.ResponseJSON(w, http.StatusCreated, obj, httpjson.IsPretty(r))
}

func (a app) update(w http.ResponseWriter, r *http.Request, id uint64) {
	new, err := a.readPayload(r)
	if err != nil {
		httperror.BadRequest(w, err)
		return
	}

	ctx := r.Context()

	old, err := a.service.Get(ctx, id)
	if err != nil {
		handleError(w, err)
		return
	}

	new.SetID(id)
	if errors := a.service.Check(old, new); len(errors) != 0 {
		writeErrors(w, errors)
		return
	}

	new, err = a.service.Update(ctx, new)
	if handleError(w, err) {
		return
	}

	httpjson.ResponseJSON(w, http.StatusOK, new, httpjson.IsPretty(r))
}

func (a app) delete(w http.ResponseWriter, r *http.Request, id uint64) {
	ctx := r.Context()

	obj, err := a.service.Get(ctx, id)
	if err != nil {
		handleError(w, err)
		return
	}

	if errors := a.service.Check(obj, nil); len(errors) != 0 {
		writeErrors(w, errors)
		return
	}

	err = a.service.Delete(ctx, obj)
	if handleError(w, err) {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
