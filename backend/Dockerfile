# 多阶段构建 - 后端服务
# 构建阶段
FROM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖（利用Docker缓存层）
RUN go mod download

# 复制源代码
COPY . .

# 设置构建环境变量并构建应用
RUN CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=arm64 \
    go build -ldflags="-w -s" -a -installsuffix cgo -o emailAlert main.go

# 运行阶段
FROM alpine:3.19

# 设置标签信息
LABEL org.opencontainers.image.title="EmailAlert Backend"
LABEL org.opencontainers.image.description="企业级邮件告警管理系统后端服务"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.authors="willnight1989"

# 安装运行时依赖
RUN apk --no-cache add ca-certificates sqlite tzdata wget && \
    update-ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder --chown=appuser:appgroup /app/emailAlert .

# 创建必要的目录
RUN mkdir -p config data logs && \
    chown -R appuser:appgroup /app

# 复制配置文件
COPY --chown=appuser:appgroup config/ ./config/

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV GIN_MODE=release \
    PORT=8080 \
    TZ=Asia/Shanghai

# 健康检查（使用GET方法）
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --no-verbose --tries=1 -O- http://localhost:8080/health > /dev/null || exit 1

# 启动应用
CMD ["./emailAlert"] 