package renderer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func TestRedirect(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		instance   Service
		request    *http.Request
		path       string
		message    Message
		want       string
		wantStatus int
		wantHeader http.Header
	}{
		"simple": {
			Service{},
			httptest.NewRequest(http.MethodGet, "https://vibioh.fr/", nil),
			"/",
			NewSuccessMessage("Created with success"),
			"<a href=\"/?messageContent=Created+with+success&amp;messageLevel=success\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{fmt.Sprintf("/?%s", NewSuccessMessage("Created with success"))},
			},
		},
		"relative URL": {
			Service{},
			httptest.NewRequest(http.MethodGet, "http://localhost:1080/", nil),
			"/success?refresh=true",
			NewSuccessMessage("Created with success"),
			"<a href=\"/success?refresh=true&amp;messageContent=Created+with+success&amp;messageLevel=success\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{fmt.Sprintf("/success?refresh=true&%s", NewSuccessMessage("Created with success"))},
			},
		},
		"path prefix": {
			Service{
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
		"anchor": {
			Service{},
			httptest.NewRequest(http.MethodGet, "http://localhost:1080/", nil),
			"/success#id",
			NewSuccessMessage("Created with success"),
			"<a href=\"/success?messageContent=Created+with+success&amp;messageLevel=success#id\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{fmt.Sprintf("/success?%s#id", NewSuccessMessage("Created with success"))},
			},
		},
		"anchor and query": {
			Service{},
			httptest.NewRequest(http.MethodGet, "http://localhost:1080/", nil),
			"/success?refresh=true#id",
			NewSuccessMessage("Created with success"),
			"<a href=\"/success?refresh=true&amp;messageContent=Created+with+success&amp;messageLevel=success#id\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{fmt.Sprintf("/success?refresh=true&%s#id", NewSuccessMessage("Created with success"))},
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			testCase.instance.Redirect(writer, testCase.request, testCase.path, testCase.message)

			if got := writer.Code; got != testCase.wantStatus {
				t.Errorf("Redirect = %d, want %d", got, testCase.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != testCase.want {
				t.Errorf("Redirect = `%s`, want `%s`", string(got), testCase.want)
			}

			for key := range testCase.wantHeader {
				want := testCase.wantHeader.Get(key)
				if got := writer.Header().Get(key); got != want {
					t.Errorf("`%s` Header = `%s`, want `%s`", key, got, want)
				}
			}
		})
	}
}
