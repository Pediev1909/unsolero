package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"
)

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{8,128}$`)

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (recorder *responseRecorder) WriteHeader(status int) {
	if recorder.status != 0 {
		return
	}
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}

func (recorder *responseRecorder) Write(data []byte) (int, error) {
	if recorder.status == 0 {
		recorder.WriteHeader(http.StatusOK)
	}
	written, err := recorder.ResponseWriter.Write(data)
	recorder.bytes += written
	return written, err
}

func requestObservability(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		requestID := request.Header.Get("X-Request-ID")
		if !requestIDPattern.MatchString(requestID) {
			requestID = newRequestID()
		}
		request.Header.Set("X-Request-ID", requestID)
		response.Header().Set("X-Request-ID", requestID)
		recorder := &responseRecorder{ResponseWriter: response}
		next.ServeHTTP(recorder, request)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		level := slog.LevelInfo
		if status < http.StatusBadRequest &&
			(request.URL.Path == "/api/v1/health/live" || request.URL.Path == "/api/v1/health/ready") {
			level = slog.LevelDebug
		}
		logger.Log(
			request.Context(),
			level,
			"HTTP request completed",
			"request_id", requestID,
			"method", request.Method,
			"path", request.URL.Path,
			"status", status,
			"response_bytes", recorder.bytes,
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}

func recoverPanics(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error(
					"panic while serving HTTP request",
					"request_id", request.Header.Get("X-Request-ID"),
					"error", recovered,
					"stack", string(debug.Stack()),
				)
				writeAPIError(
					response,
					http.StatusInternalServerError,
					"internal_error",
					"The request could not be completed.",
					nil,
					logger,
				)
			}
		}()
		next.ServeHTTP(response, request)
	})
}

func newRequestID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "request-unknown"
	}
	return hex.EncodeToString(value)
}
