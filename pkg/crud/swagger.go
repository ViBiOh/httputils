package crud

import (
	"strings"
	"text/template"

	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/swagger"
)

const (
	pathsTemplateStr = `/{{ .Path }}:
  get:
    description: List {{ .ItemName }} with pagination
    parameters:
      - name: page
        in: query
        description: Page of list
        default: 1
        schema:
          type: integer
          format: int64
      - name: pageSize
        in: query
        description: Page size of list
        default: 20
        schema:
          type: integer
          format: int64
      - name: sort
        in: query
        description: Sort key of list
        schema:
          type: string
      - name: desc
        in: query
        description: Sort by descending order
        schema:
          type: boolean

    responses:
      '200':
        description: Paginated list of {{ .ItemName }}
        content:
          application/json:
            schema:
              type: object
              properties:
                page:
                  type: integer
                  format: int64
                  description: Page of list
                pageSize:
                  type: integer
                  format: int64
                  description: Pagesize of list
                pageCount:
                  type: integer
                  format: int64
                  description: Number of pages
                total:
                  type: integer
                  format: int64
                  description: Total count of {{ .ItemName }}
                results:
                  $ref: '#/components/schemas/{{ .ItemName }}'

      '400':
        $ref: '#/components/schemas/Error'

      '416':
        description: No more data for pagination

      '500':
        $ref: '#/components/schemas/Error'

  options:
    description: Show {{ .Path }} headers

    responses:
      '204':
        description: Headers for {{ .ItemName }}

  post:
    description: Create {{ .ItemName }}
    requestBody:
      description: {{ .ItemName }} to create
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/{{ .ItemName }}'

    responses:
      '201':
        description: {{ .ItemName }} created
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/{{ .ItemName }}'

      '400':
        $ref: '#/components/schemas/Error'

      '500':
        $ref: '#/components/schemas/Error'

/{{ .Path }}/{id}:
  parameters:
    - name: id
      in: path
      description: {{ .ItemName }}'s ID
      required: true
      schema:
        type: integer
        format: int64

  get:
    description: Retrieve {{ .ItemName }}

    responses:
      '200':
        description: {{ .ItemName }}
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/{{ .ItemName }}'

      '404':
        $ref: '#/components/schemas/Error'

      '500':
        $ref: '#/components/schemas/Error'

  options:
    description: Show {{ .Path }} headers

    responses:
      '204':
        description: Headers for {{ .ItemName }}

  put:
    description: Update {{ .ItemName }}
    requestBody:
      description: {{ .ItemName }} datas
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/{{ .ItemName }}'

    responses:
      '200':
        description: {{ .ItemName }} updated
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/{{ .ItemName }}'

      '400':
        $ref: '#/components/schemas/Error'

      '404':
        $ref: '#/components/schemas/Error'

      '500':
        $ref: '#/components/schemas/Error'

  delete:
    description: Delete {{ .ItemName }}

    responses:
      '204':
        description: {{ .ItemName }} deleted
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/{{ .ItemName }}'

      '404':
        $ref: '#/components/schemas/Error'

      '500':
        $ref: '#/components/schemas/Error'`

	componentsTemplateStr = `{{ .ItemName }}:
  type: object
  properties:
    id:
      type: integer
      format: int64
      description: {{ .ItemName }}'s identifier

{{ .ItemName }}s:
  type: array
  items:
    $ref: '#/components/schemas/{{ .ItemName }}'`
)

var (
	pathsTemplate      *template.Template
	componentsTemplate *template.Template
)

func init() {
	tpl, err := template.New("components").Parse(componentsTemplateStr)
	logger.Fatal(err)
	componentsTemplate = tpl

	tpl, err = template.New("paths").Parse(pathsTemplateStr)
	logger.Fatal(err)
	pathsTemplate = tpl
}

func (a app) Swagger() (swagger.Configuration, error) {
	if len(a.path) == 0 || len(a.name) == 0 {
		return swagger.EmptyConfiguration, nil
	}

	data := map[string]string{
		"Path":     a.path,
		"ItemName": a.name,
	}

	paths := strings.Builder{}
	components := strings.Builder{}

	if err := pathsTemplate.Execute(&paths, data); err != nil {
		return swagger.EmptyConfiguration, err
	}

	if err := componentsTemplate.Execute(&components, data); err != nil {
		return swagger.EmptyConfiguration, err
	}

	return swagger.Configuration{
		Paths:      paths.String(),
		Components: components.String(),
	}, nil
}
