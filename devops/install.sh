#!/usr/bin/env sh
set -eu

OWNER=CGuiho
REPOSITORY=buda
VERSION=""
CHANNEL=""
WIKI=""
WIKI_ID=""
ASSET_DIR="${BUDA_RELEASE_ASSET_DIR:-}"

usage() { printf '%s\n' "Usage: install.sh --wiki <path> [--wiki-id <id>] [--version <semver> | --channel <channel>]"; }
semver_gt() {
  awk -v left="$1" -v right="$2" '
    function compare(a, b,    ap, bp, ac, bc, i, ai, bi, ar, br, an, bn, na, nb, np, nq) {
      sub(/\+.*/, "", a); sub(/\+.*/, "", b)
      np = split(a, ap, "-"); nq = split(b, bp, "-")
      split(ap[1], ac, "\\."); split(bp[1], bc, "\\.")
      for (i = 1; i <= 3; i++) {
        if ((ac[i] + 0) != (bc[i] + 0)) return (ac[i] + 0) > (bc[i] + 0)
      }
      ar = (np > 1 ? ap[2] : ""); br = (nq > 1 ? bp[2] : "")
      if (ar == br) return 0
      if (ar == "") return 1
      if (br == "") return 0
      na = split(ar, an, "\\."); nb = split(br, bn, "\\.")
      for (i = 1; i <= (na > nb ? na : nb); i++) {
        if (!(i in an)) return 0
        if (!(i in bn)) return 1
        ai = an[i]; bi = bn[i]
        if (ai == bi) continue
        if (ai ~ /^[0-9]+$/ && bi ~ /^[0-9]+$/) return (ai + 0) > (bi + 0)
        if (ai ~ /^[0-9]+$/) return 1
        if (bi ~ /^[0-9]+$/) return 0
        return ai > bi
      }
      return na > nb
    }
    BEGIN { print compare(left, right) ? "1" : "0" }
  '
}
while [ "$#" -gt 0 ]; do
  case "$1" in
    --version) [ "$#" -ge 2 ] || { usage >&2; exit 2; }; VERSION=$2; shift 2 ;;
    --channel) [ "$#" -ge 2 ] || { usage >&2; exit 2; }; CHANNEL=$2; shift 2 ;;
    --wiki) [ "$#" -ge 2 ] || { usage >&2; exit 2; }; WIKI=$2; shift 2 ;;
    --wiki-id) [ "$#" -ge 2 ] || { usage >&2; exit 2; }; WIKI_ID=$2; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) printf 'unknown option: %s\n' "$1" >&2; usage >&2; exit 2 ;;
  esac
