#!/usr/bin/env sh
set -eu

CLI_NAME="buda"
OWNER="CGuiho"
REPOSITORY="buda"
VERSION="${1:-latest}"
INSTALL_DIR="${GUIHO_INSTALL_DIR:-${HOME}/.local/bin}"

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'error: required command not found: %s\n' "$1" >&2
    exit 1
  fi
}

require_command curl
require_command unzip

case "$(uname -s)/$(uname -m)" in
  Linux/x86_64|Linux/amd64) ASSET="buda-linux-amd64" ;;
  Linux/aarch64|Linux/arm64) ASSET="buda-linux-arm64" ;;
  Linux/armv7l) ASSET="buda-linux-armv7" ;;
  Linux/armv6l) ASSET="buda-linux-armv6" ;;
  Darwin/x86_64) ASSET="buda-darwin-amd64" ;;
  Darwin/arm64) ASSET="buda-darwin-arm64" ;;
  *) printf 'error: unsupported platform: %s/%s\n' "$(uname -s)" "$(uname -m)" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  TAG="$(curl --fail --silent --show-error "https://api.github.com/repos/${OWNER}/${REPOSITORY}/releases/latest" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  if [ -z "$TAG" ]; then
    printf 'error: could not resolve latest Buda release tag\n' >&2
    exit 1
  fi
else
  # Non-latest values are exact full release tags; Buda owns no implicit prefix.
  TAG="$VERSION"
fi

BASE_URL="https://github.com/${OWNER}/${REPOSITORY}/releases/download/${TAG}"
SKILL_ASSET="guiho-s-0002-buda.zip"
TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT HUP INT TERM

printf '%s\n' 'Initiating Buda installation sequence...'
printf 'Target version: %s\nTarget asset: %s\nSource URL: %s/%s\n' "$TAG" "$ASSET" "$BASE_URL" "$ASSET"
curl --fail --location --progress-bar --output "${TEMP_DIR}/${ASSET}" "${BASE_URL}/${ASSET}"
curl --fail --location --progress-bar --output "${TEMP_DIR}/checksums.txt" "${BASE_URL}/checksums.txt"
curl --fail --location --progress-bar --output "${TEMP_DIR}/${SKILL_ASSET}" "${BASE_URL}/${SKILL_ASSET}"

EXPECTED_HASH="$(awk -v asset="$ASSET" '$2 == asset { print $1 }' "${TEMP_DIR}/checksums.txt")"
if [ -z "$EXPECTED_HASH" ]; then
  printf 'error: checksum entry missing for %s\n' "$ASSET" >&2
  exit 1
fi
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL_HASH="$(sha256sum "${TEMP_DIR}/${ASSET}" | awk '{ print $1 }')"
else
  require_command shasum
  ACTUAL_HASH="$(shasum -a 256 "${TEMP_DIR}/${ASSET}" | awk '{ print $1 }')"
fi
if [ "$EXPECTED_HASH" != "$ACTUAL_HASH" ]; then
  printf 'error: checksum verification failed for %s\n' "$ASSET" >&2
  exit 1
fi
printf '%s\n' '[OK] SHA-256 verification complete.'

mkdir -p "$INSTALL_DIR"
install -m 0755 "${TEMP_DIR}/${ASSET}" "${INSTALL_DIR}/${CLI_NAME}"
printf '[OK] Installed binary: %s/%s\n' "$INSTALL_DIR" "$CLI_NAME"

mkdir -p "${TEMP_DIR}/skill"
unzip -q "${TEMP_DIR}/${SKILL_ASSET}" -d "${TEMP_DIR}/skill"
SOURCE_SKILL="${TEMP_DIR}/skill/guiho-s-0002-buda"
if [ ! -f "${SOURCE_SKILL}/SKILL.md" ]; then
  printf 'error: skill archive does not contain guiho-s-0002-buda/SKILL.md\n' >&2
  exit 1
fi
for ROOT in "${HOME}/.agents/skills" "${HOME}/.claude/skills"; do
  mkdir -p "$ROOT"
  TARGET="${ROOT}/guiho-s-0002-buda"
  rm -rf "$TARGET"
  cp -R "$SOURCE_SKILL" "$TARGET"
  printf '[OK] Installed global Buda skill: %s\n' "$TARGET"
done

"${INSTALL_DIR}/${CLI_NAME}" --version
if ! command -v qmd >/dev/null 2>&1; then
  printf '%s\n' 'error: qmd is required but not installed. Install @tobilu/qmd, then run: buda doctor --wiki <path>' >&2
  exit 1
fi
qmd --version
printf '%s\n' '[OK] Buda installation complete. Repository instructions are installed only for an explicit --wiki path.'
