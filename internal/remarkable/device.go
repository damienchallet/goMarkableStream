package remarkable

import (
	"os"
	"runtime"
	"strings"
)

// DeviceModel represents the type of reMarkable device being used
type DeviceModel int

const (
	// UnknownDevice represents an unidentified reMarkable device
	UnknownDevice DeviceModel = iota
	// Remarkable1 represents the reMarkable 1 device
	Remarkable1
	// Remarkable2 represents the reMarkable 2 device
	Remarkable2
	// RemarkablePaperPro represents the reMarkable Paper Pro device
	RemarkablePaperPro
)

func (d DeviceModel) String() string {
	switch d {
	case Remarkable1:
		return "Remarkable1"
	case Remarkable2:
		return "Remarkable2"
	case RemarkablePaperPro:
		return "RemarkablePaperPro"
	default:
		return "UnknownDevice"
	}
}

// Init detects the device model and updates runtime configuration if needed.
// Must be called from main() before GetFileAndPointer() and NewEventScanner().
func Init() {
	if runtime.GOARCH == "arm64" {
		return // RMPP: already configured correctly at compile time
	}
	if detectRM1() {
		Model = Remarkable1
		PenInputDevice = "/dev/input/event0"
		TouchInputDevice = "/dev/input/event1"
		Config = FramebufferConfig{
			Width:          1404,
			Height:         1872,
			BytesPerPixel:  BytesPerPixelBGRA,
			SizeBytes:      1404 * 1872 * BytesPerPixelBGRA,
			PointerOffset:  0,
			UseBGRA:        true,
			TextureFlipped: false,
		}
	}
}

func detectRM1() bool {
	data, err := os.ReadFile("/proc/device-tree/model")
	if err != nil {
		return false
	}
	model := strings.TrimRight(string(data), "\x00")
	return strings.Contains(model, "Prototype 1") || strings.Contains(model, "reMarkable 1")
}
