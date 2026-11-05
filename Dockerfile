# Build the manager binary
# --platform=$BUILDPLATFORM forces the builder to run natively on the CI runner.
# The binary is cross-compiled via GOARCH=$TARGETARCH, so no QEMU emulation is needed.
FROM --platform=$BUILDPLATFORM mirror.gcr.io/library/golang:1.27 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/controller/ internal/controller/
COPY internal/services internal/services/
COPY internal/webhook internal/webhook/

# Build
# Cross-compile for $TARGETARCH on the native $BUILDPLATFORM runner (no emulation).
# No -a flag so Go build cache layers are reused.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -o manager cmd/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
