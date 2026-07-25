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
DESTINATION="${INSTALL_DIR}/${CLI_NAME}"
BACKUP_PATH=""
BINARY_REPLACED=0
INSTALL_SUCCEEDED=0
cleanup() {
  if [ "$BINARY_REPLACED" -eq 1 ]; then
    if [ "$INSTALL_SUCCEEDED" -eq 1 ]; then
      [ -z "$BACKUP_PATH" ] || rm -f "$BACKUP_PATH"
    else
      rm -f "$DESTINATION"
      [ -z "$BACKUP_PATH" ] || mv "$BACKUP_PATH" "$DESTINATION"
    fi
  fi
  rm -rf "$TEMP_DIR"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM

printf '%s\n' 'Initiating Buda installation sequence...'
printf 'Target version: %s\nTarget asset: %s\nSource URL: %s/%s\n' "$TAG" "$ASSET" "$BASE_URL" "$ASSET"
curl --fail --location --progress-bar --output "${TEMP_DIR}/${ASSET}" "${BASE_URL}/${ASSET}"
curl --fail --location --progress-bar --output "${TEMP_DIR}/checksums.txt" "${BASE_URL}/checksums.txt"
curl --fail --location --progress-bar --output "${TEMP_DIR}/${SKILL_ASSET}" "${BASE_URL}/${SKILL_ASSET}"

verify_asset() {
  VERIFY_NAME="$1"
  EXPECTED_HASH="$(awk -v asset="$VERIFY_NAME" '$2 == asset { print $1 }' "${TEMP_DIR}/checksums.txt")"
  if [ -z "$EXPECTED_HASH" ]; then
    printf 'error: checksum entry missing for %s\n' "$VERIFY_NAME" >&2
    exit 1
  fi
  if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL_HASH="$(sha256sum "${TEMP_DIR}/${VERIFY_NAME}" | awk '{ print $1 }')"
  else
    require_command shasum
    ACTUAL_HASH="$(shasum -a 256 "${TEMP_DIR}/${VERIFY_NAME}" | awk '{ print $1 }')"
  fi
  if [ "$EXPECTED_HASH" != "$ACTUAL_HASH" ]; then
    printf 'error: checksum verification failed for %s\n' "$VERIFY_NAME" >&2
    exit 1
  fi
}
verify_asset "$ASSET"
verify_asset "$SKILL_ASSET"
printf '%s\n' '[OK] SHA-256 verification complete for binary and skill archive.'

mkdir -p "${TEMP_DIR}/skill"
unzip -q "${TEMP_DIR}/${SKILL_ASSET}" -d "${TEMP_DIR}/skill"
SOURCE_SKILL="${TEMP_DIR}/skill/guiho-s-0002-buda"
if [ ! -f "${SOURCE_SKILL}/SKILL.md" ]; then
  printf 'error: skill archive does not contain guiho-s-0002-buda/SKILL.md\n' >&2
  exit 1
fi
if ! command -v qmd >/dev/null 2>&1; then
  printf '%s\n' 'error: qmd is required but not installed. Install @tobilu/qmd, then run: buda doctor --wiki <path>' >&2
  exit 1
fi
QMD_VERSION_OUTPUT="$(qmd --version)"
QMD_VERSION="$(printf '%s\n' "$QMD_VERSION_OUTPUT" | awk '{ for (i = 1; i <= NF; i++) if ($i ~ /^v?[0-9]+\.[0-9]+\.[0-9]+/) { sub(/^v/, "", $i); print $i; exit } }')"
QMD_MAJOR="${QMD_VERSION%%.*}"
QMD_REST="${QMD_VERSION#*.}"
QMD_MINOR="${QMD_REST%%.*}"
if [ -z "$QMD_VERSION" ] || [ "$QMD_MAJOR" -ne 2 ] || [ "$QMD_MINOR" -lt 5 ]; then
  printf "error: unsupported qmd version '%s'; Buda requires >=2.5.0,<3.0.0\n" "$QMD_VERSION_OUTPUT" >&2
  exit 1
fi
printf '%s\n' "$QMD_VERSION_OUTPUT"

mkdir -p "$INSTALL_DIR"
STAGED_BINARY="${INSTALL_DIR}/.buda-new-$$"
install -m 0755 "${TEMP_DIR}/${ASSET}" "$STAGED_BINARY"
if [ -e "$DESTINATION" ]; then
  BACKUP_PATH="${INSTALL_DIR}/.buda-backup-$$"
  mv "$DESTINATION" "$BACKUP_PATH"
fi
BINARY_REPLACED=1
mv "$STAGED_BINARY" "$DESTINATION"
printf '[OK] Installed binary: %s\n' "$DESTINATION"

"$DESTINATION" agent skill update
printf '%s\n' '[OK] Installed both global Buda skill destinations transactionally from embedded resources.'

INSTALLED_VERSION="$("$DESTINATION" --version)"
EXPECTED_VERSION="${TAG#v}"
if [ "$INSTALLED_VERSION" != "buda v${EXPECTED_VERSION}" ]; then
  printf "error: installed version '%s' does not match requested tag '%s'\n" "$INSTALLED_VERSION" "$TAG" >&2
  exit 1
fi
printf '%s\n' "$INSTALLED_VERSION"
INSTALL_SUCCEEDED=1
printf '%s\n' '[OK] Buda installation complete. Repository instructions are installed only for an explicit --wiki path.'
