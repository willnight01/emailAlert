# 多阶段构建 - 前端服务
# 构建阶段
FROM node:20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置npm镜像源（提高构建速度）
RUN npm config set registry https://registry.npmmirror.com

# 复制package文件（利用Docker缓存层）
COPY package*.json ./

# 安装依赖
RUN npm ci --only=production=false && npm cache clean --force

# 复制源代码
COPY . .

# 构建应用
RUN npm run build

# 生产阶段
FROM nginx:1.25-alpine

# 设置标签信息
LABEL org.opencontainers.image.title="EmailAlert Frontend"
LABEL org.opencontainers.image.description="企业级邮件告警管理系统前端服务"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.authors="willnight1989"

# 安装必要工具
RUN apk add --no-cache curl tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 删除默认nginx配置
RUN rm -rf /etc/nginx/conf.d/default.conf /usr/share/nginx/html/*

# 复制自定义nginx配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 从构建阶段复制构建结果
COPY --from=builder /app/dist /usr/share/nginx/html

# 设置正确的权限
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chmod -R 755 /usr/share/nginx/html

# 创建health检查文件
RUN echo "healthy" > /usr/share/nginx/html/health

# 暴露端口
EXPOSE 80

# 设置环境变量
ENV TZ=Asia/Shanghai

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost/health || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"] 