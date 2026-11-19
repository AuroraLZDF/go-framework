// Copyright 2022 Innkeeper AuroraLZDF auroralzdf@gmail.com. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/AuroraLZDF/go-framework.

package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// httpRequestsTotal 按 method/route/status 统计请求总数.
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)

	// httpRequestDuration 按 method/route 统计请求耗时分布.
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

// MetricsHandler 返回暴露给 Prometheus 抓取的 /metrics 处理器.
func MetricsHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// Metrics 是一个 Gin 中间件，按路由记录请求总数与耗时指标.
// 未匹配到注册路由的请求记为 UNKNOWN；/metrics 自身不计入.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request == nil || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "UNKNOWN"
		}
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, route, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}
