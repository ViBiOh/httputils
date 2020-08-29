package logger

import "testing"

func TestEscapeString(t *testing.T) {
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
				content: "Text with special character /\"'\b\f\t\r\n.",
			},
			`Text with special character /\"'\b\f\t\r\n.`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := EscapeString(tc.args.content); got != tc.want {
				t.Errorf("EscapeString() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func BenchmarkEscapeStringSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EscapeString("Text with simple character.")
	}
}

func BenchmarkEscapeStringComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EscapeString("Text with special character /\"'\b\f\t\r\n.")
	}
}
