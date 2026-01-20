# Buld stage
FROM golang:1.24 AS builder
ARG VERSION=dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-X main.version=${VERSION}" -o bpctl .

# Run stage
FROM gcr.io/distroless/cc:nonroot
COPY --from=builder /app/bpctl /app/bpctl
USER 65534
ENTRYPOINT ["/app/bpctl"]
