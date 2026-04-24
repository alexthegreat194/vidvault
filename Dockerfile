# Build: docker build -t vidvault:local .
# Run:   ./scripts/docker-run.sh   (or docker run --rm -p 8765:8765 -v /path/to/videos:/data vidvault:local -p 8765 /data)
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY src ./src
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /vidvault ./src

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /vidvault /vidvault
WORKDIR /data
RUN mkdir -p /data
EXPOSE 8765
ENTRYPOINT ["/vidvault"]
CMD ["-p", "8765", "/data"]
