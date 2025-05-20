// Package logger for logging incoming requests.
package logger

import (
	"context"
	"net/http"
	"time"
	"unsafe"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Initialize zap logger.
var Log *zap.Logger = zap.NewNop()

// Initializes logging.
// Returns error if level cant be determined.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Overrides writer function and saves response size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// Overrides writer function and saves response status code.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Middleware for logging incoming requests and reponses.
// Writes processing metrics such as method, path, status, size, duration of request.
func RequestLoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		Log.Info("got incoming HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
			zap.Duration("duration", duration),
		)
	})
}

// Interceptor for logging incoming requests and reponses.
// Writes processing metrics such as method, path, status, size, duration of request.
func RequestLoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		respErr := ""
		if err != nil {
			respErr = err.Error()
		}
		Log.Info("got incoming GRPC request",
			zap.String("path", info.FullMethod),
			zap.String("error", respErr),
			zap.Int("size", int(unsafe.Sizeof(resp))),
			zap.Duration("duration", duration),
		)
		return resp, err
	}
}
