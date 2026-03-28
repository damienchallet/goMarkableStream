//go:build arm64

package remarkable

const (
	// ScreenWidth of the remarkable paper pro
	ScreenWidth = 1632
	// ScreenHeight of the remarkable paper pro
	ScreenHeight = 2154

	ScreenSizeBytes = ScreenWidth * ScreenHeight * 4

	// These values are from Max values of /dev/input/event2 (ABS_X and ABS_Y)
	MaxXValue = 11180
	MaxYValue = 15340
)

// Model defines the current device model (arm64 is always RMPP)
var Model DeviceModel = RemarkablePaperPro

// PenInputDevice is the input device for pen events
var PenInputDevice = "/dev/input/event2"

// TouchInputDevice is the input device for touch events
var TouchInputDevice = "/dev/input/event3"
