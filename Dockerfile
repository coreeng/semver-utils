# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build both binaries with version info
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/coreeng/semver-utils/internal/build.BuildVersion=${VERSION} -X github.com/coreeng/semver-utils/internal/build.BuildCommit=${COMMIT} -X github.com/coreeng/semver-utils/internal/build.BuildDate=${BUILD_DATE}" \
    -o /build/semver \
    ./cmd/semver

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/coreeng/semver-utils/internal/build.BuildVersion=${VERSION} -X github.com/coreeng/semver-utils/internal/build.BuildCommit=${COMMIT} -X github.com/coreeng/semver-utils/internal/build.BuildDate=${BUILD_DATE}" \
    -o /build/semver-git \
    ./cmd/semver-git

# Final stage - using alpine for shell support (needed for entrypoint script)
FROM alpine:latest

# Copy binaries from builder
COPY --from=builder /build/semver /usr/local/bin/semver
COPY --from=builder /build/semver-git /usr/local/bin/semver-git

# Copy entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Use wrapper script as entrypoint
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["semver", "--help"]
