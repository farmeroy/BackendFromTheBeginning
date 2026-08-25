package shttp_test

import (
	"github.com/farmeroy/BackendFromTheBeginning/shttp"
	"testing"
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
