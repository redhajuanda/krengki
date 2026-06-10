#!/usr/bin/env sh
set -e

REPO="redhajuanda/krengki"
BIN="krengki"
INSTALL_DIR="/usr/local/bin"

# detect OS
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)      echo "Unsupported OS: $OS" && exit 1 ;;
esac

# detect arch
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)       echo "Unsupported arch: $ARCH" && exit 1 ;;
esac

# resolve version
if [ -z "$VERSION" ]; then
  VERSION="$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
fi

if [ -z "$VERSION" ]; then
  echo "Could not determine latest version" && exit 1
fi

ASSET="${BIN}-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

echo "Installing ${BIN} ${VERSION} (${OS}/${ARCH})..."

TMP="$(mktemp)"
curl -sSfL "$URL" -o "$TMP"
chmod +x "$TMP"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "${INSTALL_DIR}/${BIN}"
else
  sudo mv "$TMP" "${INSTALL_DIR}/${BIN}"
fi

echo "Installed to ${INSTALL_DIR}/${BIN}"
${INSTALL_DIR}/${BIN} --version 2>/dev/null || true
