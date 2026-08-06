#!/usr/bin/env sh
# Buda canonical-tag installer (POSIX sh).
#
# Environment variables:
#   GUIHO_INSTALL_DIR      Binary install directory (default: ~/.local/bin).
#   BUDA_RELEASE_ASSET_DIR Load release assets from a local directory instead of
#                          GitHub; requires an exact canonical tag argument.
#   BUDA_SKILL_DIRS        Space-separated list of extra skill destination
#                          directories. Each is created if missing and the
#                          embedded skill directory is copied into it (the skill
#                          id subdirectory is appended automatically). Defaults
#                          to empty; the built-in ~/.agents/skills and
#                          ~/.claude/skills destinations are always installed by
#                          `buda agent skill update`.
#   HERMES_SKILLS_DIR      Hermes agent skills directory. When this directory
#                          exists (default: ~/.hermes/skills) the embedded skill
#                          is additionally registered there by the installer so
#                          Hermes agents can use Buda without a manual copy.
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

require_command unzip

if [ -z "${BUDA_RELEASE_ASSET_DIR:-}" ]; then
  require_command curl
fi

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
  if [ -n "${BUDA_RELEASE_ASSET_DIR:-}" ]; then
    printf '%s\n' 'error: BUDA_RELEASE_ASSET_DIR requires an exact canonical tag such as buda/v0.0.2' >&2
    exit 1
  fi
  TAG="$(curl --fail --silent --show-error "https://api.github.com/repos/${OWNER}/${REPOSITORY}/releases/latest" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  if [ -z "$TAG" ]; then
    printf 'error: could not resolve latest Buda release tag\n' >&2
    exit 1
  fi
else
  # Non-latest values are exact full release tags; Buda owns no implicit prefix.
  TAG="$VERSION"
fi

EXPECTED_VERSION="${TAG#buda/v}"
if [ "$EXPECTED_VERSION" = "$TAG" ] || ! printf '%s\n' "$EXPECTED_VERSION" | grep -Eq '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$'; then
  printf "error: invalid Buda release tag '%s'; expected buda/v<semver>\n" "$TAG" >&2
  exit 1
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

fetch_asset() {
  FETCH_NAME="$1"
  FETCH_DESTINATION="$2"
  if [ -n "${BUDA_RELEASE_ASSET_DIR:-}" ]; then
    SOURCE_PATH="${BUDA_RELEASE_ASSET_DIR}/${FETCH_NAME}"
    if [ ! -f "$SOURCE_PATH" ]; then
      printf 'error: local release asset not found: %s\n' "$SOURCE_PATH" >&2
      exit 1
    fi
    cp "$SOURCE_PATH" "$FETCH_DESTINATION"
  else
    curl --fail --location --progress-bar --output "$FETCH_DESTINATION" "${BASE_URL}/${FETCH_NAME}"
  fi
}

printf '%s\n' 'Initiating Buda installation sequence...'
if [ -n "${BUDA_RELEASE_ASSET_DIR:-}" ]; then
  printf 'Target version: %s\nTarget asset: %s\nSource directory: %s\n' "$TAG" "$ASSET" "$BUDA_RELEASE_ASSET_DIR"
else
  printf 'Target version: %s\nTarget asset: %s\nSource URL: %s/%s\n' "$TAG" "$ASSET" "$BASE_URL" "$ASSET"
fi
fetch_asset "$ASSET" "${TEMP_DIR}/${ASSET}"
fetch_asset "checksums.txt" "${TEMP_DIR}/checksums.txt"
fetch_asset "$SKILL_ASSET" "${TEMP_DIR}/${SKILL_ASSET}"

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
# Resolve qmd even when it is installed off PATH (common in agent-managed
# environments such as ~/.hermes/node/bin). Only hard-fail when truly absent.
QMD_BIN=""
QMD_NPX_FALLBACK=0
if command -v qmd >/dev/null 2>&1; then
  QMD_BIN="qmd"
