#!/bin/sh
# install.sh — aikata installer (POSIX sh).
#
# Downloads the aikata release archive for the host OS/arch, verifies
# its SHA-256 checksum against `checksums.txt`, and installs the binary
# into $HOME/.local/bin (override with AIKATA_INSTALL_DIR).
#
# Supported targets: linux/{amd64,arm64}, darwin/{amd64,arm64}.
# Windows users: download the .zip manually from the Releases page.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/shigindo-inc/aikata/main/scripts/install.sh | sh
#   # or pin a version (avoids the GitHub /releases/latest API call):
#   curl -fsSL https://raw.githubusercontent.com/shigindo-inc/aikata/main/scripts/install.sh | AIKATA_VERSION=v0.2.1 sh
#
# Archive naming mirrors `.goreleaser.yml` `archives.name_template`.
# Keep both in sync when the template changes.

set -eu

REPO="shigindo-inc/aikata"
BINARY="aikata"
INSTALL_DIR="${AIKATA_INSTALL_DIR:-$HOME/.local/bin}"

log() { printf '%s\n' "$*" >&2; }
die() { printf 'install.sh: error: %s\n' "$*" >&2; exit 1; }

detect_os() {
    case "$(uname -s)" in
        Linux)  printf 'linux' ;;
        Darwin) printf 'darwin' ;;
        *)      die "unsupported OS '$(uname -s)'. Windows / other: install manually from https://github.com/${REPO}/releases/latest" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  printf 'amd64' ;;
        aarch64|arm64) printf 'arm64' ;;
        *)             die "unsupported architecture '$(uname -m)'" ;;
    esac
}

fetch_url() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$1"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$1"
    else
        die "neither 'curl' nor 'wget' is available"
    fi
}

download_to() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$2" "$1"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$2" "$1"
    else
        die "neither 'curl' nor 'wget' is available"
    fi
}

resolve_version() {
    if [ -n "${AIKATA_VERSION:-}" ]; then
        printf '%s' "$AIKATA_VERSION"
        return
    fi
    api="https://api.github.com/repos/${REPO}/releases/latest"
    tag=$(fetch_url "$api" \
        | grep '"tag_name":' \
        | head -n1 \
        | sed -E 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/')
    [ -n "$tag" ] || die "could not resolve latest release tag from ${api}"
    printf '%s' "$tag"
}

# Verify SHA-256 by feeding a single checksums.txt line to sha256sum/shasum.
# Selecting the line via awk avoids the "FAILED open or read" error that
# `sha256sum -c` produces for archives not present in the temp dir, and
# sidesteps the absent `--ignore-missing` flag on BSD `shasum`.
verify_sha256() {
    filepath="$1"
    checksums="$2"
    name=$(basename "$filepath")
    dir=$(dirname "$filepath")
    line=$(awk -v f="$name" '$2 == f { print; exit }' "$checksums")
    [ -n "$line" ] || die "no checksum entry for '${name}' in checksums.txt"

    if command -v sha256sum >/dev/null 2>&1; then
        printf '%s\n' "$line" | (cd "$dir" && sha256sum -c -) >/dev/null
    elif command -v shasum >/dev/null 2>&1; then
        printf '%s\n' "$line" | (cd "$dir" && shasum -a 256 -c -) >/dev/null
    else
        die "neither 'sha256sum' nor 'shasum' is available; cannot verify checksum"
    fi
}

main() {
    os=$(detect_os)
    arch=$(detect_arch)
    version=$(resolve_version)
    # GitHub release archive filenames drop the leading "v" from the tag.
    ver_no_v=${version#v}

    archive="${BINARY}_${ver_no_v}_${os}_${arch}.tar.gz"
    base_url="https://github.com/${REPO}/releases/download/${version}"
    archive_url="${base_url}/${archive}"
    checksums_url="${base_url}/checksums.txt"

    tmp=$(mktemp -d 2>/dev/null || mktemp -d -t 'aikata-install')
    trap 'rm -rf "$tmp"' EXIT INT HUP TERM

    log "==> aikata installer"
    log "    version : ${version}"
    log "    target  : ${os}/${arch}"
    log "    install : ${INSTALL_DIR}/${BINARY}"

    log "==> downloading ${archive}"
    download_to "$archive_url" "${tmp}/${archive}"

    log "==> downloading checksums.txt"
    download_to "$checksums_url" "${tmp}/checksums.txt"

    log "==> verifying SHA-256"
    verify_sha256 "${tmp}/${archive}" "${tmp}/checksums.txt"

    log "==> extracting"
    (cd "$tmp" && tar -xzf "$archive")
    [ -f "${tmp}/${BINARY}" ] || die "extracted archive does not contain '${BINARY}'"

    mkdir -p "$INSTALL_DIR"
    mv "${tmp}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    chmod +x "${INSTALL_DIR}/${BINARY}"

    # v0.6+: record the install source next to the binary so future
    # `aikata update` (self-update, planned for a v0.6.x release) can
    # pick the right upgrade path. aikata reads the first line of
    # this file at runtime via `internal/install.Detect`. Older aikata
    # binaries ignore the file harmlessly.
    cat > "${INSTALL_DIR}/aikata.install-source" <<META
install-script
version=${version}
META

    log "==> installed: ${INSTALL_DIR}/${BINARY}"

    case ":${PATH}:" in
        *":${INSTALL_DIR}:"*) ;;
        *)
            log ""
            log "warning: ${INSTALL_DIR} is not on your \$PATH."
            log "         add this to your shell profile:"
            log "           export PATH=\"${INSTALL_DIR}:\$PATH\""
            ;;
    esac

    log ""
    log "Run '${BINARY} --version' to verify."
}

main "$@"
