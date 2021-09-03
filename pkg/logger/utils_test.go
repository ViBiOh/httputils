package logger

import (
	"bytes"
	"testing"
)

func TestWriteEscapedJSON(t *testing.T) {
	type args struct {
		content string
	}

	var cases = []struct {
		intention string
		args      args
		want      string
	}{
		{
			"nothing",
			args{
				content: "test",
			},
			"test",
		},
		{
			"complex",
			args{
				content: "Text with \\ special character \"'\b\f\t\r\n.",
			},
			`Text with \\ special character \"'\b\f\t\r\n.`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			output := bytes.NewBuffer(nil)

			WriteEscapedJSON(tc.args.content, output)
			if got := output.String(); got != tc.want {
				t.Errorf("WriteEscapedJSON() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func BenchmarkWriteEscapedJSONSimple(b *testing.B) {
	output := bytes.NewBuffer(nil)

	for i := 0; i < b.N; i++ {
		output.Reset()
		WriteEscapedJSON("Text with simple character.", output)
	}
}

func BenchmarkWriteEscapedJSONComplex(b *testing.B) {
	output := bytes.NewBuffer(nil)

	for i := 0; i < b.N; i++ {
		output.Reset()
		WriteEscapedJSON("Text with special character /\"'\b\f\t\r\n.", output)
	}
}
