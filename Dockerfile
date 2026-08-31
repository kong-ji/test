# syntax=docker/dockerfile:1

# =============================================================================
# Stage 1: build — compile a static server binary.
# Multi-stage keeps the final image minimal (no Go toolchain, no source).
# =============================================================================
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Cache module downloads separately from source so rebuilds are fast when only
# code changes. go.mod carries no third-party deps (stdlib only), so nothing
# extra is pulled at build time — a narrow supply-chain surface.
COPY go.mod ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 -> fully static binary (no glibc/musl dynamic links).
# -trimpath strips local build paths for reproducible output.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/server ./cmd/server

# =============================================================================
# Stage 2: runtime — alpine (keeps busybox wget for the healthcheck), non-root
# user, read-only filesystem.
# =============================================================================
FROM alpine:3.19

# ca-certificates for any future outbound TLS; tzdata for correct timestamps.
RUN apk add --no-cache ca-certificates tzdata

# Non-root, unidentifiable runtime user. No shell is needed to run the binary;
# busybox sh remains available solely for the container healthcheck.
RUN addgroup -S -g 10001 app && adduser -S -D -u 10001 -G app app

COPY --from=builder /out/server /server
# Default rules document; normally overridden by a read-only bind mount so
# rules stay decoupled from code (see docker-compose.yml).
COPY configs/rules.json /etc/banner-fp/rules.json

USER 10001:10001

EXPOSE 8080

ENTRYPOINT ["/server"]
