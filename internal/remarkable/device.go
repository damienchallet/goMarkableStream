package remarkable

import (
	"log"
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

// Device configuration variables, initialized by Init().
var (
	Model            DeviceModel
	ScreenWidth      int
	ScreenHeight     int
	ScreenSizeBytes  int
	MaxXValue        int
	MaxYValue        int
	PenInputDevice   string
	TouchInputDevice string
)

// Init detects the device model and initializes all configuration variables.
// Must be called before using any other function or variable from this package.
func Init() {
	Model = detectModel()
	switch Model {
	case Remarkable1:
		ScreenWidth = 1408  // 1404 visible + 4 pixels stride padding
		ScreenHeight = 1872
		ScreenSizeBytes = ScreenWidth * ScreenHeight * 2
		MaxXValue = 15725
		MaxYValue = 20966
		PenInputDevice = "/dev/input/event0"
		TouchInputDevice = "/dev/input/event1"
	case RemarkablePaperPro:
		ScreenWidth = 1632
		ScreenHeight = 2154
		ScreenSizeBytes = ScreenWidth * ScreenHeight * 4
		MaxXValue = 11180
		MaxYValue = 15340
		PenInputDevice = "/dev/input/event2"
		TouchInputDevice = "/dev/input/event3"
	default: // Remarkable2
		Model = Remarkable2
		ScreenWidth = 1872
		ScreenHeight = 1404
		ScreenSizeBytes = ScreenWidth * ScreenHeight * 2
		MaxXValue = 15725
		MaxYValue = 20966
		PenInputDevice = "/dev/input/event1"
		TouchInputDevice = "/dev/input/event2"
	}
	log.Printf("Detected device: %v (%dx%d)", Model, ScreenWidth, ScreenHeight)
}

func detectModel() DeviceModel {
	if runtime.GOARCH == "arm64" {
		return RemarkablePaperPro
	}
	data, err := os.ReadFile("/proc/device-tree/model")
	if err != nil {
		return Remarkable2
	}
	model := strings.TrimRight(string(data), "\x00")
	if strings.Contains(model, "Prototype 1") || strings.Contains(model, "reMarkable 1") {
		return Remarkable1
	}
	return Remarkable2
}
