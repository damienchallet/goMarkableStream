package stream

import (
	"context"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// readAhead is how many framebuffer reads the background goroutine performs
	// per consumer tick. Reading faster than the handler encodes only produces
	// frames that are overwritten before anyone looks at them, while burning the
	// memory bandwidth xochitl needs to render the stroke in the first place.
	// Reading exactly at the consumer rate would instead let a frame sit in
	// `ready` for a full tick before being used. 2x bounds staleness at half a
	// tick for a fraction of the traffic.
	readAhead = 2

	// readErrorBackoff throttles retries when the framebuffer read fails, so a
	// vanished xochitl process cannot turn the reader into a busy loop.
	readErrorBackoff = 100 * time.Millisecond
)

// AsyncFrameReader reads the framebuffer continuously in a background goroutine,
// using triple buffering to overlap I/O with delta encoding.
//
// Three buffers rotate between three roles:
//   - writing: background goroutine is filling this via ReadAt (no lock held)
//   - ready:   latest complete frame, waiting to be consumed
//   - reading: handler is encoding this (stable, safe from overwrites)
//
// This allows the Cortex-A9's second core to read the next frame while
// the first core encodes the current one.
type AsyncFrameReader struct {
	file        io.ReaderAt
	pointerAddr int64

	mu      sync.Mutex
	writing []byte // owned by background goroutine during ReadAt
	ready   []byte // latest complete frame
	reading []byte // owned by handler during encode
	hasNew  bool

	paused int32         // atomic: 1 = paused, 0 = active
	wake   chan struct{} // signal to resume from paused state

	// interval is the consumer's current tick period in nanoseconds (atomic).
	// Zero means unpaced: read as fast as the device allows.
	interval int64
}

// SetInterval tells the reader how often the handler consumes frames, so it can
// pace itself instead of reading continuously. Safe to call concurrently with Run.
func (r *AsyncFrameReader) SetInterval(d time.Duration) {
	atomic.StoreInt64(&r.interval, int64(d))
}

// NewAsyncFrameReader creates a reader with three pre-allocated frame buffers.
func NewAsyncFrameReader(file io.ReaderAt, pointerAddr int64, frameSize int) *AsyncFrameReader {
	return &AsyncFrameReader{
		file:        file,
		pointerAddr: pointerAddr,
		writing:     make([]byte, frameSize),
		ready:       make([]byte, frameSize),
		reading:     make([]byte, frameSize),
		paused:      1, // start paused — resume when writing begins
		wake:        make(chan struct{}, 1),
	}
}

// Pause tells the reader to stop reading the framebuffer.
// The background goroutine blocks until Resume is called.
func (r *AsyncFrameReader) Pause() {
	atomic.StoreInt32(&r.paused, 1)
}

// Resume wakes the reader if it was paused.
func (r *AsyncFrameReader) Resume() {
	if atomic.CompareAndSwapInt32(&r.paused, 1, 0) {
		select {
		case r.wake <- struct{}{}:
		default:
		}
	}
}

// Run reads frames continuously until ctx is cancelled.
// Should be called in a goroutine: go reader.Run(ctx)
func (r *AsyncFrameReader) Run(ctx context.Context) {
	// Reusable timer for pacing, so the loop stays allocation-free.
	pace := time.NewTimer(0)
	if !pace.Stop() {
		<-pace.C
	}
	defer pace.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// When paused, block until resumed or cancelled.
		if atomic.LoadInt32(&r.paused) == 1 {
			select {
			case <-ctx.Done():
				return
			case <-r.wake:
			}
			continue
		}

		// ReadAt into writing buffer — no lock held, we own this buffer.
		start := time.Now()
		_, err := r.file.ReadAt(r.writing, r.pointerAddr)
		elapsed := time.Since(start)

		var sleep time.Duration
		if err != nil {
			// Publishing a partially filled buffer would stream torn or stale
			// pixels, so drop this frame and let the consumer keep the last good one.
			log.Println("Error reading framebuffer:", err)
			sleep = readErrorBackoff
		} else {
			// Swap writing and ready under lock (O(1) pointer swap).
			r.mu.Lock()
			r.writing, r.ready = r.ready, r.writing
			r.hasNew = true
			r.mu.Unlock()

			if d := time.Duration(atomic.LoadInt64(&r.interval)); d > 0 {
				sleep = d/readAhead - elapsed
			}
		}

		if sleep > 0 {
			pace.Reset(sleep)
			select {
			case <-ctx.Done():
				if !pace.Stop() {
					<-pace.C
				}
				return
			case <-pace.C:
			}
		}
	}
}

// Latest returns the latest complete frame, or nil if no new frame
// is available since the last call. The returned slice is stable
// until the next call to Latest.
func (r *AsyncFrameReader) Latest() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.hasNew {
		return nil
	}
	r.hasNew = false
	r.reading, r.ready = r.ready, r.reading
	return r.reading
}
