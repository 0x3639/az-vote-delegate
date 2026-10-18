#!/bin/sh
# Exercises the actual installer without network, root or modifying PATH dirs.
set -eu
root=$(pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT HUP INT TERM
mkdir -p "$work/mocks" "$work/assets" "$work/package" "$work/bin"
cat > "$work/package/az-vote-delegate" <<'MOCK'
#!/bin/sh
set -eu
if [ -n "${TEST_UI_MARKER:-}" ]; then
  test -t 0
  test -t 2
  printf '%s\n' "$*" > "$TEST_UI_MARKER"
fi
echo 'az-vote-delegate test'
MOCK
chmod +x "$work/package/az-vote-delegate"
tar -C "$work/package" -czf "$work/assets/az-vote-delegate_linux_amd64.tar.gz" az-vote-delegate
(cd "$work/assets" && sha256sum -- *.tar.gz > checksums.txt)
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
cat > "$work/mocks/systemctl" <<'MOCK'
#!/bin/sh
# Installing the binary must not create, start, or restart a service.
printf '%s\n' "$*" >> "$TEST_SERVICE_CALLS"
exit 1
MOCK
chmod +x "$work/mocks/"*
export TEST_ASSETS="$work/assets"
export TEST_SERVICE_CALLS="$work/service-calls"
PATH="$work/mocks:$PATH" AZ_NO_UI=1 AZ_VERSION=v0.1.0 AZ_INSTALL_DIR="$work/bin" sh "$root/install.sh"
[ -x "$work/bin/az-vote-delegate" ]
# A repeat install replaces the old file at the same path and opens the UI,
# which can reconnect to the existing service instead of creating another one.
printf 'old executable\n' > "$work/bin/az-vote-delegate"
PATH="$work/mocks:$PATH" AZ_NO_UI=1 AZ_VERSION=v0.1.0 AZ_INSTALL_DIR="$work/bin" sh "$root/install.sh" > "$work/reinstall-output"
grep -q 'Existing installation detected' "$work/reinstall-output"
cmp "$work/package/az-vote-delegate" "$work/bin/az-vote-delegate"
[ "$(find "$work/bin" -type f | wc -l | tr -d ' ')" = 1 ]
TEST_INSTALLER="$root/install.sh" PATH="$work/mocks:$PATH" AZ_VERSION=v0.1.0 AZ_INSTALL_DIR="$work/bin" python3 scripts/test-install-tty.py "$work"
printf 'existing binary\n' > "$work/bin/az-vote-delegate"
printf 'corrupted\n' >> "$work/assets/az-vote-delegate_linux_amd64.tar.gz"
if PATH="$work/mocks:$PATH" AZ_VERSION=v0.1.0 AZ_INSTALL_DIR="$work/bin" sh "$root/install.sh"; then echo 'Corrupt archive was accepted' >&2; exit 1; fi
[ "$(cat "$work/bin/az-vote-delegate")" = 'existing binary' ]
[ ! -e "$TEST_SERVICE_CALLS" ]
echo 'Installer, repeat-install replacement, no service creation, and checksum-failure checks passed.'