done
[ -n "$WIKI" ] || { printf '%s\n' '--wiki is required; Buda never selects a wiki implicitly' >&2; exit 2; }
[ -z "$VERSION" ] || [ -z "$CHANNEL" ] || { printf '%s\n' '--version and --channel are mutually exclusive' >&2; exit 2; }
case "$VERSION" in buda/*) VERSION=${VERSION#buda/} ;; esac
case "$VERSION" in v*) VERSION=${VERSION#v} ;; esac
[ -z "$VERSION" ] || printf '%s' "$VERSION" | grep -Eq '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$' || { printf 'invalid --version: %s\n' "$VERSION" >&2; exit 2; }

case "$(uname -s):$(uname -m)" in
  Linux:x86_64|Linux:amd64) OS=linux; ARCH=amd64; BINARY=buda-linux-amd64; LAUNCHER=buda-launcher-linux-amd64 ;;
  Linux:aarch64|Linux:arm64) OS=linux; ARCH=arm64; BINARY=buda-linux-arm64; LAUNCHER=buda-launcher-linux-arm64 ;;
  Linux:armv7l|Linux:armv7|Linux:armhf) OS=linux; ARCH=armv7; BINARY=buda-linux-armv7; LAUNCHER=buda-launcher-linux-armv7 ;;
  Linux:armv6l|Linux:armv6) OS=linux; ARCH=armv6; BINARY=buda-linux-armv6; LAUNCHER=buda-launcher-linux-armv6 ;;
  Darwin:x86_64|Darwin:amd64) OS=darwin; ARCH=amd64; BINARY=buda-darwin-amd64; LAUNCHER=buda-launcher-darwin-amd64 ;;
  Darwin:arm64) OS=darwin; ARCH=arm64; BINARY=buda-darwin-arm64; LAUNCHER=buda-launcher-darwin-arm64 ;;
  *) printf 'unsupported platform: %s:%s\n' "$(uname -s)" "$(uname -m)" >&2; exit 2 ;;
esac

HOME_DIR=${HOME:?HOME is required}
GUIHO_HOME=$HOME_DIR/.guiho
CLI_HOME=$GUIHO_HOME/buda
GLOBAL_CONFIG_NAME=buda.global.yaml
BIN_DIR=$GUIHO_HOME/bin
TEMP_ROOT=$GUIHO_HOME/.temp
mkdir -p "$TEMP_ROOT"
STAGE=$(mktemp -d "$TEMP_ROOT/buda-install-XXXXXX")

if [ -n "$ASSET_DIR" ]; then
  [ -n "$VERSION" ] || VERSION=${BUDA_RELEASE_VERSION:-}
  [ -n "$VERSION" ] || { printf '%s\n' 'BUDA_RELEASE_VERSION or --version is required with BUDA_RELEASE_ASSET_DIR' >&2; exit 2; }
  TAG="buda/v$VERSION"
  source_asset() { cp "$ASSET_DIR/$1" "$2"; }
else
  if [ -z "$VERSION" ]; then
    command -v jq >/dev/null 2>&1 || { printf '%s\n' 'jq is required for paginated release discovery' >&2; exit 1; }
    wanted=${CHANNEL:-stable}; page=1; selected_version=""
    while :; do
      releases=$(curl -fsSL "https://api.github.com/repos/$OWNER/$REPOSITORY/releases?per_page=100&page=$page")
      count=$(printf '%s' "$releases" | jq 'length')
    candidates=$(printf '%s' "$releases" | jq -r --arg wanted "$wanted" --arg binary "$BINARY" --arg launcher "$LAUNCHER" '.[] | select(.draft|not) | select(.tag_name|test("^buda/v")) | ((.assets // []) as $raw | select(((([$binary,$launcher,"checksums.txt","artifacts.json","guiho-s-0002-buda.zip","guiho-i-buda.md","guiho-p-buda.md","buda.schema.json","buda.global.schema.json","buda.example.yaml","buda.global.example.yaml","buda-linux-amd64","buda-linux-arm64","buda-linux-armv7","buda-linux-armv6","buda-darwin-amd64","buda-darwin-arm64","buda-windows-amd64.exe","buda-windows-arm64.exe","buda-launcher-linux-amd64","buda-launcher-linux-arm64","buda-launcher-linux-armv7","buda-launcher-linux-armv6","buda-launcher-darwin-amd64","buda-launcher-darwin-arm64","buda-launcher-windows-amd64.exe","buda-launcher-windows-arm64.exe"] - ($raw | map(.name) | unique)) | length) == 0) and (($raw | map(.name) | length) == ($raw | map(.name) | unique | length)))) | (.tag_name|sub("^buda/v";"")) as $v | select($v|test("^(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\\+[0-9A-Za-z.-]+)?$")) | select((if ($v|contains("-")) then ($v|split("-")[1]|split(".")[0]) else "stable" end)==$wanted) | $v')
      for candidate in $candidates; do
        if [ -z "$selected_version" ] || [ "$(semver_gt "$candidate" "$selected_version")" = 1 ]; then selected_version=$candidate; fi
      done
      [ "$count" -lt 100 ] && break
      page=$((page + 1))
    done
    [ -n "$selected_version" ] || { printf 'no release found for channel %s\n' "$wanted" >&2; exit 1; }
    VERSION=$selected_version; TAG="buda/v$VERSION"
  else
    TAG="buda/v$VERSION"
  fi
  BASE="https://github.com/$OWNER/$REPOSITORY/releases/download/$TAG"
  source_asset() { curl -fsSL "$BASE/$1" -o "$2"; }
fi

cleanup() { status=$?; [ "$status" -eq 0 ] || rollback 2>/dev/null || true; rm -rf "$STAGE"; exit "$status"; }
trap cleanup EXIT INT TERM

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | awk '{print $1}'; return; fi
  if command -v shasum >/dev/null 2>&1; then shasum -a 256 "$1" | awk '{print $1}'; return; fi
  printf '%s\n' 'sha256sum or shasum is required for release verification' >&2
  return 1
}
manifest_paths() { awk -F'"' '/"path"[[:space:]]*:/ { print $4 }' "$STAGE/artifacts.json"; }
manifest_has() { manifest_paths | grep -F -x "$1" >/dev/null 2>&1; }
verify_checksum() {
  name=$1
  expected=$(awk -v n="$name" '$2==n || $2=="*" n { if (++count > 1) exit 9; print $1 } END { if (count != 1) exit 8 }' "$STAGE/checksums.txt") || { printf 'missing or duplicate checksum: %s\n' "$name" >&2; return 1; }
  actual=$(sha256 "$STAGE/$name")
  [ "$expected" = "$actual" ] || { printf 'checksum mismatch: %s\n' "$name" >&2; return 1; }
}

printf 'Resolved Buda %s for %s/%s\n' "$VERSION" "$OS" "$ARCH"
printf 'CLI home: %s\nStable launcher: %s/buda\n' "$CLI_HOME" "$BIN_DIR"

# Fetch the ownership manifest and checksum index first, then fetch exactly
# the remaining manifest-declared assets. Filename prefixes are not ownership.
source_asset artifacts.json "$STAGE/artifacts.json"
source_asset checksums.txt "$STAGE/checksums.txt"
manifest_schema=$(sed -n 's/^[[:space:]]*"schema":[[:space:]]*\([0-9][0-9]*\).*/\1/p' "$STAGE/artifacts.json" | head -n 1)
manifest_cli=$(sed -n 's/^[[:space:]]*"cli":[[:space:]]*"\([^"]*\)".*/\1/p' "$STAGE/artifacts.json" | head -n 1)
manifest_version=$(sed -n 's/^[[:space:]]*"version":[[:space:]]*"\([^"]*\)".*/\1/p' "$STAGE/artifacts.json" | head -n 1)
[ "$manifest_schema" = 1 ] || { printf '%s\n' 'release artifacts.json has an unsupported schema' >&2; exit 1; }
[ "$manifest_cli" = buda ] || { printf '%s\n' 'release artifacts.json belongs to another CLI' >&2; exit 1; }
[ "$manifest_version" = "$VERSION" ] || { printf 'release manifest version mismatch: expected %s, got %s\n' "$VERSION" "$manifest_version" >&2; exit 1; }
manifest_count=$(manifest_paths | wc -l | tr -d ' ')
manifest_unique=$(manifest_paths | sort -u | wc -l | tr -d ' ')
[ "$manifest_count" = "$manifest_unique" ] || { printf '%s\n' 'manifest contains duplicate asset paths' >&2; exit 1; }
for name in $(manifest_paths); do
  case "$name" in */*|*\\*|..|../*) printf 'unsafe manifest asset path: %s\n' "$name" >&2; exit 1 ;; esac
  [ -n "$name" ] || { printf '%s\n' 'manifest contains an empty asset path' >&2; exit 1; }
  [ -f "$STAGE/$name" ] || source_asset "$name" "$STAGE/$name"
done
for name in "$BINARY" "$LAUNCHER" guiho-s-0002-buda.zip guiho-i-buda.md guiho-p-buda.md buda.schema.json buda.global.schema.json buda.example.yaml buda.global.example.yaml artifacts.json; do
  manifest_has "$name" || { printf 'manifest does not declare required asset: %s\n' "$name" >&2; exit 1; }
done
for name in $(manifest_paths); do verify_checksum "$name"; done
chmod 755 "$STAGE/$BINARY" "$STAGE/$LAUNCHER"
while read -r digest name extra; do
  [ -n "${name:-}" ] || continue
  [ -z "${extra:-}" ] || { printf 'malformed checksum entry: %s\n' "$name" >&2; exit 1; }
  name=${name#*\*}
  manifest_has "$name" || { printf 'checksum names undeclared asset: %s\n' "$name" >&2; exit 1; }
done < "$STAGE/checksums.txt"
observed=$($STAGE/$BINARY --version 2>/dev/null || true)
[ "$observed" = "$VERSION" ] || { printf 'candidate version mismatch: expected %s, got %s\n' "$VERSION" "$observed" >&2; exit 1; }
self_test=$($STAGE/$BINARY __self-test 2>/dev/null || true)
[ "$self_test" = "ok" ] || { printf '%s\n' 'candidate self-test failed' >&2; exit 1; }

VERSION_DIR=$CLI_HOME/versions/$VERSION
BACKUP_DIR=$STAGE/backup
mkdir -p "$BACKUP_DIR" "$CLI_HOME/versions" "$BIN_DIR"
# The historical 0.1.x direct-binary installation wrote the payload to
# ~/.local/bin/buda by default (or $GUIHO_INSTALL_DIR/buda when overridden).
# Migration removes it only from that exact historical path, only after the
# new launcher transaction has been verified, and never in place.
LEGACY_PATH=$HOME_DIR/.local/bin/buda
if [ -n "${GUIHO_INSTALL_DIR:-}" ]; then LEGACY_PATH=$GUIHO_INSTALL_DIR/buda; fi
had_legacy=0
if [ -f "$LEGACY_PATH" ]; then
  legacy_version=$($LEGACY_PATH --version 2>/dev/null || true)
  if printf '%s' "$legacy_version" | grep -Eq '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$'; then
    cp -p "$LEGACY_PATH" "$BACKUP_DIR/legacy-buda"
    had_legacy=1
  fi
fi
previous=""
if [ -f "$CLI_HOME/current.json" ]; then previous=$(sed -n 's/.*"active":"\([^"]*\)".*/\1/p' "$CLI_HOME/current.json" | head -n 1); fi
previous_version="${previous%%/*}"
had_current=0; had_manifest=0; had_launcher=0; had_version=0; mutated=0; committed=0
if [ -f "$CLI_HOME/current.json" ]; then cp -p "$CLI_HOME/current.json" "$BACKUP_DIR/current.json"; had_current=1; fi
if [ -f "$CLI_HOME/installed-artifacts.json" ]; then cp -p "$CLI_HOME/installed-artifacts.json" "$BACKUP_DIR/installed-artifacts.json"; had_manifest=1; fi
if [ -f "$BIN_DIR/buda" ]; then cp -p "$BIN_DIR/buda" "$BACKUP_DIR/buda"; had_launcher=1; fi
mutated=1
if [ -e "$VERSION_DIR" ]; then mv "$VERSION_DIR" "$BACKUP_DIR/version"; had_version=1; fi
rollback() {
  [ "$mutated" -eq 1 ] || return 0
  rm -rf "$VERSION_DIR" 2>/dev/null || true
  [ "$had_version" -eq 1 ] && mv "$BACKUP_DIR/version" "$VERSION_DIR" 2>/dev/null || true
  if [ "$had_current" -eq 1 ]; then cp -p "$BACKUP_DIR/current.json" "$CLI_HOME/current.json" 2>/dev/null || true; else rm -f "$CLI_HOME/current.json" 2>/dev/null || true; fi
  if [ "$had_manifest" -eq 1 ]; then cp -p "$BACKUP_DIR/installed-artifacts.json" "$CLI_HOME/installed-artifacts.json" 2>/dev/null || true; else rm -f "$CLI_HOME/installed-artifacts.json" 2>/dev/null || true; fi
  if [ "$had_launcher" -eq 1 ]; then cp -p "$BACKUP_DIR/buda" "$BIN_DIR/buda" 2>/dev/null || true; else rm -f "$BIN_DIR/buda" 2>/dev/null || true; fi
  if [ "$had_legacy" -eq 1 ]; then cp -p "$BACKUP_DIR/legacy-buda" "$LEGACY_PATH" 2>/dev/null || true; fi
}

