#!/bin/bash

set -e

# Require version argument
if [ -z "$1" ]; then
  echo "Error: Version argument is required"
  echo "Usage: ./build.sh <version>"
  echo "Example: ./build.sh v1.0.0"
  exit 1
fi

VERSION=$1
OUTPUT_DIR="dist/$VERSION"
BINARY_NAME="a2a-playground"

# Create the output directory
mkdir -p "$OUTPUT_DIR"

# Build frontend (embeds app/dist directly at go build time)
echo "Building frontend..."
make generate build-frontend

# Platforms and architectures for cross-compilation
platforms=("linux" "darwin" "windows")
architectures=("amd64" "386" "arm64")

for OS in "${platforms[@]}"; do
  for ARCH in "${architectures[@]}"; do
    # Skip darwin/386 (not supported)
    if [ "$OS" = "darwin" ] && [ "$ARCH" = "386" ]; then
      echo "Skipping unsupported GOOS/GOARCH pair darwin/386"
      continue
    fi

    EXT=""
    if [ "$OS" = "windows" ]; then
      EXT=".exe"
    fi

    OUTPUT="$OUTPUT_DIR/${BINARY_NAME}-${OS}-${ARCH}${EXT}"
    echo "Building $BINARY_NAME@$VERSION for $OS/$ARCH..."

    GOOS=$OS GOARCH=$ARCH go build \
      -ldflags "-X main.version=$VERSION" \
      -o "$OUTPUT" \
      ./cmd/a2a-playground/

    echo "  -> $OUTPUT"
  done
done

echo ""
echo "Build complete. Binaries in $OUTPUT_DIR/"
