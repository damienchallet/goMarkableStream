package stream

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/owulveryck/goMarkableStream/internal/events"
	"github.com/owulveryck/goMarkableStream/internal/pubsub"
	"github.com/owulveryck/goMarkableStream/internal/remarkable"
	"github.com/owulveryck/goMarkableStream/internal/rle"
)

var (
	rate time.Duration = 200
)

var rawFrameBuffer = sync.Pool{
	New: func() any {
		return make([]uint8, remarkable.ScreenSizeBytes) // Adjust the initial capacity as needed
	},
}

// NewStreamHandler creates a new stream handler reading from file @pointerAddr
func NewStreamHandler(file io.ReaderAt, pointerAddr int64, inputEvents *pubsub.PubSub, useRLE bool) *StreamHandler {
	return &StreamHandler{
		file:           file,
		pointerAddr:    pointerAddr,
		inputEventsBus: inputEvents,
		useRLE:         useRLE,
	}
}

// StreamHandler is an http.Handler that serves the stream of data to the client
type StreamHandler struct {
	file           io.ReaderAt
	pointerAddr    int64
	inputEventsBus *pubsub.PubSub
	useRLE         bool
}

// ServeHTTP implements http.Handler
func (h *StreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	rateStr := query.Get("rate")
	// If 'rate' parameter exists and is valid, use its value
	if rateStr != "" {
		var err error
		rateInt, err := strconv.Atoi(rateStr)
		if err != nil {
			// Handle error or keep the default value
			// For example, you can send a response with an error message
			http.Error(w, "Invalid 'rate' parameter", http.StatusBadRequest)
			return
		}
		rate = time.Duration(rateInt)
	}
	if rate < 100 {
		http.Error(w, "rate value is too low", http.StatusBadRequest)
		return
	}

	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// Send response to preflight request
		w.WriteHeader(http.StatusOK)
		return
	}

	eventC := h.inputEventsBus.Subscribe("stream")
	defer h.inputEventsBus.Unsubscribe(eventC)
	ticker := time.NewTicker(rate * time.Millisecond)
	ticker.Reset(rate * time.Millisecond)
	defer ticker.Stop()

	rawData := rawFrameBuffer.Get().([]uint8)
	defer rawFrameBuffer.Put(rawData) // Return the slice to the pool when done
	// the informations are int4, therefore store it in a uint8array to reduce data transfer
	rleWriter := rle.NewRLE(w)
	// On the rM1 there are no xochitl input events, so we poll the
	// framebuffer on every tick and use a checksum to skip unchanged frames.
	useChangeDetection := remarkable.Model == remarkable.Remarkable1
	writing := true
	lastSum := 0
	stopWriting := time.NewTicker(2 * time.Second)
	defer stopWriting.Stop()

	// Select the appropriate writer once (RLE or raw).
	var frameWriter io.Writer = w
	if h.useRLE {
		frameWriter = rleWriter
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Connection", "close")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Transfer-Encoding", "chunked")

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-eventC:
			if event.Code == 24 || event.Source == events.Touch {
				writing = true
				stopWriting.Reset(2000 * time.Millisecond)
			}
		case <-stopWriting.C:
			if !useChangeDetection {
				writing = false
			}
		case <-ticker.C:
			if useChangeDetection {
				h.fetchAndSendIfChanged(frameWriter, rawData, &lastSum)
			} else if writing {
				h.fetchAndSend(frameWriter, rawData)
			}
		}
	}
}

func (h *StreamHandler) fetchAndSendIfChanged(w io.Writer, rawData []uint8, lastSum *int) {
	_, err := h.file.ReadAt(rawData, h.pointerAddr)
	if err != nil {
		log.Println(err)
		return
	}
	s := sum(rawData)
	if s == *lastSum {
		return
	}
	*lastSum = s
	// For rM1, convert RGB565 to the RLE-compatible color-coded format.
	// The even bytes (which the RLE encoder reads at stride 2) are
	// overwritten with the converted pixel values.
	if remarkable.Model == remarkable.Remarkable1 {
		convertRGB565InPlace(rawData)
	}
	_, err = w.Write(rawData)
	if err != nil {
		log.Println("Error in writing", err)
		return
	}
	type flusher interface{ Flush() }
	if f, ok := w.(flusher); ok {
		f.Flush()
	}
}

// convertRGB565InPlace converts RGB565 little-endian pixel data to the
// single-byte color-coded format expected by the RLE encoder. Each even
// byte (the one the RLE reads) is overwritten with the converted value.
func convertRGB565InPlace(data []uint8) {
	for i := 0; i < len(data)-1; i += 2 {
		data[i] = rgb565ToColorByte(data[i], data[i+1])
	}
}

// rgb565ToColorByte converts one RGB565 LE pixel (lo, hi bytes) to the
// single-byte encoding used by the RLE stream and decoded on the client:
//
//	0-25  → grayscale (client renders value*10 per channel)
//	6,8   → red
//	12    → blue
//	20    → green
//	24    → yellow
//	30    → transparent (paper/background)
func rgb565ToColorByte(lo, hi byte) byte {
	// Extract 5-6-5 components
	r5 := (hi >> 3) & 0x1F
	g6 := uint16((hi&0x07)<<3) | uint16((lo>>5)&0x07)
	b5 := lo & 0x1F

	// Expand to 8-bit
	r := uint16(r5) * 255 / 31
	g := g6 * 255 / 63
	b := uint16(b5) * 255 / 31

	// Detect pen colors (approximate thresholds)
	if r > 180 && g < 80 && b < 80 {
		return 6 // Red
	}
	if r < 80 && g < 80 && b > 180 {
		return 12 // Blue
	}
	if g > 120 && r < 160 && b < 120 && g > r && g > b {
		return 20 // Green
	}
	if r > 180 && g > 180 && b < 120 {
		return 24 // Yellow
	}

	// Grayscale
	gray := (r + g + b) / 3
	if gray > 245 {
		return 30 // White → transparent (matches rM2 paper)
	}
	scaled := byte(gray * 25 / 245)
	// Avoid values the client interprets as colors
	switch scaled {
	case 6, 8, 12, 20, 24:
		scaled++
	}
	return scaled
}

func (h *StreamHandler) fetchAndSend(w io.Writer, rawData []uint8) {
	_, err := h.file.ReadAt(rawData, h.pointerAddr)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = w.Write(rawData)
	if err != nil {
		log.Println("Error in writing", err)
		return
	}
	type flusher interface{ Flush() }
	if f, ok := w.(flusher); ok {
		f.Flush()
	}
}

func sum(d []uint8) int {
	val := 0 // Assuming `int` is large enough to avoid overflow
	// Manual loop unrolling could be done here, but it's typically not recommended
	// for readability and maintenance reasons unless profiling identifies this loop
	// as a significant bottleneck.
	for _, v := range d {
		val += int(v)
	}
	return val
}
