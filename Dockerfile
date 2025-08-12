# Build the manager binary
ARG TARGETOS
ARG TARGETARCH

# 选择不同架构的基础镜像
FROM registry.cn-hangzhou.aliyuncs.com/testwydimage/golang-linux-amd64:1.19 AS golang_amd64
FROM crpi-3czu6jyk0smhrmuc.cn-hangzhou.personal.cr.aliyuncs.com/test_pbb/golang-arm64:1.19 AS golang_arm64

# 用多阶段构建根据架构切换
FROM golang_${TARGETARCH} as builder

WORKDIR /workspace
COPY config/ config/
RUN chmod -R 755 /workspace/config && chown -R 65532:65532 /workspace/config

COPY go.mod go.mod
COPY go.sum go.sum
ENV GOPROXY=https://goproxy.cn
RUN go mod download

COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/controller/ internal/controller/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -a -o manager cmd/main.go

# ---- 运行镜像 ----
FROM registry.cn-hangzhou.aliyuncs.com/testwydimage/gcr.io.distroless.static-linux-amd64:nonroot AS distroless_amd64
FROM crpi-3czu6jyk0smhrmuc.cn-hangzhou.personal.cr.aliyuncs.com/test_pbb/gcr.io.distroless.static-arm64:nonroot AS distroless_arm64

FROM distroless_${TARGETARCH}
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/config/ config/

USER 65532:65532
ENTRYPOINT ["/manager"]
