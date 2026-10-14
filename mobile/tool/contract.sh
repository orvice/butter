#!/usr/bin/env bash
# Run the mobile binary Connect contract against production Go HTTP handlers.
set -euo pipefail

root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
tmp=$(mktemp -d)
pid=
cleanup() {
  if [[ -n "$pid" ]]; then
    kill "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
  fi
  rm -rf -- "$tmp"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

cd -- "$root"
# Build before launching, so the trapped PID is the server, not go run's parent.
go build -o "$tmp/contractserver" ./mobile/tool/contractserver
"$tmp/contractserver" >"$tmp/url" &
pid=$!

for ((attempt = 0; attempt < 100; attempt++)); do
  if [[ -s "$tmp/url" ]]; then
    break
  fi
  if ! kill -0 "$pid" 2>/dev/null; then
    echo 'Contract fixture exited before announcing its URL.' >&2
    exit 1
  fi
  sleep 0.1
done
if [[ ! -s "$tmp/url" ]]; then
  echo 'Timed out waiting for the contract fixture URL.' >&2
  exit 1
fi
IFS= read -r url <"$tmp/url"
if [[ "$url" != http://* ]]; then
  echo "Invalid contract fixture URL: $url" >&2
  exit 1
fi

cd -- "$root/mobile"
# The files share one single-user, mutable fixture: contract_test.dart
# temporarily changes the mobile-test password and revokes sessions, which
# would fail another file's login if the files ran in parallel. flutter test
# runs files concurrently by default, so run them one at a time.
flutter test --concurrency=1 test/contract_test.dart test/native_oauth_contract_test.dart test/model_provider_contract_test.dart test/agent_contract_test.dart "--dart-define=BUTTER_TEST_SERVER=$url"
