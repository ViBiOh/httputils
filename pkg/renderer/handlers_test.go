package renderer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/stretchr/testify/assert"
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
			"<a href=\"/?msgKey=Message&amp;msgCnt=Created+with+success&amp;msgLvl=success\">Found</a>.\n\n",
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
			"<a href=\"/success?refresh=true&amp;msgKey=Message&amp;msgCnt=Created+with+success&amp;msgLvl=success\">Found</a>.\n\n",
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
			"<a href=\"/app/success?msgKey=Message&amp;msgCnt=Created+with+success&amp;msgLvl=success\">Found</a>.\n\n",
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
			"<a href=\"/success?msgKey=Message&amp;msgCnt=Created+with+success&amp;msgLvl=success#id\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{fmt.Sprintf("/success?%s#id", NewSuccessMessage("Created with success"))},
			},
		},
		"anchor query and custom": {
			Service{},
			httptest.NewRequest(http.MethodGet, "http://localhost:1080/", nil),
			"/success?refresh=true#id",
			NewKeyErrorMessage("ModalMessage", "Created with success"),
			"<a href=\"/success?refresh=true&amp;msgKey=ModalMessage&amp;msgCnt=Created+with+success&amp;msgLvl=error#id\">Found</a>.\n\n",
			http.StatusFound,
			http.Header{
				"Location": []string{fmt.Sprintf("/success?refresh=true&%s#id", NewKeyErrorMessage("ModalMessage", "Created with success"))},
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			testCase.instance.Redirect(writer, testCase.request, testCase.path, testCase.message)

			assert.Equal(t, testCase.wantStatus, writer.Code)

			actual, _ := request.ReadBodyResponse(writer.Result())
			assert.Equal(t, testCase.want, string(actual))

			for key := range testCase.wantHeader {
				assert.Equal(t, testCase.wantHeader.Get(key), writer.Header().Get(key))
			}
		})
	}
}
