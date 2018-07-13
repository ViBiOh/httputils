package cors

import (
  "fmt"
  "testing"
)

func Test_Flags(t *testing.T) {
  var cases = []struct {
    intention string
    want      string
    wantType  string
  }{
    {
      `should add string origin param to flags`,
      `origin`,
      `*string`,
    },
    {
      `should add string headers param to flags`,
      `headers`,
      `*string`,
    },
    {
      `should add string methods param to flags`,
      `methods`,
      `*string`,
    },
    {
      `should add string exposes param to flags`,
      `exposes`,
      `*string`,
    },
    {
      `should add string credentials param to flags`,
      `credentials`,
      `*bool`,
    },
  }

  for _, testCase := range cases {
    result := Flags(testCase.intention)[testCase.want]

    if result == nil {
      t.Errorf("%s\nFlags() = %+v, want `%s`", testCase.intention, result, testCase.want)
    }

    if fmt.Sprintf(`%T`, result) != testCase.wantType {
      t.Errorf("%s\nFlags() = `%T`, want `%s`", testCase.intention, result, testCase.wantType)
    }
  }
}
