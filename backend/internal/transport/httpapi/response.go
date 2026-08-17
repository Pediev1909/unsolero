package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

func writeJSON(response http.ResponseWriter, status int, value interface{}, logger *slog.Logger) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	if err := json.NewEncoder(response).Encode(value); err != nil {
		logger.Error("write JSON response", "error", err)
	}
}

func writeAPIError(
	response http.ResponseWriter,
	status int,
	code string,
	message string,
	fields map[string]string,
	logger *slog.Logger,
) {
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, status, errorEnvelope{
		Error: apiError{
			Code: code, Message: message, Fields: fields,
			RequestID: response.Header().Get("X-Request-ID"),
		},
	}, logger)
}
