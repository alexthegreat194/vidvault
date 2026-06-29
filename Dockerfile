# Build: docker build -t vidvault:local .
# Run:   ./scripts/docker-run.sh   (or docker run --rm -p 8765:8765 -v /path/to/videos:/data vidvault:local -p 8765 /data)
#
# Optional env (entrypoint → flags): VIDVAULT_DEBUG=1 (-d),
# VIDVAULT_DISABLE_BROWSER=1 (-disable-browser), VIDVAULT_PIN=... (-pin).
# Example: docker run ... -e VIDVAULT_DISABLE_BROWSER=1 vidvault:local -p 8765 /data
ARG GO_VERSION=1.24.5
ARG NODE_VERSION=22-alpine
FROM --platform=$BUILDPLATFORM node:${NODE_VERSION} AS frontend
WORKDIR /app
COPY package.json package-lock.json ./
COPY frontend/package.json frontend/package-lock.json ./frontend/
RUN npm ci && cd frontend && npm ci
COPY frontend ./frontend
RUN npm run build:frontend

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY src ./src
COPY --from=frontend /app/src/web/dist ./src/web/dist
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /vidvault ./src

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /vidvault /vidvault
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
WORKDIR /data
RUN mkdir -p /data
EXPOSE 8765
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["-p", "8765", "/data"]
