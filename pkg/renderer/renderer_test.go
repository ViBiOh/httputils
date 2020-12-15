package renderer

import (
	"reflect"
	"testing"
)

func TestIsStaticRootPaths(t *testing.T) {
	type args struct {
		requestPath string
	}

	var cases = []struct {
		intention string
		args      args
		want      bool
	}{
		{
			"empty",
			args{
				requestPath: "/",
			},
			false,
		},
		{
			"robots",
			args{
				requestPath: "/robots.txt",
			},
			true,
		},
		{
			"sitemap",
			args{
				requestPath: "/sitemap.xml",
			},
			true,
		},
		{
			"subpath",
			args{
				requestPath: "/test/sitemap.xml",
			},
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := isStaticRootPaths(tc.args.requestPath); got != tc.want {
				t.Errorf("isStaticRootPaths() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestFeedContent(t *testing.T) {
	type args struct {
		content map[string]interface{}
	}

	var cases = []struct {
		intention string
		instance  app
		args      args
		want      map[string]interface{}
	}{
		{
			"empty",
			app{},
			args{
				content: nil,
			},
			make(map[string]interface{}),
		},
		{
			"merge",
			app{
				content: map[string]interface{}{
					"Version": "test",
				},
			},
			args{
				content: map[string]interface{}{
					"Name": "Hello World",
				},
			},
			map[string]interface{}{
				"Version": "test",
				"Name":    "Hello World",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.feedContent(tc.args.content); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("feedContent() = %v, want %v", tc.args.content, tc.want)
			}
		})
	}
}
