#!/usr/bin/env bash
# Local checks; individual phases also run independently in mobile CI.
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."

phase=${1:-all}
if [[ $# -gt 1 ]]; then
  echo 'Usage: tool/check.sh [all|format|analyze|test|contract]' >&2
  exit 2
fi
case "$phase" in
  all|format|analyze|test|contract) ;;
  *) echo 'Usage: tool/check.sh [all|format|analyze|test|contract]' >&2; exit 2 ;;
esac
flutter pub get --enforce-lockfile

format() {
  paths=(lib/app lib/core lib/features lib/main.dart test)
  if [[ -d integration_test ]]; then
    paths+=(integration_test)
  fi
  dart format --output=none --set-exit-if-changed "${paths[@]}"
}

if [[ "$phase" == all || "$phase" == format ]]; then format; fi
if [[ "$phase" == all || "$phase" == analyze ]]; then flutter analyze --no-pub; fi
if [[ "$phase" == all || "$phase" == test ]]; then flutter test --no-pub; fi
if [[ "$phase" == all || "$phase" == contract ]]; then bash tool/contract.sh; fi
