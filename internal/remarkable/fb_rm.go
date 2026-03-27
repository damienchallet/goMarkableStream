//go:build linux && (arm || arm64)

package remarkable

import (
	"io"
	"os"
)

// GetFileAndPointer returns the memory file handle and pointer address for the reMarkable framebuffer.
//   - rM1: reads /dev/fb0 directly at offset 0
//   - rM2 and Paper Pro: reads xochitl's process memory at the framebuffer mapping address
func GetFileAndPointer() (io.ReaderAt, int64, error) {
	if Model == Remarkable1 {
		return rm1Framebuffer()
	}
	return xochitlFramebuffer()
}

// rm1Framebuffer opens /dev/fb0 directly. The rM1 exposes its framebuffer as a standard
// Linux framebuffer device that can be read at offset 0.
func rm1Framebuffer() (io.ReaderAt, int64, error) {
	file, err := os.OpenFile("/dev/fb0", os.O_RDONLY, os.ModeDevice)
	if err != nil {
		return nil, 0, err
	}
	return file, 0, nil
}

// xochitlFramebuffer reads the framebuffer from xochitl's process memory.
// Used by rM2 and Paper Pro, where the framebuffer is not directly accessible via /dev/fb0.
func xochitlFramebuffer() (io.ReaderAt, int64, error) {
	pid := findXochitlPID()
	file, err := os.OpenFile("/proc/"+pid+"/mem", os.O_RDONLY, os.ModeDevice)
	if err != nil {
		return file, 0, err
	}
	pointerAddr, err := getFramePointer(pid)
	if err != nil {
		return file, 0, err
	}
	return file, pointerAddr, nil
}
