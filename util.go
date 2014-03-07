package bobo

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const (
	HEADER_CONTENT_TYPE   = "Content-Type"
	HEADER_CONTENT_LENGTH = "Content-Length"
	CONTENT_TYPE_JSON     = "application/json"
	ERROR_NOT_FOUND       = "not found"
)

type Error struct {
	Error string `json:"error"`
}

type ErrorHandler func(error) http.Handler

func WriteJSON(status int, obj interface{}) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		buf, _ := json.Marshal(obj)
		rw.Header().Add(HEADER_CONTENT_TYPE, CONTENT_TYPE_JSON)
		rw.Header().Add(HEADER_CONTENT_LENGTH, strconv.Itoa(len(buf)))
		rw.WriteHeader(status)
		rw.Write(buf)
	})
}

func NotFound(err error) http.Handler {
	return WriteJSON(404, &Error{ERROR_NOT_FOUND})
}

func ServerError(err error) http.Handler {
	return WriteJSON(500, &Error{err.Error()})
}
