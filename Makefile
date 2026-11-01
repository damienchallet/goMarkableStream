export GOOS ?= linux
export CGO_ENABLED ?= 0
export GOARCH ?= amd64

# Default build (no tracing)
.PHONY: build
build:
	@echo "Building for GOOS=${GOOS}, GOARCH=${GOARCH}, CGO_ENABLED=${CGO_ENABLED}"
	go build -v -trimpath -tags tailscale -ldflags="-s -w" .

# Build with tracing support
.PHONY: build-with-trace
build-with-trace:
	@echo "Building with trace support for GOOS=${GOOS}, GOARCH=${GOARCH}"
	go build -v -trimpath -tags "tailscale,trace" -ldflags="-s -w" .

# reMarkable 1 and 2 builds
.PHONY: build-remarkable-1
build-remarkable-1: build-remarkable-2

.PHONY: build-remarkable-1-trace
build-remarkable-1-trace: build-remarkable-2-trace

.PHONY: build-remarkable-2
build-remarkable-2: GOARCH=arm
build-remarkable-2: GOARM=7
build-remarkable-2: build

.PHONY: build-remarkable-2-trace
build-remarkable-2-trace: GOARCH=arm
build-remarkable-2-trace: GOARM=7
build-remarkable-2-trace: build-with-trace

# reMarkable Paper Pro builds
.PHONY: build-remarkable-paper-pro
build-remarkable-paper-pro: GOARCH=arm64
build-remarkable-paper-pro: build

.PHONY: build-remarkable-paper-pro-trace
build-remarkable-paper-pro-trace: GOARCH=arm64
build-remarkable-paper-pro-trace: build-with-trace

# Local release builds into ./release/
.PHONY: release-rm1
release-rm1:
	@mkdir -p release
	@echo "Building RM1 release binaries into ./release/..."
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -trimpath -tags tailscale -ldflags="-s -w" -o release/gomarkablestream-RM1 .
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o release/gomarkablestream-RM1-lite .

.PHONY: release-all
release-all: release-rm1
	@echo "Building RM2 and Paper Pro release binaries into ./release/..."
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -trimpath -tags tailscale -ldflags="-s -w" -o release/gomarkablestream-RM2 .
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o release/gomarkablestream-RM2-lite .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -tags tailscale -ldflags="-s -w" -o release/gomarkablestream-RMPRO .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o release/gomarkablestream-RMPRO-lite .
