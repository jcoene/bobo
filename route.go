package bobo

// Portions of route.go borrowed from Martini, Copyright (c) 2013 Jeremy Saenz (MIT)

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

type route struct {
	method  string
	pattern string
	handler http.Handler
	regex   *regexp.Regexp
}

func (r *route) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(rw, req)
}

func (r *route) match(method string, path string) (found bool, params url.Values) {
	if r.method != method {
		return
	}

	matches := r.regex.FindStringSubmatch(path)
	if len(matches) > 0 && matches[0] == path {
		params = make(url.Values)
		for i, name := range r.regex.SubexpNames() {
			if len(name) > 0 {
				params.Set(name, matches[i])
			}
		}
		found = true
		return
	}

	return
}

func newRoute(method string, pattern string, handler http.Handler) *route {
	return &route{
		method:  method,
		pattern: pattern,
		handler: handler,
		regex:   compilePattern(pattern),
	}
}

func compilePattern(pattern string) *regexp.Regexp {
	r1 := regexp.MustCompile(`:[^/#?()\.\\]+`)
	pattern = r1.ReplaceAllStringFunc(pattern, func(m string) string {
		return fmt.Sprintf(`(?P<%s>[^/#?]+)`, m[1:])
	})
	r2 := regexp.MustCompile(`\*\*`)
	index := 0
	pattern = r2.ReplaceAllStringFunc(pattern, func(m string) string {
		index++
		return fmt.Sprintf(`(?P<_%d>[^#?]*)`, index)
	})
	pattern += `\/?`

	return regexp.MustCompile(pattern)
}
