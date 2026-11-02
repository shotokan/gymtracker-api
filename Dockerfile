# syntax=docker/dockerfile:1

# ── Build ─────────────────────────────────────────────────────────────────────
# BUILDPLATFORM compiles natively and cross-compiles to the target, which is much
# faster than building an arm64 image under emulation.
# hadolint ignore=DL3029
FROM --platform=$BUILDPLATFORM golang:1.26.10-alpine AS build
WORKDIR /src

# Dependencies first: this layer stays cached until go.mod or go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg

ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/server ./cmd/dbinit

# ── Runtime ───────────────────────────────────────────────────────────────────
# Distroless has no shell or package manager, and nonroot runs as UID 65532.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/ /app/
USER nonroot:nonroot
# 8080 serves the API; 9090 serves metrics and health checks.
EXPOSE 8080 9090
ENTRYPOINT ["/app/server"]
