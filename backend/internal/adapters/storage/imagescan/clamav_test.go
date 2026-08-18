package imagescan

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	adminports "rigmark/internal/modules/admin/ports"
)

// pngBytes is a minimal payload whose detected content type is image/png, so
// the media-family check passes and the clamd path is what is under test.
func pngBytes() []byte {
	return append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 32)...)
}

// fakeClamd answers one connection with reply and records what it received.
func fakeClamd(t *testing.T, reply string, hangUp bool) (string, func() []byte) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	received := make(chan []byte, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			received <- nil
			return
		}
		defer func() { _ = conn.Close() }()
		if hangUp {
			received <- nil
			return
		}
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		buffer := make([]byte, 0, 1024)
		chunk := make([]byte, 512)
		for {
			read, readErr := conn.Read(chunk)
			buffer = append(buffer, chunk[:read]...)
			// The client finishes an INSTREAM with a zero-length frame; PING
			// finishes with its own terminator.
			if bytes.HasSuffix(buffer, []byte{0, 0, 0, 0}) || bytes.HasSuffix(buffer, []byte("zPING\x00")) {
				break
			}
			if readErr != nil {
				break
			}
		}
		_, _ = io.WriteString(conn, reply)
		received <- buffer
	}()
	return listener.Addr().String(), func() []byte { return <-received }
}

func TestClamAVAcceptsCleanPayload(t *testing.T) {
	address, recorded := fakeClamd(t, "stream: OK\x00", false)
	scanner := NewClamAV(address, 2*time.Second, 0)

	if err := scanner.Scan(context.Background(), pngBytes(), "image/png"); err != nil {
		t.Fatalf("Scan() rejected a clean payload: %v", err)
	}

	sent := recorded()
	if !bytes.HasPrefix(sent, []byte("zINSTREAM\x00")) {
		t.Fatalf("clamd did not receive an INSTREAM command, got %q", sent[:min(16, len(sent))])
	}
	// The frame after the command must carry the payload length, otherwise
	// clamd would silently scan the wrong number of bytes.
	header := sent[len("zINSTREAM\x00"):][:4]
	if got := binary.BigEndian.Uint32(header); got != uint32(len(pngBytes())) {
		t.Fatalf("chunk header length = %d, want %d", got, len(pngBytes()))
	}
	if !bytes.HasSuffix(sent, []byte{0, 0, 0, 0}) {
		t.Fatal("the stream was not terminated with a zero-length frame")
	}
}

func TestClamAVRejectsInfectedPayload(t *testing.T) {
	address, _ := fakeClamd(t, "stream: Eicar-Test-Signature FOUND\x00", false)
	scanner := NewClamAV(address, 2*time.Second, 0)

	err := scanner.Scan(context.Background(), pngBytes(), "image/png")
	if !errors.Is(err, ErrMediaInfected) {
		t.Fatalf("Scan() error = %v, want ErrMediaInfected", err)
	}
}

// An unrecognised verdict must never be read as clean. This is the failure
// mode that would quietly turn the scanner into decoration.
func TestClamAVTreatsUnknownVerdictAsUnavailable(t *testing.T) {
	for name, reply := range map[string]string{
		"clamd error":  "ERROR: size limit exceeded\x00",
		"empty":        "\x00",
		"unrecognised": "stream: something else\x00",
	} {
		t.Run(name, func(t *testing.T) {
			address, _ := fakeClamd(t, reply, false)
			scanner := NewClamAV(address, 2*time.Second, 0)

			err := scanner.Scan(context.Background(), pngBytes(), "image/png")
			if !errors.Is(err, adminports.ErrMediaScanUnavailable) {
				t.Fatalf("Scan() error = %v, want ErrMediaScanUnavailable", err)
			}
		})
	}
}

func TestClamAVFailsClosedWhenDaemonIsUnreachable(t *testing.T) {
	// Port 1 on loopback refuses connections without waiting for a timeout.
	scanner := NewClamAV("127.0.0.1:1", time.Second, 0)

	err := scanner.Scan(context.Background(), pngBytes(), "image/png")
	if !errors.Is(err, adminports.ErrMediaScanUnavailable) {
		t.Fatalf("Scan() error = %v, want ErrMediaScanUnavailable", err)
	}
}

func TestClamAVFailsClosedWhenDaemonHangsUp(t *testing.T) {
	address, _ := fakeClamd(t, "", true)
	scanner := NewClamAV(address, time.Second, 0)

	err := scanner.Scan(context.Background(), pngBytes(), "image/png")
	if !errors.Is(err, adminports.ErrMediaScanUnavailable) {
		t.Fatalf("Scan() error = %v, want ErrMediaScanUnavailable", err)
	}
}

func TestClamAVRejectsOversizedUploadBeforeDialling(t *testing.T) {
	// No listener exists; if the size guard did not run first this would fail
	// with a dial error instead, so the assertion also proves the ordering.
	scanner := NewClamAV("127.0.0.1:1", time.Second, 16)

	err := scanner.Scan(context.Background(), pngBytes(), "image/png")
	if !errors.Is(err, adminports.ErrMediaScanUnavailable) {
		t.Fatalf("Scan() error = %v, want ErrMediaScanUnavailable", err)
	}
}

func TestClamAVRejectsMismatchedMediaType(t *testing.T) {
	address, _ := fakeClamd(t, "stream: OK\x00", false)
	scanner := NewClamAV(address, 2*time.Second, 0)

	err := scanner.Scan(context.Background(), pngBytes(), "image/jpeg")
	if !errors.Is(err, adminports.ErrMediaScanUnavailable) {
		t.Fatalf("Scan() error = %v, want ErrMediaScanUnavailable", err)
	}
}

func TestClamAVReadyReportsAHealthyDaemon(t *testing.T) {
	address, _ := fakeClamd(t, "PONG\x00", false)
	scanner := NewClamAV(address, 2*time.Second, 0)

	if err := scanner.Ready(context.Background()); err != nil {
		t.Fatalf("Ready() reported a healthy daemon as unavailable: %v", err)
	}
}

func TestClamAVReadyFailsWhenDaemonIsAbsent(t *testing.T) {
	scanner := NewClamAV("127.0.0.1:1", time.Second, 0)

	if err := scanner.Ready(context.Background()); !errors.Is(err, adminports.ErrMediaScanUnavailable) {
		t.Fatalf("Ready() error = %v, want ErrMediaScanUnavailable", err)
	}
}
