package dump

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/pkg/request"
	"github.com/ViBiOh/httputils/pkg/rollbar"
)

// Request dump request
func Request(r *http.Request) string {
	var headers bytes.Buffer
	for key, value := range r.Header {
		headers.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(value, `,`)))
	}

	var params bytes.Buffer
	for key, value := range r.URL.Query() {
		headers.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(value, `,`)))
	}

	var form bytes.Buffer
	if err := r.ParseForm(); err != nil {
		form.WriteString(fmt.Sprintf(`Error while parsing form: %v`, err))
	} else {
		for key, value := range r.PostForm {
			form.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(value, `,`)))
		}
	}

	body, err := request.ReadBody(r.Body)
	if err != nil {
		rollbar.LogError(`Error while reading body: %v`, err)
	}

	var outputPattern bytes.Buffer
	outputPattern.WriteString("%s %s\n")
	outputData := []interface{}{
		r.Method,
		r.URL.Path,
	}

	if headers.Len() != 0 {
		outputPattern.WriteString("Headers\n%s\n")
		outputData = append(outputData, headers.String())
	}

	if params.Len() != 0 {
		outputPattern.WriteString("Params\n%s\n")
		outputData = append(outputData, params.String())
	}

	if form.Len() != 0 {
		outputPattern.WriteString("Form\n%s\n")
		outputData = append(outputData, form.String())
	}

	if len(body) != 0 {
		outputPattern.WriteString("Body\n%s\n")
		outputData = append(outputData, body)
	}

	return fmt.Sprintf(outputPattern.String(), outputData...)
}
