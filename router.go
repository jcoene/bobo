package bobo

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

type Router struct {
	NotFound    ErrorHandler
	ServerError ErrorHandler
	routes      []*route
	middleware  []Middleware
}

func (r *Router) AddMiddleware(h Middleware) {
	r.middleware = append(r.middleware, h)
}

func (r *Router) AddRoute(method string, pattern string, handler http.Handler) {
	route := newRoute(method, pattern, handler)
	r.routes = append(r.routes, route)
}

func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	response := NewResponse(rw)
	handler := r.NotFound(nil)

	for _, route := range r.routes {
		if ok, params := route.match(req.Method, req.URL.Path); ok {
			if len(params) > 0 {
				req.URL.RawQuery += "&" + params.Encode()
			}

			handler = route
			break
		}
	}

	for _, m := range r.middleware {
		handler = m(handler)
	}

	handler.ServeHTTP(response, req)
}

func NewRouter() *Router {
	return &Router{
		NotFound:    NotFound,
		ServerError: ServerError,
		routes:      make([]*route, 0),
		middleware:  make([]Middleware, 0),
	}
}
