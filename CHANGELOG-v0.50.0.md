# goMarkableStream v0.50.0 - Release Notes

Hello! This is the **best release ever**! 🎉

Version 0.50.0 represents a massive leap forward for goMarkableStream with comprehensive improvements across performance, visual design, and user experience. This release delivers massive bandwidth optimizations without increasing CPU usage, a completely redesigned UI, full color support, and a host of quality-of-life improvements.

## 🚀 Performance & Bandwidth Optimization

### Delta Compression
- **Revolutionary bandwidth reduction**: Implemented intelligent delta compression that only transmits changed pixels between frames
- Typical bandwidth usage reduced to **1-5% of full frame** for normal e-ink usage
- **Massive optimization** without increasing CPU footprint (~10% CPU usage maintained)
- Smart threshold system: automatically sends full frames when >30% of content changes (configurable via `RK_DELTA_THRESHOLD`)
- Full frames are gzip-compressed achieving **5-10x reduction** in size
- Delta frames encode runs of changed pixels with their positions for maximum efficiency

### Compression Pipeline
- Optimized compression strategy: removed ZSTD and RLE in favor of pure delta compression
- Browser-native DecompressionStream API for gzip handling (zero client overhead)
- Intelligent frame selection algorithm minimizes bandwidth while maintaining quality

## 🎨 Full Color Support

### RGBA/BGRA Format Handling
- **Full color streaming** from PDFs and documents (firmware 3.24+)
- Native BGRA framebuffer support for reMarkable Paper Pro
- Proper color rendering for all color modes
- Fixed color handling for firmware version 3.9.4.2018+
- Plain colors now render accurately without artifacts

## 💎 UI/UX Redesign

### Modern Glassmorphism Theme
- **Complete UI overhaul** with premium glassmorphism dark theme
- Alternative reMarkable Ark light theme available
- Frosted glass effects with smooth transparency
- Modern, minimalist design language
- Responsive layout optimized for both desktop and mobile

### Connection Status & Feedback
- **Visual connection indicator** showing real-time connection state
- Clear status messages: "Connected", "Connecting...", "Disconnected"
- Fixed status indicator bug showing "Connecting..." when already connected
- Auto-reconnection with proper retry logic
- Persistent status message display without race conditions

### Mobile & Desktop UX Improvements
- Comprehensive mobile responsiveness improvements
- Enhanced accessibility across all devices
- Improved touch interactions and gesture handling
- Better sidebar menu behavior on mobile devices
- Optimized layout for portrait and landscape orientations

## 🔄 Display & Interaction Features

### Rotation Handling
- **Smooth rotation** with immediate redraw
- Fixed portrait mode rotation issues
- Proper coordinate transformation for all orientations
- Keyboard shortcut `R` for quick rotation toggle
- Rotation state persists across connections

### Laser Pointer
- **Toggleable laser pointer** feature (keyboard shortcut `L`)
- Red laser pointer follows pen hover position
- Throttled updates for optimal performance
- Laser disappears after 300ms of inactivity
- Fixed laser pointer coordinates for all rotation modes
- Replaced camera icon with dedicated laser pointer icon

### Fullscreen Support
- Native fullscreen mode support
- Seamless transition between windowed and fullscreen
- Maintains all functionality in fullscreen mode

### Help System
- Press `?` to display keyboard shortcuts overlay
- Comprehensive list of available gestures
- Context-sensitive help information

## 🌐 Tailscale Integration

### Remote Access
- **Full Tailscale integration** for secure remote access
- Dual listener support (local + Tailscale)
- Access your reMarkable from anywhere on your tailnet
- No exposure to public internet required
- Ephemeral mode with random hostname suffixes

### Tailscale Funnel
- **Public sharing** via Tailscale Funnel toggle
- Automatic QR code generation for easy access
- Dynamic Funnel enable/disable from UI
- Fixed stream reconnection after Funnel toggle
- Side menu toggle for quick Funnel control

### Configuration & Robustness
- Comprehensive Tailscale configuration documentation
- Improved robustness and error handling
- Support for headless setup with auth keys
- Verbose logging option for debugging
- Optional TLS certificate integration

## 🔧 Firmware & Device Support

### reMarkable 2 (Firmware 3.24+)
- Fixed framebuffer size mismatch on firmware v3.24+
- Reads firmware version from IMG_VERSION in /etc/os-release
- Auto-detection of firmware capabilities
- Optimized for latest firmware features

### reMarkable Paper Pro (Experimental)
- Initial support for reMarkable Paper Pro
- Fixed screen width and dimensions
- Proper pen and touch device handling
- BGRA color format support
- Fixed max X and Y coordinate values

## 🎯 Presentation Mode Enhancements

