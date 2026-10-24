// Command match serves /match?pattern=...&text=..., answering whether text matches the regular expression.
// Compiling a pattern costs far more than matching with it, so the server keeps the 256 it used last compiled.
//
//	go run ./examples/match
//	curl 'localhost:8080/match?pattern=^go&text=gopher'
package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	lrucache "github.com/vardius/lru-cache"
)

type server struct {
	compiled *lrucache.Cache[string, *regexp.Regexp]
}

// compile returns the compiled pattern, from the cache if it is there. Two requests for a new pattern at once
// may both compile it; the second Set replaces the first's, which is just as good.
func (s *server) compile(pattern string) (*regexp.Regexp, error) {
	if re, ok := s.compiled.Get(pattern); ok {
		return re, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	s.compiled.Set(pattern, re)

	return re, nil
}

func (s *server) match(w http.ResponseWriter, r *http.Request) {
	re, err := s.compile(r.FormValue("pattern"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintln(w, re.MatchString(r.FormValue("text")))
}

func main() {
	s := &server{compiled: lrucache.New[string, *regexp.Regexp](256)}

	http.HandleFunc("GET /match", s.match)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
