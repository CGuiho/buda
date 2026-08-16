#!/usr/bin/env sh
set -eu
PRESERVE_CONFIG=0
PRESERVE_DATA=0
DRY_RUN=0
YES=0
WIKI=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --preserve-config) PRESERVE_CONFIG=1; shift ;;
    --preserve-data) PRESERVE_DATA=1; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    --yes) YES=1; shift ;;
    --wiki) [ "$#" -ge 2 ] || exit 2; WIKI=$2; shift 2 ;;
    -h|--help) printf '%s\n' 'Usage: uninstall.sh [--wiki <path>] [--preserve-config] [--preserve-data] [--dry-run] [--yes]'; exit 0 ;;
    *) printf 'unknown option: %s\n' "$1" >&2; exit 2 ;;
  esac
done
HOME_DIR=${HOME:?HOME is required}
GUIHO_HOME=$HOME_DIR/.guiho
CLI_HOME=$GUIHO_HOME/buda
BIN_DIR=$GUIHO_HOME/bin
TEMP_ROOT=$GUIHO_HOME/.temp

# The installed Cobra command is the canonical manifest-driven planner. The
# script remains usable for recovery when the launcher is missing or corrupt.
if [ -x "$BIN_DIR/buda" ] && [ "${BUDA_UNINSTALL_FALLBACK:-0}" != "1" ]; then
  set -- uninstall --yes
  [ "$PRESERVE_CONFIG" -eq 1 ] && set -- "$@" --preserve-config
  [ "$PRESERVE_DATA" -eq 1 ] && set -- "$@" --preserve-data
  [ "$DRY_RUN" -eq 1 ] && set -- "$@" --dry-run
  [ -n "$WIKI" ] && set -- "$@" --wiki "$WIKI"
  exec "$BIN_DIR/buda" "$@"
fi

# If the stable launcher is missing but the active immutable payload remains,
# let that payload execute the same Go ownership planner. This recovery route
# keeps the shell fallback from guessing ownership or recursively removing the
# whole CLI home.
if [ "${BUDA_UNINSTALL_FALLBACK:-0}" != "1" ] && [ ! -x "$BIN_DIR/buda" ] && [ -f "$CLI_HOME/current.json" ]; then
  active=$(sed -n 's/.*"active"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$CLI_HOME/current.json" | head -n 1)
  case "$active" in
    ''|/*|../*|*/../*|*\\*) ;;
    *)
      payload="$CLI_HOME/versions/$active"
      if [ -x "$payload" ]; then
        set -- uninstall --yes
        [ "$PRESERVE_CONFIG" -eq 1 ] && set -- "$@" --preserve-config
        [ "$PRESERVE_DATA" -eq 1 ] && set -- "$@" --preserve-data
        [ "$DRY_RUN" -eq 1 ] && set -- "$@" --dry-run
        [ -n "$WIKI" ] && set -- "$@" --wiki "$WIKI"
        exec "$payload" "$@"
      fi
      ;;
  esac
fi

if [ -e "$CLI_HOME" ]; then
  [ -f "$CLI_HOME/installed-artifacts.json" ] || { printf '%s\n' 'refusing fallback uninstall without Buda ownership manifest' >&2; exit 1; }
  grep -Eq '"schema"[[:space:]]*:[[:space:]]*1' "$CLI_HOME/installed-artifacts.json" || { printf '%s\n' 'refusing fallback uninstall with an invalid Buda manifest' >&2; exit 1; }
  grep -Eq '"cli"[[:space:]]*:[[:space:]]*"buda"' "$CLI_HOME/installed-artifacts.json" || { printf '%s\n' 'refusing fallback uninstall for a foreign manifest' >&2; exit 1; }
  printf '%s\n' 'refusing fallback uninstall: the installed Buda payload is unavailable; no ownership-safe executor remains' >&2
  exit 1
fi

printf 'REMOVE %s\n' "$BIN_DIR/buda"
printf 'REMOVE %s\n' "$HOME_DIR/.agents/skills/guiho-s-0002-buda"
printf 'REMOVE %s\n' "$HOME_DIR/.claude/skills/guiho-s-0002-buda"
printf 'PRESERVE %s\n' "$TEMP_ROOT"
if [ "$PRESERVE_DATA" -eq 1 ]; then printf 'PRESERVE %s\n' "$CLI_HOME"; else printf 'REMOVE %s\n' "$CLI_HOME"; fi
if [ -n "$WIKI" ]; then
  if [ "$PRESERVE_CONFIG" -eq 1 ]; then printf 'PRESERVE %s\n' "$WIKI/buda.yaml"; else printf 'REMOVE %s\n' "$WIKI/buda.yaml"; fi
  printf 'REMOVE managed block %s/AGENTS.md\n' "$WIKI"
fi
[ "$DRY_RUN" -eq 1 ] && exit 0
if [ "$YES" -ne 1 ]; then
  [ -t 0 ] || { printf '%s\n' '--yes is required for non-interactive uninstall' >&2; exit 2; }
  printf 'Remove Buda-owned files? [y/N] '; read answer; case "$answer" in y|Y|yes|YES) ;; *) exit 2 ;; esac
fi
rm -f "$BIN_DIR/buda"
rm -f "$BIN_DIR/buda" "$HOME_DIR/.agents/skills/guiho-s-0002-buda/SKILL.md" "$HOME_DIR/.claude/skills/guiho-s-0002-buda/SKILL.md"
if [ "$PRESERVE_DATA" -ne 1 ]; then
  rm -rf "$CLI_HOME/versions" "$CLI_HOME/state" "$CLI_HOME/data" "$CLI_HOME/database" "$CLI_HOME/current.json" "$CLI_HOME/installed-artifacts.json" "$CLI_HOME/cache.json"
else
  rm -rf "$CLI_HOME/versions" "$CLI_HOME/state" "$CLI_HOME/current.json" "$CLI_HOME/installed-artifacts.json" "$CLI_HOME/cache.json"
fi
[ "$PRESERVE_CONFIG" -eq 1 ] || rm -f "$CLI_HOME/buda.global.yaml"
if [ "$PRESERVE_CONFIG" -eq 0 ] && [ "$PRESERVE_DATA" -eq 0 ]; then rmdir "$CLI_HOME" 2>/dev/null || true; fi
if [ -n "$WIKI" ]; then
  [ "$PRESERVE_CONFIG" -eq 1 ] || rm -f "$WIKI/buda.yaml"
  for instruction in "$WIKI/AGENTS.md" "$WIKI/CLAUDE.md"; do
    if [ -f "$instruction" ]; then
      temporary=$(mktemp "$TEMP_ROOT/buda-uninstall-XXXXXX")
      sed '/<!-- BEGIN BUDA INSTRUCTIONS -->/,/<!-- END BUDA INSTRUCTIONS -->/d' "$instruction" > "$temporary"
      cat "$temporary" > "$instruction"
      rm -f "$temporary"
    fi
  done
  rm -f "$WIKI/.agents/skills/guiho-s-0002-buda/SKILL.md" "$WIKI/.claude/skills/guiho-s-0002-buda/SKILL.md"
fi
printf '%s\n' 'Buda uninstall completed synchronously.'
