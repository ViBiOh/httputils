package logger

import (
	"bytes"
	"testing"
)

func TestWriteEscapedJSON(t *testing.T) {
	type args struct {
		content string
	}

	cases := map[string]struct {
		args args
		want string
	}{
		"nothing": {
			args{
				content: "test",
			},
			"test",
		},
		"complex": {
			args{
				content: "Text with \\ special character \"'\b\f\t\r\n.",
			},
			`Text with \\ special character \"'\b\f\t\r\n.`,
		},
	}

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			output := bytes.NewBuffer(nil)

			WriteEscapedJSON(testCase.args.content, output)
			if got := output.String(); got != testCase.want {
				t.Errorf("WriteEscapedJSON() = `%s`, want `%s`", got, testCase.want)
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
