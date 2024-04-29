package middleware

import (
	"fourth-exam/api_gateway_evrone/api/response"
	"fourth-exam/api_gateway_evrone/internal/pkg/otlp"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
)

func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := response.NewResponseWriter(w, http.StatusOK)
		// tracing
		ctx, span := otlp.Start(r.Context(), "", r.URL.Path)
		// add request id to header
		w.Header().Add(RequestIDHeader, string(span.SpanContext().TraceID().String()))
		next.ServeHTTP(rw, r.WithContext(ctx))
		// add attributes
		span.SetAttributes(
			attribute.Key("http.method").String(r.Method),
			attribute.Key("http.url").String(r.URL.Path),
			attribute.Key("http.status_code").Int(rw.StatusCode()),
		)

		// end completes the span
		span.End()
	})
}
