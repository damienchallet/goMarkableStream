//go:build linux && (arm || arm64)

package remarkable

import (
	"io"
	"os"

	"github.com/owulveryck/goMarkableStream/internal/trace"
)

// FramebufferReader wraps an os.File to provide framebuffer reading with proper cleanup.
type FramebufferReader struct {
	file   *os.File
	closed bool
}

// ReadAt implements io.ReaderAt interface.
func (r *FramebufferReader) ReadAt(p []byte, off int64) (n int, err error) {
	span := trace.BeginSpan("frame_capture")
	defer func() {
		trace.EndSpan(span, map[string]any{
			"bytes_read": n,
			"error":      err != nil,
		})
	}()

	return r.file.ReadAt(p, off)
}

// Close closes the underlying file handle. Safe to call multiple times.
func (r *FramebufferReader) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	return r.file.Close()
}

// GetFileAndPointer returns the memory file handle and pointer address for the reMarkable framebuffer
func GetFileAndPointer() (io.ReaderAt, int64, error) {
	if Model == Remarkable1 {
		return rm1Framebuffer()
	}
	pid, err := findXochitlPID()
	if err != nil {
		return nil, 0, err
	}
	file, err := os.OpenFile("/proc/"+pid+"/mem", os.O_RDONLY, os.ModeDevice)
	if err != nil {
		return nil, 0, err
	}
	pointerAddr, err := getFramePointer(pid)
	if err != nil {
		file.Close() // Close file on error
		return nil, 0, err
	}
	return &FramebufferReader{file: file}, pointerAddr, nil
}

// rm1Framebuffer opens /dev/fb0 for the rM1. It returns a reader that converts
// the raw RGB565 framebuffer to BGRA32 on every ReadAt call.
func rm1Framebuffer() (io.ReaderAt, int64, error) {
	file, err := os.OpenFile("/dev/fb0", os.O_RDONLY, os.ModeDevice)
	if err != nil {
		return nil, 0, err
	}
	return &rm1FramebufferReader{file: file}, 0, nil
}

// rm1FramebufferReader reads RGB565 data from /dev/fb0 and converts it to
// BGRA32 so the rest of the pipeline (delta encoder, client) can treat rM1
// identically to RM2 firmware 3.24+ and RMPP.
//
// The rM1 raw buffer is 1408×1872 pixels (4-pixel stride padding per row).
// We strip that padding and output the visible 1404×1872 BGRA pixels that
// match Config.Width/Height set by Init().
type rm1FramebufferReader struct {
	file *os.File
}

const (
	rm1StrideWidth  = 1408
	rm1VisibleWidth = 1404
	rm1Height       = 1872
)

// ReadAt converts one full frame of RGB565 data into BGRA32.
// p must be rm1VisibleWidth * rm1Height * 4 bytes (= Config.SizeBytes for rM1).
// The off parameter is ignored: the rM1 framebuffer is always read at offset 0.
func (r *rm1FramebufferReader) ReadAt(p []byte, off int64) (int, error) {
	raw := make([]byte, rm1StrideWidth*rm1Height*2)
	if _, err := r.file.ReadAt(raw, 0); err != nil && err != io.EOF {
		return 0, err
	}

	for row := 0; row < rm1Height; row++ {
		// The rM1 framebuffer is stored 180° rotated relative to the display:
		// read rows bottom-to-top and pixels right-to-left.
		srcRow := (rm1Height - 1 - row) * rm1StrideWidth * 2
		dstRow := row * rm1VisibleWidth * 4
		for col := 0; col < rm1VisibleWidth; col++ {
			src := srcRow + col*2
			dst := dstRow + col*4
			lo := raw[src]
			hi := raw[src+1]
			// RGB565 little-endian: lo = GGGBBBBB, hi = RRRRRGGG
			r5 := (hi >> 3) & 0x1F
			g6 := ((hi & 0x07) << 3) | ((lo >> 5) & 0x07)
			b5 := lo & 0x1F
			// Expand to 8-bit channels, output as BGRA
			p[dst+0] = (b5 << 3) | (b5 >> 2) // B
			p[dst+1] = (g6 << 2) | (g6 >> 4) // G
			p[dst+2] = (r5 << 3) | (r5 >> 2) // R
			p[dst+3] = 255                    // A
		}
	}
	return len(p), nil
}
