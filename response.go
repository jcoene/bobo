package bobo

import (
	"net/http"
)

type Response interface {
	http.ResponseWriter
	Status() int
	Written() bool
}

type response struct {
	http.ResponseWriter
	status int
}

func NewResponse(rw http.ResponseWriter) Response {
	return &response{
		ResponseWriter: rw,
	}
}

func (r *response) Header() http.Header {
	return r.ResponseWriter.Header()
}

func (r *response) Write(data []byte) (int, error) {
	return r.ResponseWriter.Write(data)
}

func (r *response) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *response) Written() bool {
	return r.status != 0
}

func (r *response) Status() int {
	return r.status
}
