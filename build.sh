#!/bin/bash
# This script cross-compiles the PeerTunnel application for various platforms.

# Exit immediately if a command exits with a non-zero status.
set -e

# The name of the application.
APP_NAME="pt"

# The directory where the compiled binaries will be stored.
OUTPUT_DIR="dist"

# The platforms to build for (OS/Architecture).
PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64")

# Clean up the output directory.
echo "Cleaning up old builds..."
rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

# Build for each platform.
for platform in "${PLATFORMS[@]}"; do
    # Split the platform string into OS and architecture.
    IFS='/' read -r -a arr <<< "$platform"
    GOOS=${arr[0]}
    GOARCH=${arr[1]}

    # Set the output binary name.
    OUTPUT_NAME="${APP_NAME}-${GOOS}-${GOARCH}"

    # Announce the build.
    echo "Building for ${GOOS}/${GOARCH}..."

    # Build the application.
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -o "${OUTPUT_DIR}/${OUTPUT_NAME}" .

done

echo "
Build complete. Binaries are in the '${OUTPUT_DIR}' directory:"
ls -l "${OUTPUT_DIR}"
