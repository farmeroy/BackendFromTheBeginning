package shttp_test

import (
	"reflect"
	"testing"

	"github.com/farmeroy/BackendFromTheBeginning/shttp"
)

func TestTitleCaseKey(t *testing.T) {
	for input, want := range map[string]string{
		"foo-bar":      "Foo-Bar",
		"cONTEnt-tYPE": "Content-Type",
		"host":         "Host",
		"Host-":        "Host-",
		"ha22-o3st":    "Ha22-O3st",
	} {
		if got := shttp.AsTitle(input); got != want {
			t.Errorf("TitleCaseKey(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestHTTPRespone(t *testing.T) {
	for name, tt := range map[string]struct {
		input string
		want *shttp.Response
	}{
		"200 OK (no body)": {
			input: "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n",
			want: &shttp.Response{
				StatusCode: 200,
				Headers: []shttp.Header{
					{"Content-Length", "0"},
				},
			},
		},
		"200 OK, handles whitespace": {
			input: "HTTP/1.1 200    OK  \r\nContent-Length: 0  \r\n\r\n",
			want: &shttp.Response{
				StatusCode: 200,
				Headers: []shttp.Header{
					{"Content-Length", "0"},
				},
			},
		},
		"404 Not Found (w/ body)": {
			input: "HTTP/1.1 404 Not Found\r\nContent-Length: 11\r\n\r\nHello World\r\n\r\n",
			want: &shttp.Response{
				StatusCode: 404,
				Headers: []shttp.Header{
					{"Content-Length", "11"},
				},
				Body: "Hello World",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := shttp.ParseResponse(tt.input)
			if err != nil {
				t.Errorf("ParseResponse(%q) returned err: %v", tt.input, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseResponse(%q) = %#+v, want %#+v", tt.input, got, tt.want)
			}

			if got2, err := shttp.ParseResponse(got.String()); err != nil {
				t.Errorf("ParseResponse(%q) returned error: %v", got.String(), err)
			} else if !reflect.DeepEqual(got2, got) {
				t.Errorf("ParseResponse(%q) = %#+v, want %#+v", got.String(), got2, got)
			}
		})
	}
}
