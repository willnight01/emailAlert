# 多阶段构建 - 后端服务
# 构建阶段
FROM golang:1.22-alpine AS builder

# 设置构建参数（自动检测目标平台）
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS=linux
ARG TARGETARCH

# 设置工作目录
WORKDIR /app

# 安装构建依赖（包含x86特定的优化）
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev \
    sqlite-dev \
    make \
    pkgconfig \
    && apk cache clean

# 设置Go环境变量（针对交叉编译优化）
ENV CGO_ENABLED=1 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖（利用Docker缓存层）
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 构建应用（自动适应目标架构）
RUN echo "Building for platform: ${TARGETPLATFORM}, OS: ${TARGETOS}, Arch: ${TARGETARCH}" && \
    go build \
    -ldflags="-w -s -extldflags '-static'" \
    -a -installsuffix cgo \
    -tags netgo \
    -o emailAlert main.go

# 验证构建结果
RUN file emailAlert && ldd emailAlert || echo "静态链接成功"

# 运行阶段
FROM alpine:3.19 AS runtime

# 设置标签信息
LABEL org.opencontainers.image.title="EmailAlert Backend"
LABEL org.opencontainers.image.description="企业级邮件告警管理系统后端服务"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.authors="willnight1989"
LABEL org.opencontainers.image.source="https://github.com/willnight1989/emailalert"

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    sqlite \
    tzdata \
    wget \
    dumb-init \
    && update-ca-certificates \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk cache clean

# 创建非root用户（统一UID/GID，兼容不同系统）
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup -s /bin/sh

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder --chown=appuser:appgroup /app/emailAlert .

# 创建必要的目录并设置权限
RUN mkdir -p config data logs tmp && \
    chown -R appuser:appgroup /app && \
    chmod +x emailAlert

# 复制配置文件
COPY --chown=appuser:appgroup config/ ./config/

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV GIN_MODE=release \
    PORT=8080 \
    TZ=Asia/Shanghai \
    TMPDIR=/app/tmp

# 健康检查（使用GET方法，兼容不同架构）
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=5 \
    CMD wget --no-verbose --tries=1 --timeout=5 -O- http://localhost:8080/health > /dev/null || exit 1

# 使用dumb-init作为PID 1，提高信号处理
ENTRYPOINT ["dumb-init", "--"]

# 启动应用
CMD ["./emailAlert"] 