### Layer Control
- Smart layer menu visibility (only shown in present mode)
- Toggle drawing layer above or below embedded content
- Improved overlay rendering

### Reveal.js Integration
- Seamless slide control from reMarkable
- Swipe gesture navigation
- Full presentation mode support
- Enhanced compatibility

## 🐛 Bug Fixes

### Stream & Connection
- Fixed stream reconnection logic with proper retry attempts (multiple connection retries)
- Fixed spurious error when terminating stream worker
- Fixed race condition in persistent status message display
- Fixed debug logging and stream initialization on new connections
- Improved client error handling in CI workflow

### Input & Coordinates
- Fixed laser pointer coordinates across all rotation modes
- Improved pen input event accuracy
- Fixed touch gesture detection
- Prevented finding PID of xochitl_pdf_renderer when seeking framebuffer

### UI & Display
- Removed dark mode toggle and contrast slider (simplified UI)
- Removed color menu (automatic color detection)
- Removed recording functionality (streamlined interface)
- Removed version display from sidebar menu (cleaner design)
- Fixed flip functionality for 180-degree rotation

## 🗑️ Removed Features

To streamline the experience and improve maintainability, the following features were removed:
- Recording functionality from client
- Dark mode toggle (replaced with unified theme)
- Contrast slider (automatic optimization)
- Color menu (automatic color handling)
- Version display in sidebar
- ZSTD and RLE compression (superseded by delta compression)

## 📦 Build & Distribution

### Binary Naming
- Device-specific binary names: `gomarkablestream-RM2`, `gomarkablestream-RMPRO`
- Lite variants without Tailscale: `gomarkablestream-RM2-lite`, `gomarkablestream-RMPRO-lite`
- Improved goreleaser configuration
- Updated CI workflow for better reliability

### Code Quality
- Applied gofmt across entire codebase
- Fixed linting issues
- Removed unused dependencies (websocket, protoc)
- Code formatting improvements
- Better error handling throughout

## 🔄 Breaking Changes

- Removed support for old compression methods (ZSTD, RLE)
- Changed default compression to delta-based streaming
- UI theme is no longer user-selectable (unified design)
- Recording feature removed from client interface

## 📚 Documentation

- Comprehensive README restructuring
- Updated features section for v0.40.0+
- Simplified installation instructions
- Added Tailscale configuration guide
- Improved systemd service documentation
- Added troubleshooting section for missing packages

## 🙏 Community Contributions

Special thanks to all contributors:
- @yggi49 - reMarkable Paper Pro screen width fix
- @alexander-akhmetov - reMarkable Paper Pro color support
- @AngleOSaxon - PDF renderer PID detection fix
- @timnon - Flip functionality fixes
- @zihaooo - Flip endpoint implementation
- @ajmedeio - Systemd service documentation
- @mpldr - Code cleanup and compilation docs
- @hashworks - Background color styling fix

## 🎯 Migration Guide

### From v0.20 to v0.50.0

1. **Update your binary**: Download the new device-specific binary (RM2 or RMPRO)
2. **Configuration changes**:
   - Remove any ZSTD/RLE compression environment variables
   - Optionally configure `RK_DELTA_THRESHOLD` (default: 0.30)
3. **UI changes**: The new glassmorphism theme is automatic - no configuration needed
4. **Tailscale (optional)**: If using Tailscale, review the new configuration options
5. **Keyboard shortcuts**: Learn the new shortcuts (`R` for rotation, `L` for laser, `?` for help)

## 📈 Performance Comparison

| Metric | v0.20 | v0.50.0 | Improvement |
|--------|-------|---------|-------------|
| Bandwidth (typical) | 100% | 1-5% | 95-99% reduction |
| CPU usage | ~10% | ~10% | No increase |
| Full frame size | Uncompressed | Gzip compressed | 5-10x reduction |
| Color support | Grayscale only | Full RGBA | Complete color |
| UI responsiveness | Good | Excellent | Significantly improved |

## 🚀 What's Next

This release establishes a solid foundation for future enhancements. Upcoming areas of focus:
- Further performance optimizations
- Enhanced presentation mode features
- Additional device support
- Community-requested features

---

**Installation:**
```bash
# Set your device: RM2 for reMarkable 2, RMPRO for Paper Pro
DEVICE=RM2

# Download latest version
VERSION=v0.50.0
wget -O goMarkableStream https://github.com/owulveryck/goMarkableStream/releases/download/$VERSION/gomarkablestream-$DEVICE
chmod +x goMarkableStream
./goMarkableStream
```

**Access:** https://10.11.99.1:2001 (default credentials: `admin` / `password`)

---

Enjoy the best goMarkableStream experience yet! 🎉
