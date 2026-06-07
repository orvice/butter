FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/butter-daemon ./cmd/butter-daemon

FROM node:24-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends bash ca-certificates curl git tar \
    && rm -rf /var/lib/apt/lists/*

ARG OPENAI_CODEX_VERSION=0.137.0
ARG CODEX_ACP_VERSION=0.15.0

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
