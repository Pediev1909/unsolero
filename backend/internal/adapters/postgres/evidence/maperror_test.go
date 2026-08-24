package evidencepostgres

import (
	"errors"
	"testing"
)

// mapError wrapped unconditionally, so handing it a nil error produced a
// non-nil one. Every caller until ListSources passed an error it already knew
// was non-nil, which hid it; the first caller to pass rows.Err() after a
// successful read turned a working query into a 500.
func TestMapErrorLetsSuccessThrough(t *testing.T) {
	if err := mapError("list evidence sources", nil); err != nil {
		t.Fatalf("mapError(nil) = %v, want nil", err)
	}
}

func TestMapErrorStillWrapsARealFailure(t *testing.T) {
	cause := errors.New("connection reset")
	err := mapError("list evidence sources", cause)
	if err == nil {
		t.Fatal("a real failure was swallowed")
	}
	if !errors.Is(err, cause) {
		t.Errorf("the cause was lost: %v", err)
	}
}
