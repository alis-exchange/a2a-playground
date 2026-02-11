#!/bin/bash

set -e

REPO="alis-exchange/a2a-playground"
BINARY_NAME="a2a-playground"

usage() {
  echo "Usage: ./install.sh [OPTIONS]"
  echo ""
  echo "Options:"
  echo "  -v, --version <version>  Specify version to install (e.g., v1.0.0)"
  echo "  -h, --help               Show this help message"
  echo ""
  echo "If no version is specified, the latest release will be installed."
  echo ""
  echo "Examples:"
  echo "  ./install.sh                    # Install latest version"
  echo "  ./install.sh -v v1.0.0          # Install specific version"
  echo "  ./install.sh --version v1.0.0   # Install specific version"
}

VERSION=""
while [[ $# -gt 0 ]]; do
  case $1 in
    -v|--version)
      VERSION="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Error: Unknown option $1"
      usage
      exit 1
      ;;
  esac
done

# If no version specified, fetch the latest release
if [ -z "$VERSION" ]; then
  echo "Fetching latest release version..."
  VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  if [ -z "$VERSION" ]; then
    echo "Error: Could not determine latest version. Please specify a version with -v flag."
    exit 1
  fi
  echo "Latest version: $VERSION"
fi

# Detect platform and architecture
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map to Go-style platform names
if [[ "$PLATFORM" == "darwin" ]]; then
  PLATFORM="darwin"
elif [[ "$PLATFORM" == "linux" ]]; then
  PLATFORM="linux"
fi

# Map architecture
if [[ "$ARCH" == "x86_64" ]] || [[ "$ARCH" == "amd64" ]]; then
  ARCH="amd64"
elif [[ "$ARCH" == "aarch64" ]] || [[ "$ARCH" == "arm64" ]]; then
  ARCH="arm64"
elif [[ "$ARCH" == "i686" ]] || [[ "$ARCH" == "i386" ]]; then
  ARCH="386"
fi

# Build asset name
EXT=""
if [[ "$PLATFORM" == "windows" ]]; then
  EXT=".exe"
fi
ASSET_NAME="${BINARY_NAME}-${PLATFORM}-${ARCH}${EXT}"
BINARY_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET_NAME}"

# Download the binary
echo "Downloading $BINARY_NAME $VERSION for $PLATFORM/$ARCH..."
if ! curl -fL -o "$BINARY_NAME" "$BINARY_URL"; then
  echo "Error: Failed to download binary. Check that version $VERSION exists and has a binary for $PLATFORM/$ARCH."
  exit 1
fi

chmod +x "$BINARY_NAME"

# Default installation directory
if [ -z "$INSTALL_DIR" ]; then
  if [ -n "$GOPATH" ]; then
    INSTALL_DIR="$GOPATH/bin"
  elif [ -d "$HOME/go/bin" ]; then
    INSTALL_DIR="$HOME/go/bin"
  else
    INSTALL_DIR="/usr/local/bin"
  fi
fi

# Install
echo "Installing $BINARY_NAME to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
if [ -w "$INSTALL_DIR" ]; then
  mv "$BINARY_NAME" "$INSTALL_DIR/"
else
  sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
fi

# Verify
echo "Installation complete."
if "$INSTALL_DIR/$BINARY_NAME" --version 2>/dev/null; then
  :
elif "$INSTALL_DIR/$BINARY_NAME" --help >/dev/null 2>&1; then
  echo "Installed successfully."
fi
echo "Run: $BINARY_NAME --agent-url=localhost:8080"
