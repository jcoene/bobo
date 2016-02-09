package bobo

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/golang/protobuf/proto"
	"github.com/jcoene/gologger"
	"github.com/jcoene/statsd-client"
)

var Logger = logger.NewDefaultLogger("service")
var sentry *raven.Client
var sentryMu sync.Mutex

type ServiceUnavailableError struct {
	Err error
}

func (e ServiceUnavailableError) Error() string {
	return e.Err.Error()
}

func getSentry() *raven.Client {
	sentryMu.Lock()
	if sentry == nil {
		sentry, _ = raven.NewClient(os.Getenv("SENTRY_DSN"), map[string]string{})
	}
	sentryMu.Unlock()
	return sentry
}

// Request dumping middleware
func RequestDebugging(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if b, err := httputil.DumpRequest(req, true); err == nil {
			fmt.Println(string(b))
		}

		handler.ServeHTTP(rw, req)
	})
}

// Panic recovery middleware
func Recovery(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			reason := ""
			switch v := recover().(type) {
			case nil:
				return
			case error:
				reason = v.Error()
			default:
				reason = fmt.Sprint(v)
			}

			getSentry().CaptureMessage(reason, map[string]string{}, raven.NewHttp(req))
			ServerError(fmt.Errorf(reason)).ServeHTTP(rw, req)
		}()

		handler.ServeHTTP(rw, req)
	})
}

// Request logging middleware
func Logging(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		t := time.Now()
		Logger.Info("method=%s path=%s action=start", req.Method, req.URL.Path)
		handler.ServeHTTP(rw, req)
		resp := rw.(Response)
		Logger.Info("method=%s path=%s result=%d duration=%v", req.Method, req.URL.Path, resp.Status(), time.Since(t))
	})
}

// Turns a transport-agnostic service method into an instrumented JSON endpoint.
func JSON(name string, fn func(*http.Request) (interface{}, bool, error)) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		t := time.Now()
		obj, found, err := fn(req)
		if err != nil {
			switch err.(type) {
			case ValidationErrors:
				WriteJSON(400, &Errors{err.(ValidationErrors)}).ServeHTTP(rw, req)
				return
			case TimeoutError, ServiceUnavailableError:
				statsd.Count(fmt.Sprintf("service.%s.timeout", name), 1)
				WriteJSON(503, &Error{err.Error()}).ServeHTTP(rw, req)
			default:
				getSentry().CaptureError(err, map[string]string{}, raven.NewHttp(req))
				statsd.Count(fmt.Sprintf("service.%s.error", name), 1)
				ServerError(err).ServeHTTP(rw, req)
			}
			return
		}

		if !found {
			statsd.Count(fmt.Sprintf("service.%s.notfound", name), 1)
			NotFound(nil).ServeHTTP(rw, req)
			return
		}

		// Determine whether or not the client wants protobuf
		acceptPb := req.Header.Get("Accept") == CONTENT_TYPE_PROTOBUF

		// Determine whether or not we have a proto.Message
		objPb, isPb := obj.(proto.Message)

		// Write protobuf if we can, otherwise write JSON
		if isPb && acceptPb {
			WritePB(200, objPb).ServeHTTP(rw, req)
		} else {
			WriteJSON(200, obj).ServeHTTP(rw, req)
		}

		statsd.Count(fmt.Sprintf("service.%s.success", name), 1)
		statsd.MeasureDur(fmt.Sprintf("service.%s.runtime", name), time.Since(t))
	})
}
