//go:build !arm64

package remarkable

const (
	// ScreenWidth of the remarkable 2
	ScreenWidth = 1872
	// ScreenHeight of the remarkable 2
	ScreenHeight = 1404

	// ScreenSizeBytes is the total memory size of the screen buffer in bytes
	ScreenSizeBytes = ScreenWidth * ScreenHeight * 2

	// MaxXValue represents the maximum X coordinate value from /dev/input/event1 (ABS_X)
	MaxXValue = 15725
	// MaxYValue represents the maximum Y coordinate value from /dev/input/event1 (ABS_Y)
	MaxYValue = 20966
)

// Model defines the current device model being used (may be updated by Init())
var Model DeviceModel = Remarkable2

// PenInputDevice is the input device for pen events (may be updated by Init() for rM1)
var PenInputDevice = "/dev/input/event1"

// TouchInputDevice is the input device for touch events (may be updated by Init() for rM1)
var TouchInputDevice = "/dev/input/event2"
