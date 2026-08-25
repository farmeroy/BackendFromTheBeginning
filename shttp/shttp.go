// shttp is a basic http library

package shttp

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Header struct{ Key, Value string }

type Request struct {
	Method  string
	Path    string
	Headers []Header
	Body    string
}

type Response struct {
	StatusCode int
	Headers    []Header
	Body       string
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
			Headers:    headers,
			Body:       body,
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
	for i := range key {
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

// write the request to the given io.Writer
func (r *Request) WriteTo(w io.Writer) (n int64, err error) {
	// write & count bytes written
	// using small closuers like this to cut down on repition
	// can be nice but sometiems thee is a performance penalty
	printf := func(format string, args ...any) error {
		m, err := fmt.Fprintf(w, format, args...)
		n += int64(m)
		return err
	}
	// remember, a HTTP request looks like this:
	// <METHOD>  <PATH>  <PROTOCOL/VERSION>
	// <HEADER>: <VALUE>
	// <HEADER>: <VALUE>
	//
	// <REQUEST BODY>

	// write the request line: like "GET /index.html HTTP/1.1"

	if err := printf("%s: %s\r\n", r.Method, r.Path); err != nil {
		return n, err
	}

	// write the headers. we don't bother here to combine them or merge duplicated headers, this is a very basic implementation
	for _, h := range r.Headers {
		if err := printf("%s: %s\r\n", h.Key, h.Value); err != nil {
			return n, err
		}
	}
	printf("\r\n")                 // there must be an empty line between headers and body
	err = printf("%s\r\n", r.Body) // write body and terminate with newline
	return n, err
}

func (resp *Response) WriteTo(w io.Writer) (n int64, err error) {
	printf := func(format string, args ...any) error {
		m, err := fmt.Fprintf(w, format, args...)
		n += int64(m)
		return err
	}
	if err := printf("HTTP/1.1 %d %s\r\n", resp.StatusCode, http.StatusText(resp.StatusCode)); err != nil {
		return n, err
	}
	for _, h := range resp.Headers {
		if err := printf("%s: %s\r\n", h.Key, h.Value); err != nil {
			return n, err
		}
	}
	printf("\r\n")
	if err := printf("%s\r\n", resp.Body); err != nil {
		return n, err
	}
	return n, err
}

// We can improve our library by writing implementations of std library interfaces

var _, _ fmt.Stringer = (*Request)(nil), (*Response)(nil) // compile time check

var _, _ encoding.TextMarshaler = (*Request)(nil), (*Response)(nil)

func (r *Request) String() string { 
	b:= new(strings.Builder)
	r.WriteTo(b) // we take advantage of our WriteTo implementation
	return b.String()
}

func (resp *Response) String() string {
	b := new(strings.Builder)
	resp.WriteTo(b)
	return b.String()
}

// TextMarshaler is used to get a bylte slice representation of a type in order to serialize it across the network or to disk
func (r *Request) MarshalText() ([]byte, error) {
	b := new(bytes.Buffer)
	r.WriteTo(b)
	return b.Bytes(), nil
}

func (resp *Response) MarshalText() ([]byte, error) {
	b := new(bytes.Buffer)
	resp.WriteTo(b)
	return b.Bytes(), nil
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	i := 0
	for {
		j := strings.Index(s[i:], "\r\n")
		if j == -1 {
			lines = append(lines, s[i:])
			return lines
		}
		lines = append(lines, s[i:i+j]) // up to but not includeing the \r\n
		i += j + 2 // skip the \r\n
	}
}

// ParseRequest parses a HTTP request from given Text
func ParseRequest(raw string) (r Request, err error) {
	// request has three parts:
	// 1. request line
	// 2. headers
	// 3. body (optional)
	lines := splitLines(raw)

	log.Println(lines)
	if len(lines) < 3 {
		return Request{}, fmt.Errorf("malformed request: should have at least 3 lines")
	}
	first := strings.Fields(lines[0])
	r.Method, r.Path = first[0], first[1]
	if !strings.HasPrefix(r.Path, "/") {
		return Request{}, fmt.Errorf("malformed request: path should start with /")
	}
	if !strings.Contains(first[2], "HTTP") {
		return Request{}, fmt.Errorf("malformed request:first line should contain HTTP version")
	}
	var foundhost bool
	var bodyStart int
	// then we have headers, up until an empty line
	for i := 1; i < len(lines); i ++ {
		if lines[i] == "" { // empty line
			bodyStart = i + 1
			break
		}
		key, val, ok := strings.Cut(lines[i], ": ")
		if !ok {
			return Request{}, fmt.Errorf("malformed request: header %q should be of form 'key: value'", lines[i])
		}
		if AsTitle(key) == "Host" { //host header is required
			foundhost = true
		}
		r.WithHeader(key, val)
	}
	end := len(lines) - 1 // recombine the body using normal newlines; skip last empty line
	r.Body = strings.Join(lines[bodyStart:end], "\r\n")
	if !foundhost {
		return Request{}, fmt.Errorf("malformed request: missing Host header")
	}
	return r, nil
}


