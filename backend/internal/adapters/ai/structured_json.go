package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var ErrInvalidStructuredOutput = errors.New("invalid structured AI output")

// DecodeStructuredJSON is shared by concrete provider adapters. It bounds the
// response, rejects unknown fields and trailing JSON, and never accepts a
// provider-specific untyped map into the application layer.
func DecodeStructuredJSON[T interface{}](reader io.Reader, maxBytes int64) (T, error) {
	var result T
	if reader == nil || maxBytes < 1 {
		return result, fmt.Errorf("%w: reader or response limit", ErrInvalidStructuredOutput)
	}
	raw, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return result, fmt.Errorf("%w: read response: %v", ErrInvalidStructuredOutput, err)
	}
	if int64(len(raw)) > maxBytes {
		return result, fmt.Errorf("%w: response exceeds %d bytes", ErrInvalidStructuredOutput, maxBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("%w: %v", ErrInvalidStructuredOutput, err)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return result, fmt.Errorf("%w: trailing content", ErrInvalidStructuredOutput)
	}
	return result, nil
}
