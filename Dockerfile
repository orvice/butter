FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/butter ./cmd/butter

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/butter /app/butter

ENTRYPOINT ["/app/butter"]
