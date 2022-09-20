package main

import (
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

type port struct {
	template renderer.TemplateFunc
}

func newPort(config configuration, client client, adapter adapter) port {
	var output port

	output.template = func(w http.ResponseWriter, r *http.Request) (renderer.Page, error) {
		resp, err := request.Get("https://api.vibioh.fr/dump/").Send(r.Context(), nil)
		if err != nil {
			return renderer.Page{}, err
		}

		if err = request.DiscardBody(resp.Body); err != nil {
			return renderer.Page{}, err
		}

		return renderer.NewPage("public", http.StatusOK, nil), nil
	}

	return output
}
