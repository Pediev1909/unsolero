package imagescan

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	adminports "rigmark/internal/modules/admin/ports"
)

// ClamAV scans upload bytes with a clamd daemon over its INSTREAM command.
//
// Every failure path returns ErrMediaScanUnavailable rather than nil. An
// unreachable daemon, a timeout, a truncated reply or an unrecognised response
// all mean the same thing operationally: nobody has confirmed these bytes are
// safe, so they must not reach durable storage. Failing open here would make
// the scanner decorative.
type ClamAV struct {
	// Address is the clamd host:port. clamd speaks a plaintext protocol and is
	// expected to be reachable only on a private network.
	Address string
	Timeout time.Duration
	// MaxBytes bounds what is streamed. clamd enforces its own StreamMaxLength
	// and aborts the connection when exceeded, which surfaces as a confusing
	// write error, so the limit is applied here first.
	MaxBytes int
}

const (
	defaultClamAVTimeout  = 20 * time.Second
	defaultClamAVMaxBytes = 20 << 20
	clamAVChunkSize       = 64 << 10
)

func NewClamAV(address string, timeout time.Duration, maxBytes int) ClamAV {
	if timeout <= 0 {
		timeout = defaultClamAVTimeout
	}
	if maxBytes <= 0 {
		maxBytes = defaultClamAVMaxBytes
	}
	return ClamAV{Address: address, Timeout: timeout, MaxBytes: maxBytes}
}

// Ready reports whether clamd answers. It exists so a misconfigured scanner is
// discovered at startup rather than by the first administrator who uploads an
// image.
func (scanner ClamAV) Ready(ctx context.Context) error {
	conn, err := scanner.dial(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if _, err = conn.Write([]byte("zPING\x00")); err != nil {
		return fmt.Errorf("%w: ping clamd: %v", adminports.ErrMediaScanUnavailable, err)
	}
	reply, err := bufio.NewReader(conn).ReadString(0)
	if err != nil {
		return fmt.Errorf("%w: read clamd ping reply: %v", adminports.ErrMediaScanUnavailable, err)
	}
	if strings.TrimRight(reply, "\x00") != "PONG" {
		return fmt.Errorf("%w: unexpected clamd ping reply", adminports.ErrMediaScanUnavailable)
	}
	return nil
}

func (scanner ClamAV) Scan(ctx context.Context, data []byte, declaredType string) error {
	// The declared media family is still checked. clamd finds malware, not a
	// payload whose bytes disagree with the content type it claims to be.
	if err := (Development{}).Scan(ctx, data, declaredType); err != nil {
		return err
	}
	if len(data) > scanner.MaxBytes {
		return fmt.Errorf("%w: upload exceeds the scannable size", adminports.ErrMediaScanUnavailable)
	}

	conn, err := scanner.dial(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if _, err = conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return fmt.Errorf("%w: start clamd stream: %v", adminports.ErrMediaScanUnavailable, err)
	}
	// INSTREAM frames each chunk with a big-endian length prefix and ends with
	// a zero-length frame.
	for start := 0; start < len(data); start += clamAVChunkSize {
		end := start + clamAVChunkSize
		if end > len(data) {
			end = len(data)
		}
		var header [4]byte
		binary.BigEndian.PutUint32(header[:], uint32(end-start))
		if _, err = conn.Write(header[:]); err != nil {
			return fmt.Errorf("%w: write clamd chunk header: %v", adminports.ErrMediaScanUnavailable, err)
		}
		if _, err = conn.Write(data[start:end]); err != nil {
			return fmt.Errorf("%w: write clamd chunk: %v", adminports.ErrMediaScanUnavailable, err)
		}
	}
	if _, err = conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return fmt.Errorf("%w: terminate clamd stream: %v", adminports.ErrMediaScanUnavailable, err)
	}

	reply, err := bufio.NewReader(conn).ReadString(0)
	if err != nil {
		return fmt.Errorf("%w: read clamd verdict: %v", adminports.ErrMediaScanUnavailable, err)
	}
	return interpretClamAVVerdict(reply)
}

// ErrMediaInfected reports a positive detection. It is distinct from
// ErrMediaScanUnavailable so the upload path can tell "we found something" from
// "we could not look".
var ErrMediaInfected = errors.New("media failed malware scanning")

func interpretClamAVVerdict(reply string) error {
	verdict := strings.TrimSpace(strings.TrimRight(reply, "\x00"))
	switch {
	case strings.HasSuffix(verdict, "OK"):
		return nil
	case strings.HasSuffix(verdict, "FOUND"):
		return fmt.Errorf("%w: %s", ErrMediaInfected, verdict)
	default:
		// Includes clamd's ERROR replies and anything unrecognised. Treating an
		// unknown verdict as clean is the one mistake this adapter must not make.
		return fmt.Errorf("%w: clamd replied %q", adminports.ErrMediaScanUnavailable, verdict)
	}
}

func (scanner ClamAV) dial(ctx context.Context) (net.Conn, error) {
	if strings.TrimSpace(scanner.Address) == "" {
		return nil, fmt.Errorf("%w: no clamd address is configured", adminports.ErrMediaScanUnavailable)
	}
	timeout := scanner.Timeout
	if timeout <= 0 {
		timeout = defaultClamAVTimeout
	}
	dialContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(dialContext, "tcp", scanner.Address)
	if err != nil {
		return nil, fmt.Errorf("%w: dial clamd: %v", adminports.ErrMediaScanUnavailable, err)
	}
	// The deadline covers the whole exchange, so a daemon that accepts the
	// connection and then stalls cannot hold an upload open indefinitely.
	if err = conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("%w: set clamd deadline: %v", adminports.ErrMediaScanUnavailable, err)
	}
	return conn, nil
}

var _ adminports.ImageScanner = ClamAV{}
