#!/usr/bin/env bash
# Outline for human release verification (#469). Does not auto-pass acceptance.
set -euo pipefail
cat <<'EOF'
Butter mobile — manual release checklist
======================================
Run on signed release-candidate builds against a staging Butter server.
Record: git SHA, device model, OS version, pass/fail per step.

[ ] 1. Staging HTTPS server; native OAuth URIs configured
[ ] 2. Fresh install; password login; OAuth login; logout clears secrets
[ ] 3. Change server address; confirm re-login; no old Bearer on new host
[ ] 4. Member: denied workspace edit; denied API token create
[ ] 5. Owner/admin: Telegram/Linear write; Memory delete; Git binding save
[ ] 6. Chat: text + image; detach/background; foreground attach; Stop
[ ] 7. Web starts detached run; App attach same thread (then reverse)
[ ] 8. Telegram/Linear resend-only; confirm no agent rerun
[ ] 9. TalkBack/VoiceOver on chat + one management form
[ ] 10. File issues for any parity-matrix GAP still open

See mobile/docs/release.md and mobile/docs/parity-matrix.md for detail.
EOF
