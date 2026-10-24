package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	lrucache "github.com/vardius/lru-cache"
)

func TestMatch(t *testing.T) {
	s := &server{compiled: lrucache.New[string, *regexp.Regexp](2)}
	srv := httptest.NewServer(http.HandlerFunc(s.match))
	defer srv.Close()

	tests := []struct {
		pattern, text string
		status        int
		body          string
	}{
		{`^go`, "gopher", http.StatusOK, "true\n"},
		{`^go`, "mongo", http.StatusOK, "false\n"},
		{`(`, "anything", http.StatusBadRequest, "error parsing regexp: missing closing ): `(`\n"},
	}
	for _, tt := range tests {
		resp, err := http.Get(srv.URL + "?" + url.Values{"pattern": {tt.pattern}, "text": {tt.text}}.Encode())
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != tt.status || string(body) != tt.body {
			t.Errorf("match %q against %q = %d %q, want %d %q", tt.text, tt.pattern, resp.StatusCode, body, tt.status, tt.body)
		}
	}

	// The valid pattern was compiled once and reused; the invalid one wasn't cached.
	if n := s.compiled.Len(); n != 1 {
		t.Errorf("%d patterns cached, want 1", n)
	}
}

func TestCompileReusesAndEvicts(t *testing.T) {
	s := &server{compiled: lrucache.New[string, *regexp.Regexp](2)}
	compile := func(pattern string) *regexp.Regexp {
		t.Helper()
		re, err := s.compile(pattern)
		if err != nil {
			t.Fatal(err)
		}
		return re
	}

	a := compile("a")
	if compile("a") != a {
		t.Error("the second compile of a pattern didn't reuse the first")
	}
	compile("b")
	compile("c") // evicts a, the least recently used
	if compile("a") == a {
		t.Error("a was still cached after two newer patterns filled the cache")
	}
}
