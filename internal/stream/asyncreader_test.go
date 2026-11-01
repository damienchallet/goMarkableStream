package stream

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// countingReaderAt records how many reads the background goroutine performs and
// fills every byte with a marker so callers can tell frames apart.
type countingReaderAt struct {
	reads  int64
	marker atomic.Int32
	err    error
	delay  time.Duration
}

func (c *countingReaderAt) ReadAt(p []byte, off int64) (int, error) {
	atomic.AddInt64(&c.reads, 1)
	if c.delay > 0 {
		time.Sleep(c.delay)
	}
	if c.err != nil {
		return 0, c.err
	}
	m := byte(c.marker.Load())
	for i := range p {
		p[i] = m
	}
	return len(p), nil
}

func (c *countingReaderAt) count() int64 { return atomic.LoadInt64(&c.reads) }

// TestAsyncFrameReader_PacingLimitsReads is the regression guard for the
// unthrottled read loop: the reader must track the consumer's tick rate instead
// of re-reading the framebuffer as fast as the device allows.
func TestAsyncFrameReader_PacingLimitsReads(t *testing.T) {
	const (
		interval = 50 * time.Millisecond
		runFor   = 400 * time.Millisecond
	)

	src := &countingReaderAt{}
	r := NewAsyncFrameReader(src, 0, 4096)
	r.SetInterval(interval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx)
	r.Resume()

	time.Sleep(runFor)
	cancel()
	got := src.count()

	// Expected ~ runFor / (interval/readAhead) = 400/25 = 16 reads. Bounds are
	// generous for slow CI, but still orders of magnitude below the thousands an
	// unpaced loop would perform against this in-memory reader.
	expected := int64(runFor / (interval / readAhead))
	if got < expected/4 {
		t.Errorf("only %d reads in %v, expected at least %d — reader is not keeping up", got, runFor, expected/4)
	}
	if got > expected*3 {
		t.Errorf("%d reads in %v, expected at most %d — reader is not paced", got, runFor, expected*3)
	}
	t.Logf("reads=%d (target ~%d)", got, expected)
}

// TestAsyncFrameReader_UnpacedByDefault documents that a zero interval keeps the
// original free-running behaviour, which tests and benchmarks rely on.
func TestAsyncFrameReader_UnpacedByDefault(t *testing.T) {
	src := &countingReaderAt{}
	r := NewAsyncFrameReader(src, 0, 4096)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx)
	r.Resume()

	time.Sleep(50 * time.Millisecond)
	cancel()

	if got := src.count(); got < 100 {
		t.Errorf("unpaced reader performed only %d reads, expected it to run freely", got)
	}
}

// TestAsyncFrameReader_ReadErrorDoesNotPublish checks that a failed read never
// hands a partially filled buffer to the encoder.
func TestAsyncFrameReader_ReadErrorDoesNotPublish(t *testing.T) {
	src := &countingReaderAt{err: errors.New("framebuffer gone")}
	r := NewAsyncFrameReader(src, 0, 4096)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx)
	r.Resume()

	time.Sleep(150 * time.Millisecond)

	if frame := r.Latest(); frame != nil {
		t.Error("Latest() returned a frame after every read failed")
	}

	// The backoff must keep a persistent failure from becoming a busy loop.
	if got := src.count(); got > 10 {
		t.Errorf("%d read attempts in 150ms despite errors, backoff is not applied", got)
	}
}

// TestAsyncFrameReader_LatestSemantics covers the triple-buffer contract:
// Latest returns the newest complete frame once, then nil until a new one lands.
func TestAsyncFrameReader_LatestSemantics(t *testing.T) {
	src := &countingReaderAt{}
	src.marker.Store(0xAB)
	r := NewAsyncFrameReader(src, 0, 4096)
	r.SetInterval(20 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx)
	r.Resume()

	var frame []byte
	deadline := time.Now().Add(time.Second)
	for frame == nil && time.Now().Before(deadline) {
		frame = r.Latest()
		if frame == nil {
			time.Sleep(5 * time.Millisecond)
		}
	}
	if frame == nil {
		t.Fatal("no frame produced within 1s")
	}
	for i, b := range frame {
		if b != 0xAB {
			t.Fatalf("byte %d = 0x%02x, want 0xAB", i, b)
		}
	}

	if again := r.Latest(); again != nil {
		t.Error("Latest() returned a frame twice without a new read")
	}
}

// TestAsyncFrameReader_PauseStopsReads verifies pausing actually stops the I/O,
// which is what keeps the device idle between writing sessions.
func TestAsyncFrameReader_PauseStopsReads(t *testing.T) {
	src := &countingReaderAt{}
	r := NewAsyncFrameReader(src, 0, 4096)
	r.SetInterval(20 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx)

	r.Resume()
	time.Sleep(60 * time.Millisecond)
	r.Pause()
	time.Sleep(20 * time.Millisecond) // let an in-flight read settle

	paused := src.count()
	time.Sleep(100 * time.Millisecond)

	if got := src.count(); got != paused {
		t.Errorf("%d reads while paused (was %d)", got-paused, paused)
	}
}
