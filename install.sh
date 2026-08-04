#!/bin/sh
#
# Install the latest Zod binary from GitHub Releases.
#
#   curl -fsSL https://raw.githubusercontent.com/theawakener0/Zod/main/install.sh | sh
#

set -e

repo="theawakener0/Zod"
base_url="https://github.com/${repo}/releases/latest/download"

case "$(uname -s)" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *)
    echo "error: unsupported operating system: $(uname -s)" >&2
    echo "Hmm: Windows users should download the binary directly from the Releases page." >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch="amd64" ;;
  aarch64 | arm64) arch="arm64" ;;
  *)
    echo "error: unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

asset="zod_${os}_${arch}.tar.gz"

if command -v curl >/dev/null 2>&1; then
  download() { curl -fsSL -o "$1" "$2"; }
elif command -v wget >/dev/null 2>&1; then
  download() { wget -q -O "$1" "$2"; }
else
  echo "error: need curl or wget to download the binary" >&2
  exit 1
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Downloading ${asset}"
download "${tmp}/${asset}" "${base_url}/${asset}"
download "${tmp}/checksums.txt" "${base_url}/checksums.txt"

echo "Verifying checksum..."
if command -v sha256sum >/dev/null 2>&1; then
  (cd "${tmp}" && grep "${asset}" checksums.txt | sha256sum -c -)
elif command -v shasum >/dev/null 2>&1; then
  (cd "${tmp}" && grep "${asset}" checksums.txt | shasum -a 256 -c -)
else
  echo "warning: no sha256 tool found; skipping checksum verification" >&2
fi

tar -xzf "${tmp}/${asset}" -C "${tmp}"
binary="${tmp}/zod_${os}_${arch}/zod"

if [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
  dest_dir="/usr/local/bin"
else
  dest_dir="$HOME/.local/bin"
  mkdir -p "${dest_dir}"
fi

install -m 0755 "${binary}" "${dest_dir}/zod"
echo "Installed zod to ${dest_dir}/zod"

case ":${PATH}:" in
  *":${dest_dir}:"*) ;;
  *)
    echo "NOTE: ${dest_dir} is not in your PATH. Add it with:"
    echo "  export PATH=\"${dest_dir}:\$PATH\""
    ;;
esac

echo "Run 'zod -version' to verify."
