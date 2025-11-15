package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

func TracingMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)

	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), c.FullPath())
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		status := c.Writer.Status()
		if status >= http.StatusInternalServerError {
			span.SetStatus(codes.Error, http.StatusText(status))
		}
	}
}
