package templates

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/stretchr/testify/assert"
)

func TestWriteTemplate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		tpl     *template.Template
		want    string
		wantErr error
	}{
		"simple": {
			template.Must(template.New("css_template.tmpl").ParseFiles("../../templates/css_template.tmpl")),
			"html{height:100vh;width:100vw}",
			nil,
		},
		"error": {
			template.Must(template.New("invalidName").ParseFiles("../../templates/html5_template.tmpl")),
			"",
			fmt.Errorf("template: \"invalidName\" is an incomplete or empty template"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := bytes.NewBuffer(nil)
			err := WriteTemplate(context.Background(), nil, testCase.tpl, writer, nil, "text/css")

			result := writer.String()

			failed := false

			if testCase.wantErr != nil && (err == nil || err.Error() != testCase.wantErr.Error()) {
				failed = true
			} else if result != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("WriteTemplate() = (`%s`, `%s`), want error (`%s`, `%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestResponseHTMLTemplate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		tpl     *template.Template
		want    string
		wantErr error
	}{
		"simple": {
			template.Must(template.New("html5_template.tmpl").ParseFiles("../../templates/html5_template.tmpl")),
			`<!doctype html><html lang=fr><meta charset=utf-8><title>Golang Testing</title><meta name=description content="Golang Testing"><meta name=author content="ViBiOh"><script>function helloWorld(){console.info("Hello world!")}</script><style>html{height:100vh;width:100vw}</style><body onload=helloWorld()><h1>It works!</h1>`,
			nil,
		},
		"error": {
			template.Must(template.New("invalidName").ParseFiles("../../templates/html5_template.tmpl")),
			"",
			fmt.Errorf("template: \"invalidName\" is an incomplete or empty template"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			err := ResponseHTMLTemplate(context.Background(), nil, testCase.tpl, writer, nil, 200)

			result, _ := request.ReadBodyResponse(writer.Result())

			failed := false

			if testCase.wantErr != nil && (err == nil || err.Error() != testCase.wantErr.Error()) {
				failed = true
			} else if string(result) != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("ResponseHTMLTemplate() = (`%s`, `%s`), want error (`%s`, `%s`)", string(result), err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestResponseHTMLTemplateRaw(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		tpl     *template.Template
		want    string
		wantErr error
	}{
		"simple": {
			template.Must(template.New("html5_template.tmpl").ParseFiles("../../templates/html5_template.tmpl")),
			`<!doctype html>

<html lang="fr">
<head>
  <meta charset="utf-8">

  <title>Golang Testing</title>
  <meta name="description" content="Golang Testing">
  <meta name="author" content="ViBiOh">

  <script>
    function helloWorld() {
      console.info('Hello world!');
    }
  </script>

  <style>
    html {
      height: 100vh;
      width: 100vw;
    }
  </style>
</head>

<body onload="helloWorld()">
  <h1>It works!</h1>
</body>
</html>
`,
			nil,
		},
		"error": {
			template.Must(template.New("invalidName").ParseFiles("../../templates/html5_template.tmpl")),
			"",
			fmt.Errorf("template: \"invalidName\" is an incomplete or empty template"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			err := ResponseHTMLTemplateRaw(context.Background(), nil, testCase.tpl, writer, nil, 200)

			result, _ := request.ReadBodyResponse(writer.Result())

			assert.Equal(t, testCase.want, string(result))
			assert.Equal(t, testCase.wantErr, err)
		})
	}
}

func TestResponseXMLTemplate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		tpl     *template.Template
		want    string
		wantErr error
	}{
		"simple": {
			template.Must(template.New("sitemap.xml").ParseFiles("../../templates/sitemap.xml")),
			`<urlset xmlns="https://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>https://vibioh.fr</loc><changefreq>weekly</changefreq><priority>1.00</priority></url></urlset>`,
			nil,
		},
		"error": {
			template.Must(template.New("invalidName").ParseFiles("../../templates/sitemap.xml")),
			"",
			fmt.Errorf("template: \"invalidName\" is an incomplete or empty template"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			err := ResponseXMLTemplate(context.Background(), nil, testCase.tpl, writer, nil, 200)

			result, _ := request.ReadBodyResponse(writer.Result())

			failed := false

			if testCase.wantErr != nil && (err == nil || err.Error() != testCase.wantErr.Error()) {
				failed = true
			} else if string(result) != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("ResponseXMLTemplate() = (`%s`, `%s`), want error (`%s`, `%s`)", string(result), err, testCase.want, testCase.wantErr)
			}
		})
	}
}

func BenchmarkWriteTemplateRaw(b *testing.B) {
	tpl := template.Must(template.New("html5_template.tmpl").ParseFiles("../../templates/html5_template.tmpl"))

	for b.Loop() {
		if err := tpl.Execute(io.Discard, nil); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkWriteTemplate(b *testing.B) {
	tpl := template.Must(template.New("html5_template.tmpl").ParseFiles("../../templates/html5_template.tmpl"))

	for b.Loop() {
		if err := WriteTemplate(context.Background(), nil, tpl, io.Discard, nil, "text/html"); err != nil {
			b.Error(err)
		}
	}
}
