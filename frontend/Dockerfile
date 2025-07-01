# 多阶段构建 - 前端服务
# 构建阶段
FROM node:20-alpine AS builder

# 设置构建参数（自动检测目标平台）
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS=linux
ARG TARGETARCH

# 设置工作目录
WORKDIR /app

# 安装构建依赖（针对x86架构优化）
RUN apk add --no-cache \
    python3 \
    make \
    g++ \
    git \
    && apk cache clean

# 设置npm配置（针对国内环境和x86架构优化）
RUN npm config set registry https://registry.npmmirror.com && \
    npm config set sass_binary_site https://npmmirror.com/mirrors/node-sass/ && \
    npm config set phantomjs_cdnurl https://npmmirror.com/mirrors/phantomjs/ && \
    npm config set electron_mirror https://npmmirror.com/mirrors/electron/ && \
    npm config set sqlite3_binary_host_mirror https://npmmirror.com/mirrors/ && \
    npm config set profiler_binary_host_mirror https://npmmirror.com/mirrors/node-inspector/ && \
    npm config set chromedriver_cdnurl https://npmmirror.com/mirrors/chromedriver/

# 复制package文件（利用Docker缓存层）
COPY package*.json ./

# 安装依赖（优化安装过程）
RUN echo "Installing dependencies for platform: ${TARGETPLATFORM}, Arch: ${TARGETARCH}" && \
    npm ci --no-audit --no-fund --prefer-offline && \
    npm cache clean --force

# 复制源代码
COPY . .

# 构建应用（添加构建信息）
RUN echo "Building frontend for ${TARGETPLATFORM}" && \
    npm run build && \
    echo "Build completed successfully"

# 验证构建结果
RUN ls -la dist/ && echo "Frontend build verification complete"

# 生产阶段
FROM nginx:1.25-alpine AS runtime

# 设置标签信息
LABEL org.opencontainers.image.title="EmailAlert Frontend"
LABEL org.opencontainers.image.description="企业级邮件告警管理系统前端服务"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.authors="willnight1989"
LABEL org.opencontainers.image.source="https://github.com/willnight1989/emailalert"

# 安装必要工具（针对不同架构优化）
RUN apk add --no-cache \
    curl \
    wget \
    tzdata \
    dumb-init \
    tini \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk cache clean

# 删除默认nginx配置
RUN rm -rf /etc/nginx/conf.d/default.conf /usr/share/nginx/html/*

# 复制自定义nginx配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 从构建阶段复制构建结果
COPY --from=builder /app/dist /usr/share/nginx/html

# 设置正确的权限（增强安全性）
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chown -R nginx:nginx /etc/nginx/conf.d && \
    chmod -R 755 /usr/share/nginx/html && \
    chmod 644 /etc/nginx/conf.d/default.conf

# 创建health检查文件和必要的目录
RUN echo "healthy" > /usr/share/nginx/html/health && \
    mkdir -p /var/run/nginx && \
    chown nginx:nginx /var/run/nginx

# 创建nginx用户的临时目录
RUN mkdir -p /tmp/nginx && \
    chown -R nginx:nginx /tmp/nginx

# 暴露端口
EXPOSE 80

# 设置环境变量
ENV TZ=Asia/Shanghai \
    NGINX_ENTRYPOINT_QUIET_LOGS=1

# 健康检查（增强兼容性，支持wget作为备选）
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD curl -f http://localhost/health || wget --no-verbose --tries=1 --timeout=5 -O- http://localhost/health > /dev/null || exit 1

# 使用tini作为PID 1，提高信号处理（轻量级选择）
ENTRYPOINT ["tini", "--"]

# 启动nginx（增加配置测试）
CMD ["sh", "-c", "nginx -t && exec nginx -g 'daemon off;'"] 