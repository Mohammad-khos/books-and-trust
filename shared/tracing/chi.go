package tracing

import (
	"net/http"
	"github.com/riandyrn/otelchi"
)

func ChiMiddleware(serviceName string) func(http.Handler) http.Handler {
	return otelchi.Middleware(serviceName)
}