else
  QMD_PROBE_DIRS=""
  if command -v npm >/dev/null 2>&1; then
    NPM_GLOBAL_BIN="$(npm root -g 2>/dev/null)/bin"
    [ -n "$NPM_GLOBAL_BIN" ] && QMD_PROBE_DIRS="$QMD_PROBE_DIRS $NPM_GLOBAL_BIN"
  fi
  QMD_PROBE_DIRS="$QMD_PROBE_DIRS ${HOME}/.hermes/node/bin ${HOME}/.local/bin ${HOME}/.npm-global/bin"
  for probe_dir in $QMD_PROBE_DIRS; do
    if [ -n "$probe_dir" ] && [ -x "$probe_dir/qmd" ]; then
      QMD_BIN="$probe_dir/qmd"
      break
    fi
  done
  if [ -z "$QMD_BIN" ] && command -v npx >/dev/null 2>&1; then
    if npx --no-install qmd --version >/dev/null 2>&1; then
      QMD_BIN="npx --no-install qmd"
      QMD_NPX_FALLBACK=1
    fi
  fi
  if [ -z "$QMD_BIN" ]; then
    printf '%s\n' 'error: qmd is required but not installed. Install @tobilu/qmd@2.5.3, then run: buda doctor --wiki <path>' >&2
    exit 1
  fi
  if [ "$QMD_NPX_FALLBACK" -eq 1 ]; then
    # The npx fallback is not a filesystem path, so directory-to-PATH advice
    # does not apply; suggest a global install instead.
    printf '%s\n' 'info: qmd found via the npx fallback (npx --no-install qmd). For reliable operation install it globally: npm install -g @tobilu/qmd@2.5.3. Then run: buda doctor --wiki <path>' >&2
  else
    printf 'warning: qmd was not on PATH but was found at %s; add its directory to PATH (e.g. export PATH="%s:$PATH") for reliable operation.\n' "$QMD_BIN" "${QMD_BIN%/*}" >&2
  fi
fi
QMD_VERSION_OUTPUT="$($QMD_BIN --version)"
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
if [ "$INSTALLED_VERSION" != "buda v${EXPECTED_VERSION}" ]; then
  printf "error: installed version '%s' does not match requested tag '%s'\n" "$INSTALLED_VERSION" "$TAG" >&2
  exit 1
fi
printf '%s\n' "$INSTALLED_VERSION"
# The binary is verified and installed; mark the install successful before the
# optional skill-destination registrations below so an optional registration
# failure can never roll back a valid binary replacement.
INSTALL_SUCCEEDED=1

# Register the embedded skill into optional destinations: BUDA_SKILL_DIRS
# (space-separated) and the Hermes skills directory when it exists. The
# built-in ~/.agents/skills and ~/.claude/skills destinations are already
# handled by `buda agent skill update` above; these are additive and
# non-fatal: a failure only warns and leaves the installation complete.
register_skill_dir() {
  rs_target="$1"
  if [ -z "$rs_target" ]; then
    return 0
  fi
  rs_dest="$rs_target/guiho-s-0002-buda"
  if ! mkdir -p "$rs_target" 2>/dev/null; then
    printf 'warning: could not create %s; skipping optional Buda skill registration (install remains complete).\n' "$rs_target" >&2
    return 0
  fi
  if ! rm -rf "$rs_dest" 2>/dev/null; then
    printf 'warning: could not refresh %s; skipping optional Buda skill registration (install remains complete).\n' "$rs_dest" >&2
    return 0
  fi
  if ! cp -R "$SOURCE_SKILL" "$rs_dest" 2>/dev/null; then
    printf 'warning: could not copy the Buda skill to %s; skipping optional Buda skill registration (install remains complete).\n' "$rs_dest" >&2
    return 0
  fi
  printf '[OK] Registered Buda skill: %s\n' "$rs_dest"
}

for skill_dir in ${BUDA_SKILL_DIRS:-}; do
  register_skill_dir "$skill_dir"
done

HERMES_SKILLS_DIR_RESOLVED="${HERMES_SKILLS_DIR:-${HOME}/.hermes/skills}"
if [ -d "$HERMES_SKILLS_DIR_RESOLVED" ]; then
  register_skill_dir "$HERMES_SKILLS_DIR_RESOLVED"
fi

# Verify the binary is callable on PATH; warn (not fail) with an actionable
# fix when the install directory is not on PATH for the current shell.
if ! command -v "$CLI_NAME" >/dev/null 2>&1; then
  printf 'warning: %s was installed to %s which is not on PATH for this shell.\n' "$CLI_NAME" "$INSTALL_DIR" >&2
  printf '         Fix one of:\n' >&2
  printf '           export PATH="%s:$PATH"\n' "$INSTALL_DIR" >&2
  printf '           ln -s "%s/%s" /usr/local/bin/%s\n' "$INSTALL_DIR" "$CLI_NAME" "$CLI_NAME" >&2
  printf '         Then run: buda doctor --wiki <path>\n' >&2
fi

printf '%s\n' '[OK] Buda installation complete. Repository instructions are installed only for an explicit --wiki path.'
