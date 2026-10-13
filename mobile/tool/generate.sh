#!/usr/bin/env bash
# Regenerate raw bindings with the local generators resolved by pubspec.lock.
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."

mode=${1:-generate}
if [[ $# -gt 1 || ( "$mode" != generate && "$mode" != --check ) ]]; then
  echo 'Usage: tool/generate.sh [--check]' >&2
  exit 2
fi
if [[ $(buf --version) != 1.71.0 ]]; then
  echo 'Buf 1.71.0 is required.' >&2
  exit 1
fi
flutter pub get --enforce-lockfile

if [[ "$mode" == --check ]]; then
  tmp=$(mktemp -d)
  trap 'rm -rf -- "$tmp"' EXIT
  trap 'exit 130' INT
  trap 'exit 143' TERM
  mkdir -p "$tmp/baseline" "$tmp/regenerated/lib/gen"
  # Snapshot the filesystem, not Git: untracked files must participate too.
  if [[ -d lib/gen ]]; then
    cp -R lib/gen/. "$tmp/baseline/"
  fi
  # A clean output tree catches obsolete generated files without changing the
  # working copy (important when checking an uncommitted mobile scaffold).
  buf generate --output "$tmp/regenerated"
  if ! diff -ru "$tmp/baseline" "$tmp/regenerated/lib/gen"; then
    echo 'Generated bindings drifted; run mobile/tool/generate.sh.' >&2
    exit 1
  fi
else
  rm -rf -- lib/gen
  buf generate
fi
# Never format lib/gen: analyzer excludes the raw generator output.
