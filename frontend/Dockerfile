# ===================
# 构建阶段
# ===================
FROM node:20-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /app

RUN echo "https://mirrors.aliyun.com/alpine/v3.19/main" > /etc/apk/repositories && \
    echo "https://mirrors.aliyun.com/alpine/v3.19/community" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache \
    python3 \
    make \
    g++ \
    git

RUN echo "registry=https://registry.npmmirror.com" >> ~/.npmrc && \
    echo "chromedriver_cdnurl=https://npmmirror.com/mirrors/chromedriver/" >> ~/.npmrc && \
    echo "electron_mirror=https://npmmirror.com/mirrors/electron/" >> ~/.npmrc && \
    echo "sqlite3_binary_host_mirror=https://npmmirror.com/mirrors/" >> ~/.npmrc

COPY package*.json ./

RUN npm ci --no-audit --no-fund --prefer-offline --cache .npm && \
    npm cache clean --force

COPY . .

RUN npm run build && echo "✅ 前端构建完成"

RUN ls -la dist/

# ===================
# 运行阶段
# ===================
FROM nginx:1.25-alpine AS runtime

LABEL org.opencontainers.image.title="EmailAlert Frontend"
LABEL org.opencontainers.image.description="企业级邮件告警系统前端"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.authors="willnight1989"

# 安装 tzdata、curl 用于健康检查
RUN apk add --no-cache tzdata curl tini && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 删除默认页面，替换为前端打包内容
RUN rm -rf /usr/share/nginx/html/* /etc/nginx/conf.d/default.conf

# 拷贝构建好的前端页面
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制自定义 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 健康检查路径
RUN echo "healthy" > /usr/share/nginx/html/health

# 设置 nginx 用户权限（可选）
RUN chown -R nginx:nginx /usr/share/nginx && \
    chmod -R 755 /usr/share/nginx/html

EXPOSE 80

# 健康检查（可选）
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
  CMD curl -f http://localhost/health || exit 1

# 设置 tini 作为 PID 1
ENTRYPOINT ["tini", "--"]

# 启动 nginx（务必用非守护模式）
CMD ["sh", "-c", "nginx -t && exec nginx -g 'daemon off;'"]