package observability

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func randID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// GinMiddleware starts an OTEL span for the request and ensures there's an
// X-Request-Id header. It stores a request-scoped zap logger in the gin
// context under key "logger" and also stores the span in the request context
// so downstream code can use it.
func GinMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ensure request id
		reqID := c.GetHeader("X-Request-Id")
		if reqID == "" {
			reqID = randID()
			c.Request.Header.Set("X-Request-Id", reqID)
		}
		c.Writer.Header().Set("X-Request-Id", reqID)

		// Start an OTEL span (if tracer is available). Use method+path as span name.
		ctx := c.Request.Context()
		tracer := otel.Tracer("http.server")
		spanName := c.Request.Method + " " + c.FullPath()
		// If FullPath() is empty, fallback to RequestURI
		if spanName == " " {
			spanName = c.Request.Method + " " + c.Request.RequestURI
		}
		ctx, span := tracer.Start(ctx, spanName)
		// attach context with span to request so handlers can access it
		c.Request = c.Request.WithContext(ctx)

		// attach a logger with request id and trace/span ids if present
		lg := logger.With(zap.String("request_id", reqID))
		sc := trace.SpanContextFromContext(ctx)
		if sc.IsValid() {
			lg = lg.With(zap.String("trace_id", sc.TraceID().String()), zap.String("span_id", sc.SpanID().String()))
		}
		c.Set("logger", lg)

		// Emit a small, structured log for the incoming request that includes
		// the trace and span identifiers. This helps validate tracing end-to-end
		// without requiring Jaeger to be fully wired yet.
		lg.Info("incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", func() string {
				if p := c.FullPath(); p != "" {
					return p
				}
				return c.Request.RequestURI
			}()),
		)

		// ensure span.End() after handler
		c.Next()
		span.End()

		// if response status is error, set span status (optional future enhancement)
		if status := c.Writer.Status(); status >= http.StatusInternalServerError {
			// no-op for now; could set status on span
		}
	}
}
