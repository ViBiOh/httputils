package renderer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func TestRedirect(t *testing.T) {
	var cases = []struct {
		intention  string
		instance   app
		request    *http.Request
		path       string
		message    Message
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		{
			"simple",
			app{},
			httptest.NewRequest(http.MethodGet, "http://vibioh.fr/", nil),
			"/",
			NewSuccessMessage("Created with success"),
			"<a href=\"/?messageContent=Created+with+success&amp;messageLevel=success\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{fmt.Sprintf("/?%s", NewSuccessMessage("Created with success"))},
			},
		},
		{
			"relative URL",
			app{},
			httptest.NewRequest(http.MethodGet, "http://localhost:1080/", nil),
			"/success?refresh=true",
			NewSuccessMessage("Created with success"),
			"<a href=\"/success?refresh=true&amp;messageContent=Created+with+success&amp;messageLevel=success\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{fmt.Sprintf("/success?refresh=true&%s", NewSuccessMessage("Created with success"))},
			},
		},
		{
			"path prefix",
			app{
				pathPrefix: "/app",
			},
			httptest.NewRequest(http.MethodGet, "http://localhost:1080/", nil),
			"/success",
			NewSuccessMessage("Created with success"),
			"<a href=\"/app/success?messageContent=Created+with+success&amp;messageLevel=success\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{fmt.Sprintf("/app/success?%s", NewSuccessMessage("Created with success"))},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			tc.instance.Redirect(writer, tc.request, tc.path, tc.message)

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("Redirect = %d, want %d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != tc.want {
				t.Errorf("Redirect = `%s`, want `%s`", string(got), tc.want)
			}

			for key := range tc.wantHeader {
				want := tc.wantHeader.Get(key)
				if got := writer.Header().Get(key); got != want {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, got, want)
				}
			}
		})
	}
}
