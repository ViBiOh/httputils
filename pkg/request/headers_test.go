package request

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateBasicAuth(t *testing.T) {
	var cases = []struct {
		username string
		password string
		want     string
	}{
		{
			"",
			"",
			"Basic Og==",
		},
		{
			"admin",
			"password",
			"Basic YWRtaW46cGFzc3dvcmQ=",
		},
	}

	for _, testCase := range cases {
		if result := GenerateBasicAuth(testCase.username, testCase.password); result != testCase.want {
			t.Errorf("GenerateBasicAuth(%v, %v) = %v, want %v", testCase.username, testCase.password, result, testCase.want)
		}
	}
}

func TestSetIP(t *testing.T) {
	var cases = []struct {
		intention string
		request   *http.Request
		ip        string
	}{
		{
			"should set header with given string",
			httptest.NewRequest(http.MethodGet, "/", nil),
			"test",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			if SetIP(testCase.request, testCase.ip); testCase.request.Header.Get(ForwardedForHeader) != testCase.ip {
				t.Errorf("SetIP(%+v, %+v) = %+v, want %+v", testCase.request, testCase.ip, testCase.request.Header.Get(ForwardedForHeader), testCase.ip)
			}
		})
	}
}

func TestGetIP(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "localhost"

	requestWithProxy := httptest.NewRequest(http.MethodGet, "/", nil)
	requestWithProxy.RemoteAddr = "localhost"
	requestWithProxy.Header.Set(ForwardedForHeader, "proxy")

	var cases = []struct {
		r    *http.Request
		want string
	}{
		{
			request,
			"localhost",
		},
		{
			requestWithProxy,
			"proxy",
		},
	}

	for _, testCase := range cases {
		if result := GetIP(testCase.r); result != testCase.want {
			t.Errorf("GetIP(%v) = %v, want %v", testCase.r, result, testCase.want)
		}
	}
}
