FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/butter-daemon ./cmd/butter-daemon

FROM node:24-slim

COPY --from=builder /usr/local/go /usr/local/go

ENV PATH="/usr/local/go/bin:${PATH}"

RUN apt-get update \
    && apt-get install -y --no-install-recommends bash ca-certificates curl git tar \
    && rm -rf /var/lib/apt/lists/*

ARG TARGETARCH
ARG GH_VERSION=2.93.0
ARG GLAB_VERSION=1.102.0
ARG OPENAI_CODEX_VERSION=0.137.0
ARG CODEX_ACP_VERSION=0.15.0

# Install GitHub CLI and GitLab CLI for coding agents that need repository
# context beyond plain git.
RUN set -eux; \
    arch="${TARGETARCH:-$(dpkg --print-architecture)}"; \
    case "$arch" in \
      amd64|arm64) cli_arch="$arch" ;; \
      *) echo "unsupported architecture: $arch" >&2; exit 1 ;; \
    esac; \
    tmpdir="$(mktemp -d)"; \
    curl -fsSL \
      -o "$tmpdir/gh.tar.gz" \
      "https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_${cli_arch}.tar.gz"; \
    tar -xzf "$tmpdir/gh.tar.gz" -C "$tmpdir"; \
    install -m 755 "$tmpdir/gh_${GH_VERSION}_linux_${cli_arch}/bin/gh" /usr/local/bin/gh; \
    curl -fsSL \
      -o "$tmpdir/glab.tar.gz" \
      "https://gitlab.com/gitlab-org/cli/-/releases/v${GLAB_VERSION}/downloads/glab_${GLAB_VERSION}_linux_${cli_arch}.tar.gz"; \
    tar -xzf "$tmpdir/glab.tar.gz" -C "$tmpdir"; \
    install -m 755 "$tmpdir/bin/glab" /usr/local/bin/glab; \
    rm -rf "$tmpdir"; \
    go version; \
    gh --version; \
    glab --version

# Install the Codex CLI and its ACP adapter from npm. codex-acp shells out to
# the "codex" command, so both must be present in the runtime image.
RUN npm install -g \
      "@openai/codex@${OPENAI_CODEX_VERSION}" \
      "@zed-industries/codex-acp@${CODEX_ACP_VERSION}" \
    && codex --version \
    && codex-acp --help >/dev/null

# Install the opencode CLI; the daemon launches "opencode acp" as a subprocess.
RUN curl -fsSL https://opencode.ai/install | SHELL=/bin/bash bash -s -- --no-modify-path \
    && ln -sf /root/.opencode/bin/opencode /usr/local/bin/opencode \
    && opencode --version

COPY --from=builder /app/butter-daemon /app/butter-daemon

ENTRYPOINT ["/app/butter-daemon"]
