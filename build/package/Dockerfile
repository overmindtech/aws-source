# Build the source binary
FROM golang:1.21-alpine3.18 as builder
ARG TARGETOS
ARG TARGETARCH

ENV GOEXPERIMENT=loopvar

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY cmd/ cmd/
COPY sources/ sources/
COPY tracing/ tracing/

# Build
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o source main.go

FROM alpine:3.19
WORKDIR /
COPY --from=builder /workspace/source .
USER 65534:65534

ENTRYPOINT ["/source"]
