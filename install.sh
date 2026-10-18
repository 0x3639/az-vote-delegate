#!/bin/sh
# Install a verified prebuilt binary. Does not touch node keys, config or votes.
set -eu
repo=0x3639/az-vote-delegate
case "$(uname -s)" in Linux) os=linux ;; *) echo 'This installer supports Linux pillar hosts.' >&2; exit 1 ;; esac
case "$(uname -m)" in x86_64|amd64) arch=amd64 ;; aarch64|arm64) arch=arm64 ;; *) echo 'Supported architectures: amd64 and arm64.' >&2; exit 1 ;; esac
for tool in curl tar awk mktemp install; do command -v "$tool" >/dev/null 2>&1 || { echo "Missing required tool: $tool" >&2; exit 1; }; done
command -v sha256sum >/dev/null 2>&1 || { echo 'sha256sum is required (coreutils).' >&2; exit 1; }
version=${AZ_VERSION:-}
if [ -z "$version" ]; then
  release_url=$(curl --proto '=https' --tlsv1.2 -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/$repo/releases/latest")
  version=${release_url##*/}
fi
printf '%s\n' "$version" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+([.-][A-Za-z0-9.-]+)?$' || { echo 'Cannot resolve a valid release tag.' >&2; exit 1; }
if [ "$(id -u)" = 0 ]; then default_dir=/usr/local/bin; else default_dir=$HOME/.local/bin; fi
dest=${AZ_INSTALL_DIR:-$default_dir}
case "$dest" in /*) ;; *) echo 'AZ_INSTALL_DIR must be absolute.' >&2; exit 1 ;; esac
work=$(mktemp -d)
staged=
cleanup() { rm -rf "$work"; if [ -n "$staged" ]; then rm -f "$staged"; fi; }
trap cleanup EXIT HUP INT TERM
asset=az-vote-delegate_${os}_${arch}.tar.gz
base=https://github.com/$repo/releases/download/$version
curl --proto '=https' --tlsv1.2 -fsSL "$base/$asset" -o "$work/$asset"
curl --proto '=https' --tlsv1.2 -fsSL "$base/checksums.txt" -o "$work/checksums.txt"
expected=$(awk -v name="$asset" '$2 == name { print $1 }' "$work/checksums.txt")
printf '%s\n' "$expected" | grep -Eq '^[0-9a-f]{64}$' || { echo 'Missing or invalid release checksum.' >&2; exit 1; }
printf '%s  %s\n' "$expected" "$asset" > "$work/verify.txt"
(cd "$work" && sha256sum --check --status verify.txt) || { echo 'Checksum verification failed. Nothing installed.' >&2; exit 1; }
tar -xzf "$work/$asset" -C "$work" az-vote-delegate
if [ ! -f "$work/az-vote-delegate" ] || [ -L "$work/az-vote-delegate" ]; then
  echo 'Invalid release archive.' >&2
  exit 1
fi
mkdir -p "$dest"
staged=$(mktemp "$dest/.az-vote-delegate.XXXXXX")
install -m 0755 "$work/az-vote-delegate" "$staged"
mv -f "$staged" "$dest/az-vote-delegate"
staged=
printf '\nInstalled az-vote-delegate %s to %s\n' "$version" "$dest/az-vote-delegate"
printf 'Run %s to open setup and the interactive menu.\n' "$dest/az-vote-delegate"
printf 'Run it as the user who owns your producer key (use sudo if your node runs as root).\n'
printf 'For an upgrade, restart the voting service after installation.\n'
