#!/bin/sh
# Exercises the actual installer without network, root or modifying PATH dirs.
set -eu
root=$(pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT HUP INT TERM
mkdir -p "$work/mocks" "$work/assets" "$work/package" "$work/bin"
printf '#!/bin/sh\necho "az-vote-delegate test"\n' > "$work/package/az-vote-delegate"
chmod +x "$work/package/az-vote-delegate"
tar -C "$work/package" -czf "$work/assets/az-vote-delegate_linux_amd64.tar.gz" az-vote-delegate
(cd "$work/assets" && sha256sum *.tar.gz > checksums.txt)
cat > "$work/mocks/uname" <<'MOCK'
#!/bin/sh
case "$1" in -s) echo Linux ;; -m) echo x86_64 ;; esac
MOCK
cat > "$work/mocks/curl" <<'MOCK'
#!/bin/sh
set -eu
url= out=
while [ "$#" -gt 0 ]; do
 case "$1" in https://*) url=$1 ;; -o) shift; out=$1 ;; esac
 shift
done
cp "$TEST_ASSETS/${url##*/}" "$out"
MOCK
chmod +x "$work/mocks/"*
export TEST_ASSETS="$work/assets"
PATH="$work/mocks:$PATH" AZ_VERSION=v0.1.0 AZ_INSTALL_DIR="$work/bin" sh "$root/install.sh"
[ -x "$work/bin/az-vote-delegate" ]
printf 'existing binary\n' > "$work/bin/az-vote-delegate"
printf 'corrupted\n' >> "$work/assets/az-vote-delegate_linux_amd64.tar.gz"
if PATH="$work/mocks:$PATH" AZ_VERSION=v0.1.0 AZ_INSTALL_DIR="$work/bin" sh "$root/install.sh"; then echo 'Corrupt archive was accepted' >&2; exit 1; fi
[ "$(cat "$work/bin/az-vote-delegate")" = 'existing binary' ]
echo 'Installer success and checksum-failure checks passed.'
