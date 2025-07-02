### 构建阶段 ###
FROM golang:1.22-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /app

# 使用稳定国内镜像源（推荐只用一个）
RUN echo "https://mirrors.aliyun.com/alpine/v3.19/main" > /etc/apk/repositories && \
    echo "https://mirrors.aliyun.com/alpine/v3.19/community" >> /etc/apk/repositories && \
    apk update && apk add --no-cache \
    git gcc musl-dev sqlite-dev make pkgconfig

ENV CGO_ENABLED=1 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN echo "Building for platform: ${TARGETPLATFORM}" && \
    go build \
      -ldflags="-w -s -extldflags '-static'" \
      -a -installsuffix cgo \
      -tags netgo \
      -o emailAlert main.go

RUN file emailAlert && ldd emailAlert || echo "静态链接成功"


### 运行阶段 ###
FROM alpine:3.19 AS runtime

LABEL org.opencontainers.image.title="EmailAlert Backend"
LABEL org.opencontainers.image.description="企业级邮件告警管理系统后端服务"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.authors="willnight1989"
LABEL org.opencontainers.image.source="https://github.com/willnight1989/emailalert"

# 替换apk源，加速安装
RUN echo "https://mirrors.aliyun.com/alpine/v3.19/main" > /etc/apk/repositories && \
    echo "https://mirrors.aliyun.com/alpine/v3.19/community" >> /etc/apk/repositories && \
    apk update && apk --no-cache add \
    ca-certificates sqlite tzdata wget dumb-init && \
    update-ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup -s /bin/sh

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /app/emailAlert .
COPY --chown=appuser:appgroup config/ ./config/

RUN mkdir -p data logs tmp && \
    chown -R appuser:appgroup /app && \
    chmod +x emailAlert

USER appuser

EXPOSE 8080

ENV GIN_MODE=release \
    PORT=8080 \
    TZ=Asia/Shanghai \
    TMPDIR=/app/tmp

HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=5 \
    CMD wget --no-verbose --tries=1 --timeout=5 -O- http://localhost:8080/health > /dev/null || exit 1

ENTRYPOINT ["dumb-init", "--"]
CMD ["./emailAlert"]