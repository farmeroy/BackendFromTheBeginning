// shttp is a basic http library

package shttp

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Header struct{Key, Value string}

type Request struct {
	Method string
	Path string
	Headers []Header
	Body string
}

type Response struct {
	StatusCode int
	Headers []Header
	Body string
}

func NewRequest(method, path, host, body string) (*Request, error) {
	switch {
	case method == "":
		return nil, errors.New("missing required argument: method")
	case path == "":
		return nil, errors.New("missing required argument: path")
	case !strings.HasPrefix(path, "/"):
		return nil, errors.New("path must start with /")
	case host == "":
		return nil, errors.New("missing required argument: host")
	default:
		headers := make([]Header, 2)
		headers[0] = Header{"host", host}
		if body != "" {
			headers = append(headers, Header{"Content-Length", fmt.Sprintf("%d", len(body))})
		}
		return &Request{Method: method, Path: path, Headers: headers, Body: body}, nil
	}
}

func NewResponse(status int, body string) (*Response, error) {
	switch {
	case status < 100 || status > 599:
		return nil, errors.New("invalid status code")
	default:
		if body == "" {
			body = http.StatusText(status)
		}
		headers := make([]Header, 1)
		headers[0] = Header{"Content-Length", fmt.Sprintf("%d", len(body))}
		return &Response{
			StatusCode: status,
			Headers: headers,
			Body: body,
		}, nil
	}
}

func (resp *Response) WithHeader(key, value string) *Response {
	resp.Headers = append(resp.Headers, Header{AsTitle(key), value})
	return resp
}

func (r *Request) WithHeader(key, value string) *Request {
	r.Headers = append(r.Headers, Header{AsTitle(key), value})
	return r
}

func AsTitle(key string) string {
	if key == "" {
		panic("empty header key")
	}
	if isTitleCase(key) {
		return key
	}
	return newTitleCase(key)
}

func newTitleCase(key string) string {
	var b strings.Builder
	b.Grow(len(key))
	for i := range key {
		if i == 0 || key[i-1] == '-' {
			b.WriteByte(upper(key[i]))
		} else {
			b.WriteByte(lower(key[i]))
		}
	}
	return b.String()
}

// straight from K&R C, 2nd edition, page 43. some classics never go out of style.
func lower(c byte) byte {
    /* if you're having trouble understanding this:
        the idea is as follows: A..=Z are 65..=90, and a..=z are 97..=122.
        so upper-case letters are 32 less than their lower-case counterparts (or 'a'-'A' == 32).
        rather than using the 'magic' number 32, we use 'a'-'A' to get the same result.
    */
	if c >= 'A' && c <= 'Z' {
		return c + 'a' - 'A'
	}
	return c
}

func upper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c + 'A' - 'a'
	}
	return c
}

func isTitleCase(key string) bool {
	for i:= range key {
		if i == 0 || key[i-1] == '-' {
			if key[i] >= 'a' && key[i] <= 'z' {
				return false
			}
		} else if key[i] >= 'A' && key[i] <= 'Z' {
			return false
		}
	}
	return true
}
