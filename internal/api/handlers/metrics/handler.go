package metrics_handler

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsHandler struct{}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// ServeMetrics отдаёт все собранные Prometheus метрики
// @Summary Get Prometheus metrics
// @Description Returns Prometheus metrics for monitoring
// @Tags metrics
// @Produce plain
// @Success 200 "Prometheus metrics in text format"
// @Router /metrics [get]
func (h *MetricsHandler) ServeMetrics(ctx *gin.Context) {
	promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
}
