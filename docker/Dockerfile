# go-blog 应用镜像（多阶段构建，build context 为项目根目录）
# syntax=docker/dockerfile:1

ARG BASE_REGISTRY=docker.1ms.run/library
ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.google.cn
ARG APK_MIRROR=mirrors.aliyun.com

FROM ${BASE_REGISTRY}/golang:1.25-alpine AS builder

ARG GOPROXY
ARG GOSUMDB
ARG APK_MIRROR

WORKDIR /build

RUN sed -i "s/dl-cdn.alpinelinux.org/${APK_MIRROR}/g" /etc/apk/repositories \
    && apk add --no-cache git ca-certificates tzdata

ENV GOPROXY=${GOPROXY} \
    GOSUMDB=${GOSUMDB}

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    set -eu; \
    for i in 1 2 3 4 5; do \
      echo "go mod download attempt ${i}/5..."; \
      go mod download && exit 0; \
      sleep 10; \
    done; \
    echo "go mod download failed after 5 attempts"; \
    exit 1

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /build/go-blog \
    ./main.go

ARG BASE_REGISTRY
ARG APK_MIRROR
FROM ${BASE_REGISTRY}/alpine:3.21

ARG APK_MIRROR

RUN sed -i "s/dl-cdn.alpinelinux.org/${APK_MIRROR}/g" /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

ENV TZ=Asia/Shanghai \
    GIN_MODE=release

WORKDIR /app

COPY --from=builder /build/go-blog /app/go-blog

RUN mkdir -p /app/logs \
    && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=90s --retries=5 \
    CMD wget -qO- http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/app/go-blog"]
