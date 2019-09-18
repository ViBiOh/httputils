package templates

import (
	"fmt"
	"html/template"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/v2/pkg/request"
)

func TestWriteHTMLTemplate(t *testing.T) {
	var cases = []struct {
		intention string
		tpl       *template.Template
		want      string
		wantErr   error
	}{
		{
			"simple minify",
			template.Must(template.New("html5_template.html").ParseFiles("html5_template.html")),
			`<!doctype html><html lang=fr><meta charset=utf-8><title>Golang Testing</title><meta name=description content="Golang Testing"><meta name=author content="ViBiOh"><script>
    function helloWorld() {
      console.info('Hello world!');
    }
  </script><style>
    html {
      height: 100vh;
      width: 100vw;
    }
  </style><body onload=helloWorld()><h1>It works!</h1>`,
			nil,
		},
		{
			"handle template error",
			template.Must(template.New("invalidName").ParseFiles("html5_template.html")),
			"",
			fmt.Errorf("template: \"invalidName\" is an incomplete or empty template"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()

			err := WriteHTMLTemplate(testCase.tpl, writer, nil, 200)

			failed := false

			if err == nil && testCase.wantErr != nil {
				failed = true
			} else if err != nil && testCase.wantErr == nil {
				failed = true
			} else if err != nil && err.Error() != testCase.wantErr.Error() {
				failed = true
			}

			if failed {
				t.Errorf("WriteHTMLTemplate() = %#v, want error %#v", err, testCase.wantErr)
			}

			if result, _ := request.ReadBody(writer.Result().Body); string(result) != testCase.want {
				t.Errorf("WriteHTMLTemplate() = `%s`, want `%s`", string(result), testCase.want)
			}
		})
	}
}
