FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/butter-daemon ./cmd/butter-daemon

FROM debian:trixie-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends bash ca-certificates curl tar unzip \
    && rm -rf /var/lib/apt/lists/*

# Install the opencode CLI; the daemon launches "opencode acp" as a subprocess.
RUN curl -fsSL https://opencode.ai/install | SHELL=/bin/bash bash -s -- --no-modify-path \
    && ln -sf /root/.opencode/bin/opencode /usr/local/bin/opencode \
    && opencode --version

ARG TARGETARCH
ARG OPENAI_CODEX_VERSION=0.137.0
ARG CODEX_ACP_VERSION=0.15.0

# Install the Codex CLI; codex-acp shells out to the "codex" command.
RUN set -eux; \
    arch="${TARGETARCH:-$(dpkg --print-architecture)}"; \
    case "$arch" in \
      amd64) codex_arch="x86_64" ;; \
      arm64) codex_arch="aarch64" ;; \
      *) echo "unsupported architecture: $arch" >&2; exit 1 ;; \
    esac; \
    tmpdir="$(mktemp -d)"; \
    curl -fsSL \
      -o "$tmpdir/codex.tar.gz" \
      "https://github.com/openai/codex/releases/download/rust-v${OPENAI_CODEX_VERSION}/codex-${codex_arch}-unknown-linux-musl.tar.gz"; \
    tar -xzf "$tmpdir/codex.tar.gz" -C "$tmpdir"; \
    find "$tmpdir" -type f -name codex -exec mv {} /usr/local/bin/codex \; ; \
    chmod 755 /usr/local/bin/codex; \
    rm -rf "$tmpdir"; \
    codex --version

# Install codex-acp; the daemon exposes the "codex" ACP runtime by default.
RUN set -eux; \
    arch="${TARGETARCH:-$(dpkg --print-architecture)}"; \
    case "$arch" in \
      amd64) codex_arch="x86_64" ;; \
      arm64) codex_arch="aarch64" ;; \
      *) echo "unsupported architecture: $arch" >&2; exit 1 ;; \
    esac; \
    tmpdir="$(mktemp -d)"; \
    curl -fsSL \
      -o "$tmpdir/codex-acp.tar.gz" \
      "https://github.com/zed-industries/codex-acp/releases/download/v${CODEX_ACP_VERSION}/codex-acp-${CODEX_ACP_VERSION}-${codex_arch}-unknown-linux-gnu.tar.gz"; \
    tar -xzf "$tmpdir/codex-acp.tar.gz" -C "$tmpdir"; \
    find "$tmpdir" -type f -name codex-acp -exec mv {} /usr/local/bin/codex-acp \; ; \
    chmod 755 /usr/local/bin/codex-acp; \
    rm -rf "$tmpdir"; \
    codex-acp --version

COPY --from=builder /app/butter-daemon /app/butter-daemon

ENTRYPOINT ["/app/butter-daemon"]
