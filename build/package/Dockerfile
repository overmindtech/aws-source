# Build the source binary
FROM golang:1.23-alpine AS builder
ARG TARGETOS
ARG TARGETARCH

# required for generating the version descriptor
RUN apk add --no-cache git

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . .

# Build
RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    go generate ./... \
    && GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -trimpath -o source main.go

FROM alpine:3.21
WORKDIR /
COPY --from=builder /workspace/source .
USER 65534:65534

ENTRYPOINT ["/source"]
