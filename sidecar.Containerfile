# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY . .

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download


# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o main sidecar/main.go

# Use distroless as minimal base image to package the sidecar binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:latest
WORKDIR /
COPY --from=builder /workspace/main .

ENTRYPOINT ["/main"]
