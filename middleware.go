package bobo

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/jcoene/gologger"
	"github.com/jcoene/statsd-client"
)

var Logger = logger.NewDefaultLogger("service")
var sentry *raven.Client
var sentryMu sync.Mutex

func getSentry() *raven.Client {
	sentryMu.Lock()
	if sentry == nil {
		sentry, _ = raven.NewClient(os.Getenv("SENTRY_DSN"), map[string]string{})
	}
	sentryMu.Unlock()
	return sentry
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

		statsd.Count(fmt.Sprintf("service.%s.success", name), 1)
		statsd.MeasureDur(fmt.Sprintf("service.%s.runtime", name), time.Since(t))
		WriteJSON(200, obj).ServeHTTP(rw, req)
	})
}