mkdir -p "$VERSION_DIR/artifacts"
cp "$STAGE/$BINARY" "$VERSION_DIR/buda"
chmod 755 "$VERSION_DIR/buda"
for name in $(manifest_paths); do cp "$STAGE/$name" "$VERSION_DIR/artifacts/$name"; done
cp "$STAGE/$LAUNCHER" "$BIN_DIR/.buda-launcher-new"
chmod 755 "$BIN_DIR/.buda-launcher-new"
mv -f "$BIN_DIR/.buda-launcher-new" "$BIN_DIR/buda"
mkdir -p "$CLI_HOME/state"
printf '{"schema":1,"active":"%s/buda","previous":"%s","active_version":"%s","previous_version":"%s"}\n' "$VERSION" "$previous" "$VERSION" "$previous_version" > "$CLI_HOME/.current.json.new"
mv -f "$CLI_HOME/.current.json.new" "$CLI_HOME/current.json"
cp "$STAGE/artifacts.json" "$CLI_HOME/.installed-artifacts.json.new"
mv -f "$CLI_HOME/.installed-artifacts.json.new" "$CLI_HOME/installed-artifacts.json"
committed=1
observed=$($BIN_DIR/buda --version 2>/dev/null || true)
[ "$observed" = "$VERSION" ] || { printf 'launcher version mismatch: expected %s, got %s\n' "$VERSION" "$observed" >&2; exit 1; }
self_test=$($BIN_DIR/buda __self-test 2>/dev/null || true)
[ "$self_test" = "ok" ] || { printf '%s\n' 'launcher self-test failed' >&2; exit 1; }
case ":${PATH:-}:" in
  *":$BIN_DIR:"*) ;;
  *) if [ ! -f "$HOME_DIR/.profile" ] || ! grep -F '$HOME/.guiho/bin' "$HOME_DIR/.profile" >/dev/null 2>&1; then printf '\n# GUIHO Buda\nexport PATH="$HOME/.guiho/bin:$PATH"\n' >> "$HOME_DIR/.profile"; fi ;;
esac
# Invoke the required explicit-wiki init without shell evaluation so paths
# remain opaque to the installer.
if [ -n "$WIKI_ID" ]; then "$BIN_DIR/buda" init --wiki "$WIKI" --wiki-id "$WIKI_ID"; else "$BIN_DIR/buda" init --wiki "$WIKI"; fi || { printf '%s\n' 'Buda installed, but explicit-wiki init failed; rolling back lifecycle files.' >&2; exit 1; }
if [ "$had_legacy" -eq 1 ]; then rm -f "$LEGACY_PATH"; fi
printf 'Installed Buda %s\nLauncher: %s/buda\nPayload: %s\nCLI home: %s\n' "$VERSION" "$BIN_DIR" "$VERSION_DIR/buda" "$CLI_HOME"
