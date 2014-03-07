package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jcoene/bobo"
)

func main() {
	// Make a router
	router := bobo.NewRouter()

	// Add a middleware (signature func(http.Handler) http.Handler)
	router.AddMiddleware(logger)

	// Add routes which take a method string, pattern string and http.Handler
	router.AddRoute("GET", "/status", deliver(GetStatus))
	router.AddRoute("GET", "/people/:id", deliver(GetPerson))

	// Start serving, bobo style!
	http.ListenAndServe("0.0.0.0:3000", router)
}

type Person struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

// Returns a healthy status
func GetStatus(params bobo.Params) (interface{}, bool, error) {
	return map[string]string{"status": "ok"}, true, nil
}

func GetPerson(params bobo.Params) (interface{}, bool, error) {
	id := params.Int64("id")
	name := params.Get("name")
	if name == "" {
		name = "Probably Bobo!"
	}
	p := &Person{id, name}

	return p, true, nil
}

// Request logging middleware
func logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		t := time.Now()
		fmt.Printf("Started %s %s\n", req.Method, req.URL.Path)
		handler.ServeHTTP(rw, req)
		// Cast the response to a Bobo response to get written status
		resp := rw.(bobo.Response)
		fmt.Printf("Completed %s %s %d in %v\n", req.Method, req.URL.Path, resp.Status(), time.Since(t))
	})
}

// Turns a transport-agnostic service method into an instrumented endpoint.
// Service method must have signature func(bobo.Params) (interface{}, bool, error)
func deliver(fn func(bobo.Params) (interface{}, bool, error)) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		// Turn the url.Values object into a bobo.Params for convenience methods
		params := bobo.Params(req.URL.Query())

		obj, found, err := fn(params)
		if err != nil {
			bobo.ServerError(err).ServeHTTP(rw, req)
			return
		}

		if !found {
			bobo.NotFound(nil).ServeHTTP(rw, req)
			return
		}

		bobo.WriteJSON(200, obj).ServeHTTP(rw, req)
	})
}
