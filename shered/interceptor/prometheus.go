package interceptor

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type MetricsCollector struct{
	// Счетчик активных запросов (Gauge)
	inflightRequests *prometheus.GaugeVec

	// Гистограмма длительности запросов
	requestDuration *prometheus.HistogramVec

	// Счетчик завершенных запросов (по статусам)
	requestsTotal *prometheus.CounterVec

	// Счетчик ошибок (для быстрого мониторинга)
	errorsTotal *prometheus.CounterVec
} 

func NewMetricsCollector() * MetricsCollector{
	//датчик, который показывает число активных запросов
	inflight := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
		Name: "grpc_server_inflight_request", 
		Help: "Current number of in-flight gRPC requests", 
	},
	[]string{"method"},
	)

	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "grpc_server_handling_seconds",
			Help: "Histogram of gRPC request durations in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}, // Это "ведра" для подсчета: сколько запросов выполнилось за 0.005с, 0.01с и тd
		},
		[]string{"method", "code"},
	)

	total := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_server_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "code"},
	)

	errors := prometheus. NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_server_errors_total",
			Help: "Total number of gRPC errors by type",
		},
		[]string{"method", "code"},
	)

	// Регистрируем метрики
	prometheus.MustRegister(inflight, duration, total, errors)

	return &MetricsCollector{
		inflightRequests: inflight,
		requestDuration: duration,
		requestsTotal: total,
		errorsTotal: errors,
	}
}
// UnaryServerInterceptor возвращает gRPC интерсептор для сбора метрик
func (m *MetricsCollector) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := info.FullMethod
		
		// Увеличиваем счетчик активных запросов
		m.inflightRequests.WithLabelValues(method).Inc()
		
		start := time.Now()
		
		resp, err := handler(ctx, req)
		
		duration := time.Since(start).Seconds()
		
		// Определяем gRPC статус
		code := status.Code(err).String()
		
		// Уменьшаем счетчик активных запросов
		m.inflightRequests.WithLabelValues(method).Dec()
		
		// 7. Записываем метрики
		m.requestDuration.WithLabelValues(method, code).Observe(duration)
		m.requestsTotal.WithLabelValues(method, code).Inc()
		
		if err != nil {
			m.errorsTotal.WithLabelValues(method, code).Inc()
		}
		
		return resp, err
	}
}