FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/butter-daemon ./cmd/butter-daemon

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/butter-daemon /app/butter-daemon

ENTRYPOINT ["/app/butter-daemon"]
CMD ["-config", "/etc/butter/daemon.yaml"]
