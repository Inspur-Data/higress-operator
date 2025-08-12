# Build the manager binary
ARG TARGETOS
ARG TARGETARCH
FROM registry.cn-hangzhou.aliyuncs.com/testwydimage/golang-linux-${TARGETARCH}:1.19 as builder

WORKDIR /workspace
COPY config/ config/
RUN chmod -R 755 /workspace/config && chown -R 65532:65532 /workspace/config

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source
ENV GOPROXY=https://goproxy.cn
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/controller/ internal/controller/

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -a -o manager cmd/main.go

# Use distroless as minimal base image to package the manager binary
FROM registry.cn-hangzhou.aliyuncs.com/testwydimage/gcr.io.distroless.static-linux-${TARGETARCH}:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/config/ config/

USER 65532:65532
ENTRYPOINT ["/manager"]

