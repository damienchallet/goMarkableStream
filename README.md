This is a fork of the original [goMarkableStream](https://github.com/owulveryck/goMarkableStream) to make it work on a RM1. I would welcome help to propose a pull request.

## Overview

The goMarkableStream is a lightweight and user-friendly application designed specifically for the reMarkable tablet.

Its primary goal is to enable users to stream their reMarkable tablet screen to a web browser without the need for any hacks or modifications that could void the warranty.

## Device support

- **Remarkable 1: this repo adds experimental support for Remarkable 1. I let you compile the executable** with this command (Linux) `GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -v -trimpath
-ldflags="-s -w" .`
- Remarkable 2 (not sure it still works)
- Remarkable Paper Pro (ditto)

