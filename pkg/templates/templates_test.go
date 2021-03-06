package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/request"
)

func TestGetTemplates(t *testing.T) {
	var cases = []struct {
		intention string
		dir       string
		ext       string
		want      []string
		wantErr   error
	}{
		{
			"simple",
			"../../templates/",
			".xml",
			[]string{"../../templates/sitemap.xml"},
			nil,
		},
		{
			"error",
			".xml",
			".xml",
			nil,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			result, err := GetTemplates(tc.dir, tc.ext)

			failed := false

			if tc.wantErr != nil && (err == nil || err.Error() != tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(result, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetTemplates() = (%#v, `%s`), want (%#v, `%s`)", result, err, tc.want, tc.wantErr)
			}
		})
	}
}

func TestWriteTemplate(t *testing.T) {
	var cases = []struct {
		intention string
		tpl       *template.Template
		want      string
		wantErr   error
	}{
		{
			"simple",
			template.Must(template.New("css_template.html").ParseFiles("../../templates/css_template.html")),
			"html{height:100vh;width:100vw}",
			nil,
		},
		{
			"error",
			template.Must(template.New("invalidName").ParseFiles("../../templates/html5_template.html")),
			"",
			fmt.Errorf("template: \"invalidName\" is an incomplete or empty template"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := bytes.NewBuffer(nil)
			err := WriteTemplate(tc.tpl, writer, nil, "text/css")

			result := writer.String()

			failed := false

			if tc.wantErr != nil && (err == nil || err.Error() != tc.wantErr.Error()) {
				failed = true
			} else if result != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("WriteTemplate() = (`%s`, `%s`), want error (`%s`, `%s`)", result, err, tc.want, tc.wantErr)
			}
		})
	}
}

func TestResponseHTMLTemplate(t *testing.T) {
	var cases = []struct {
		intention string
		tpl       *template.Template
		want      string
		wantErr   error
	}{
		{
			"simple",
			template.Must(template.New("html5_template.html").ParseFiles("../../templates/html5_template.html")),
			`<!doctype html><html lang=fr><meta charset=utf-8><title>Golang Testing</title><meta name=description content="Golang Testing"><meta name=author content="ViBiOh"><script>function helloWorld(){console.info('Hello world!')}</script><style>html{height:100vh;width:100vw}</style><body onload=helloWorld()><h1>It works!</h1>`,
			nil,
		},
		{
			"error",
			template.Must(template.New("invalidName").ParseFiles("../../templates/html5_template.html")),
			"",
			fmt.Errorf("template: \"invalidName\" is an incomplete or empty template"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			err := ResponseHTMLTemplate(tc.tpl, writer, nil, 200)

			result, _ := request.ReadBodyResponse(writer.Result())

			failed := false

			if tc.wantErr != nil && (err == nil || err.Error() != tc.wantErr.Error()) {
				failed = true
			} else if string(result) != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("ResponseHTMLTemplate() = (`%s`, `%s`), want error (`%s`, `%s`)", string(result), err, tc.want, tc.wantErr)
			}
		})
	}
}

func TestResponseHTMLTemplateRaw(t *testing.T) {
	var cases = []struct {
		intention string
		tpl       *template.Template
		want      string
		wantErr   error
	}{
		{
			"simple",
			template.Must(template.New("html5_template.html").ParseFiles("../../templates/html5_template.html")),
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
		{
			"error",
			template.Must(template.New("invalidName").ParseFiles("../../templates/html5_template.html")),
			"",
			fmt.Errorf("template: \"invalidName\" is an incomplete or empty template"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			err := ResponseHTMLTemplateRaw(tc.tpl, writer, nil, 200)

			result, _ := request.ReadBodyResponse(writer.Result())

			failed := false

			if tc.wantErr != nil && (err == nil || err.Error() != tc.wantErr.Error()) {
				failed = true
			} else if string(result) != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("ResponseHTMLTemplateRaw() = (`%s`, `%s`), want error (`%s`, `%s`)", string(result), err, tc.want, tc.wantErr)
			}
		})
	}
}

func TestResponseXMLTemplate(t *testing.T) {
	var cases = []struct {
		intention string
		tpl       *template.Template
		want      string
		wantErr   error
	}{
		{
			"simple",
			template.Must(template.New("sitemap.xml").ParseFiles("../../templates/sitemap.xml")),
			`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd"><url><loc>https://vibioh.fr</loc><changefreq>weekly</changefreq><priority>1.00</priority></url></urlset>`,
			nil,
		},
		{
			"error",
			template.Must(template.New("invalidName").ParseFiles("../../templates/sitemap.xml")),
			"",
			fmt.Errorf("template: \"invalidName\" is an incomplete or empty template"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			err := ResponseXMLTemplate(tc.tpl, writer, nil, 200)

			result, _ := request.ReadBodyResponse(writer.Result())

			failed := false

			if tc.wantErr != nil && (err == nil || err.Error() != tc.wantErr.Error()) {
				failed = true
			} else if string(result) != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("ResponseXMLTemplate() = (`%s`, `%s`), want error (`%s`, `%s`)", string(result), err, tc.want, tc.wantErr)
			}
		})
	}
}
