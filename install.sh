#!/bin/sh
# mk installer — https://github.com/subut0n/mk
# Usage: curl -fsSL https://raw.githubusercontent.com/subut0n/mk/main/install.sh | sh
set -eu

REPO="subut0n/mk"
BINARY="mk"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

log()   { printf '  \033[1;35m%s\033[0m %s\n' "$1" "$2"; }
err()   { printf '  \033[1;31merror:\033[0m %s\n' "$1" >&2; exit 1; }

need() {
    command -v "$1" >/dev/null 2>&1 || err "$1 is required but not found"
}

fetch() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$1"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$1"
    else
        err "curl or wget is required"
    fi
}

fetch_to() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$2" "$1"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$2" "$1"
    else
        err "curl or wget is required"
    fi
}

# ---------------------------------------------------------------------------
# Detect platform
# ---------------------------------------------------------------------------

detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux"  ;;
        Darwin*) echo "darwin" ;;
        *)       err "unsupported OS: $(uname -s)" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *)             err "unsupported architecture: $(uname -m)" ;;
    esac
}

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------

VERSION=""
while [ $# -gt 0 ]; do
    case "$1" in
        --version) VERSION="$2"; shift 2 ;;
        *)         err "unknown option: $1" ;;
    esac
done

# ---------------------------------------------------------------------------
# Resolve version
# ---------------------------------------------------------------------------

if [ -z "$VERSION" ]; then
    VERSION=$(fetch "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' | head -1 | cut -d'"' -f4)
    [ -n "$VERSION" ] || err "could not determine latest version"
fi

# ---------------------------------------------------------------------------
# Download and verify
# ---------------------------------------------------------------------------

OS=$(detect_os)
ARCH=$(detect_arch)
ASSET="${BINARY}-${OS}-${ARCH}"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

log "mk" "${VERSION}  (${OS}/${ARCH})"

log "downloading" "${ASSET}"
fetch_to "${BASE_URL}/${ASSET}" "${TMPDIR}/${ASSET}"

log "verifying" "checksum"
fetch_to "${BASE_URL}/checksums.txt" "${TMPDIR}/checksums.txt"

(cd "$TMPDIR" && grep "${ASSET}" checksums.txt | sha256sum -c --quiet -) \
    || err "checksum verification failed"

chmod +x "${TMPDIR}/${ASSET}"

# ---------------------------------------------------------------------------
# Install
# ---------------------------------------------------------------------------

INSTALL_DIR="/usr/local/bin"
if [ -w "$INSTALL_DIR" ]; then
    mv "${TMPDIR}/${ASSET}" "${INSTALL_DIR}/${BINARY}"
    log "installed" "${INSTALL_DIR}/${BINARY}"
elif command -v sudo >/dev/null 2>&1; then
    sudo mv "${TMPDIR}/${ASSET}" "${INSTALL_DIR}/${BINARY}"
    log "installed" "${INSTALL_DIR}/${BINARY} (via sudo)"
else
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "$INSTALL_DIR"
    mv "${TMPDIR}/${ASSET}" "${INSTALL_DIR}/${BINARY}"
    log "installed" "${INSTALL_DIR}/${BINARY}"
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) ;;
        *) log "note" "add ${INSTALL_DIR} to your PATH" ;;
    esac
fi

log "done" "run 'mk' in any directory with a Makefile"
