package crud

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ViBiOh/httputils/v3/pkg/swagger"
)

var (
	expectedOutput = `/crud:
  get:
    description: List Item with pagination
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
        description: Paginated list of Item
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
                  description: Total count of Item
                results:
                  $ref: '#/components/schemas/Item'

      '400':
        $ref: '#/components/schemas/Error'

      '416':
        description: No more data for pagination

      '500':
        $ref: '#/components/schemas/Error'

  options:
    description: Show crud headers

    responses:
      '204':
        description: Headers for Item

  post:
    description: Create Item
    requestBody:
      description: Item to create
      required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Item'

    responses:
      '201':
        description: Item created
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Item'

      '400':
        $ref: '#/components/schemas/Error'

      '500':
        $ref: '#/components/schemas/Error'

/crud/{id}:
  parameters:
    - name: id
      in: path
      description: Item's ID
      required: true
      schema:
        type: integer
        format: int64

  get:
    description: Retrieve Item

    responses:
      '200':
        description: Item
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Item'

      '404':
        $ref: '#/components/schemas/Error'

      '500':
        $ref: '#/components/schemas/Error'

  options:
    description: Show crud headers

    responses:
      '204':
        description: Headers for Item

  put:
    description: Update Item
    requestBody:
      description: Item datas
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Item'

    responses:
      '200':
        description: Item updated
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Item'

      '400':
        $ref: '#/components/schemas/Error'

      '404':
        $ref: '#/components/schemas/Error'

      '500':
        $ref: '#/components/schemas/Error'

  delete:
    description: Delete Item

    responses:
      '204':
        description: Item deleted
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Item'

      '404':
        $ref: '#/components/schemas/Error'

      '500':
        $ref: '#/components/schemas/Error'`

	expectedOutputComponents = `Item:
  type: object
  properties:
    id:
      type: integer
      format: int64
      description: Item's identifier

Items:
  type: array
  items:
    $ref: '#/components/schemas/Item'`
)

func TestSwagger(t *testing.T) {
	var cases = []struct {
		intention string
		path      string
		itemName  string
		want      swagger.Configuration
		wantErr   error
	}{
		{
			"no path",
			"",
			"",
			swagger.EmptyConfiguration,
			nil,
		},
		{
			"no name",
			"crud",
			"",
			swagger.EmptyConfiguration,
			nil,
		},
		{
			"simple",
			"crud",
			"Item",
			swagger.Configuration{
				Paths:      expectedOutput,
				Components: expectedOutputComponents,
			},
			nil,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			result, err := app{
				path: testCase.path,
				name: testCase.itemName,
			}.Swagger()

			failed := false

			if testCase.wantErr != nil && !errors.Is(err, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(result, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("Swagger() = (%v, `%s`), want (%v, `%s`)", result, err, testCase.want, testCase.wantErr)
			}
		})
	}
}
