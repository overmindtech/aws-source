# Build the source binary
FROM golang:1.17 as builder

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

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o source main.go

# Use distroless as minimal base image to package the source binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/source .
USER 65532:65532

ENTRYPOINT ["/source"]
