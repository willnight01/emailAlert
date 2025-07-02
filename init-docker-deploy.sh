#!/bin/bash

# EmailAlert Docker Compose 环境初始化脚本
# 用于创建部署所需的目录结构和配置文件

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 打印函数
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_header() {
    echo -e "${BLUE}"
    echo "======================================"
    echo "  EmailAlert Docker 部署初始化"
    echo "======================================"
    echo -e "${NC}"
}

# 检查当前目录是否合适
check_environment() {
    print_info "检查部署环境..."
    
    # 检查是否存在docker-compose.yml
    if [[ ! -f "docker-compose.yml" ]]; then
        print_warning "当前目录没有找到 docker-compose.yml"
        print_info "请确保将此脚本与 docker-compose.yml 放在同一目录"
        read -p "是否继续创建目录结构？(y/N): " confirm
        if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
            print_info "脚本已退出"
            exit 0
        fi
    fi
    
    print_success "环境检查完成"
}

# 创建目录结构
create_directories() {
    print_info "创建目录结构..."
    
    # 后端相关目录
    mkdir -p backend/data
    mkdir -p backend/logs
    mkdir -p backend/config
    
    # 前端相关目录（如果需要自定义配置）
    mkdir -p frontend/config
    
    # 其他目录
    mkdir -p scripts
    mkdir -p backups
    
    print_success "目录结构创建完成"
    
    # 显示目录结构
    print_info "创建的目录结构："
    echo "📁 当前目录/"
    echo "├── 📁 backend/"
    echo "│   ├── 📁 data/          # 数据库文件存储"
    echo "│   ├── 📁 logs/          # 应用日志"
    echo "│   └── 📁 config/        # 后端配置文件"
    echo "├── 📁 frontend/"
    echo "│   └── 📁 config/        # 前端配置文件(nginx等)"
    echo "├── 📁 scripts/           # 维护脚本"
    echo "└── 📁 backups/           # 备份文件"
}

# 创建nginx配置文件
create_nginx_config() {
    print_info "创建自定义 nginx 配置..."
    
    cat > frontend/config/nginx.conf << 'EOF'
server {
    listen 80;
    server_name localhost;
    
    # 启用gzip压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/json application/xml+rss;
    
    # 设置根目录
    root /usr/share/nginx/html;
    index index.html;
    
    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
    }
    
    # API请求代理到后端
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 处理跨域
        proxy_hide_header Access-Control-Allow-Origin;
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
        add_header Access-Control-Allow-Headers "Content-Type, Authorization, X-Requested-With";
        
        # 处理预检请求
        if ($request_method = 'OPTIONS') {
            add_header Access-Control-Allow-Origin *;
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
            add_header Access-Control-Allow-Headers "Content-Type, Authorization, X-Requested-With";
            add_header Access-Control-Max-Age 86400;
            add_header Content-Length 0;
            add_header Content-Type text/plain;
            return 204;
        }
        
        # 超时设置
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
    
    # 前端路由支持 (SPA应用)
    location / {
        try_files $uri $uri/ /index.html;
        
        # 安全头
        add_header X-Frame-Options "SAMEORIGIN" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    }
    
    # 健康检查端点
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
    
    # 错误页面
    error_page 404 /index.html;
    error_page 500 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
    
    # 禁止访问隐藏文件
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
}
EOF

    print_success "nginx配置文件已创建: frontend/config/nginx.conf"
    print_warning "注意：默认情况下容器使用内置配置，如需使用自定义配置请修改docker-compose.yml"
}

# 创建后端配置文件示例
create_backend_config() {
    print_info "创建后端配置文件示例..."
    
    # 创建用户配置示例
    cat > backend/config/users.json << 'EOF'
{
    "users": [
        {
            "username": "admin",
            "password": "admin",
            "role": "admin"
        }
    ]
}
EOF

    # 创建环境变量配置示例
    cat > backend/config/.env.example << 'EOF'
# EmailAlert 后端配置示例
# 复制此文件为 .env 并修改相应配置

# 应用配置
GIN_MODE=release
PORT=8080
TZ=Asia/Shanghai

# 数据库配置
DB_PATH=/app/data/emailalert.db

# 日志配置
LOG_LEVEL=info
LOG_FILE=/app/logs/app.log

# 邮件监控配置
EMAIL_CHECK_INTERVAL=60s
EMAIL_TIMEOUT=30s

# 通知配置
NOTIFICATION_TIMEOUT=10s
NOTIFICATION_RETRY=3
EOF

    print_success "后端配置文件已创建:"
    echo "  - backend/config/users.json (用户配置)"
    echo "  - backend/config/.env.example (环境变量示例)"
}

# 创建维护脚本
create_maintenance_scripts() {
    print_info "创建维护脚本..."
    
    # 创建启动脚本
    cat > scripts/start.sh << 'EOF'
#!/bin/bash
echo "🚀 启动 EmailAlert 服务..."
docker compose up -d
echo "✅ 服务启动完成"
echo "📱 前端地址: http://localhost:3000"
echo "🔧 后端API: http://localhost:8080"
EOF

    # 创建停止脚本
    cat > scripts/stop.sh << 'EOF'
#!/bin/bash
echo "🛑 停止 EmailAlert 服务..."
docker compose down
echo "✅ 服务已停止"
EOF

    # 创建状态检查脚本
    cat > scripts/status.sh << 'EOF'
#!/bin/bash
echo "📊 EmailAlert 服务状态:"
echo "================================"
docker compose ps
echo ""
echo "📊 资源使用情况:"
docker compose top
EOF

    # 创建日志查看脚本
    cat > scripts/logs.sh << 'EOF'
#!/bin/bash
if [ "$1" = "backend" ]; then
    echo "📋 查看后端日志..."
    docker compose logs -f backend
elif [ "$1" = "frontend" ]; then
    echo "📋 查看前端日志..."
    docker compose logs -f frontend
else
    echo "📋 查看所有日志..."
    docker compose logs -f
fi
EOF

    # 创建备份脚本
    cat > scripts/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="../backups/backup_${BACKUP_DATE}"

echo "💾 开始备份 EmailAlert 数据..."
mkdir -p "$BACKUP_DIR"

# 备份数据库
if [ -f "backend/data/emailalert.db" ]; then
    cp backend/data/emailalert.db "$BACKUP_DIR/"
    echo "✅ 数据库备份完成"
fi

# 备份配置文件
cp -r backend/config "$BACKUP_DIR/"
echo "✅ 配置文件备份完成"

# 备份日志文件(最近7天)
find backend/logs -name "*.log" -mtime -7 -exec cp {} "$BACKUP_DIR/" \;
echo "✅ 日志文件备份完成"

echo "💾 备份完成: $BACKUP_DIR"
EOF

    # 设置执行权限
    chmod +x scripts/*.sh
    
    print_success "维护脚本创建完成:"
    echo "  - scripts/start.sh   (启动服务)"
    echo "  - scripts/stop.sh    (停止服务)"
    echo "  - scripts/status.sh  (查看状态)"
    echo "  - scripts/logs.sh    (查看日志)"
    echo "  - scripts/backup.sh  (备份数据)"
}

# 创建README文件
create_readme() {
    print_info "创建部署说明文档..."
    
    cat > DEPLOY_README.md << 'EOF'
# EmailAlert Docker 部署环境

本目录包含了 EmailAlert 项目的 Docker 部署配置和必要文件。

## 📁 目录结构

```
./
├── docker-compose.yml     # Docker Compose 配置
├── init-docker-deploy.sh  # 环境初始化脚本
├── DEPLOY_README.md       # 本说明文档
├── backend/              # 后端相关文件
│   ├── data/            # 数据库文件(持久化存储)
│   ├── logs/            # 应用日志
│   └── config/          # 配置文件
├── frontend/            # 前端相关文件
│   └── config/          # nginx等配置
├── scripts/            # 维护脚本
│   ├── start.sh        # 启动服务
│   ├── stop.sh         # 停止服务
│   ├── status.sh       # 查看状态
│   ├── logs.sh         # 查看日志
│   └── backup.sh       # 备份数据
└── backups/            # 备份文件存储
```

## 🚀 快速开始

1. **初始化环境**（首次部署）:
   ```bash
   chmod +x init-docker-deploy.sh
   ./init-docker-deploy.sh
   ```

2. **启动服务**:
   ```bash
   # 方法1：使用 docker compose
   docker compose up -d
   
   # 方法2：使用维护脚本
   ./scripts/start.sh
   ```

3. **访问服务**:
   - 前端界面: http://localhost:3000
   - 后端API: http://localhost:8080
   - 默认账号: admin / admin

## 🔧 常用操作

### 查看服务状态
```bash
docker compose ps
# 或
./scripts/status.sh
```

### 查看日志
```bash
# 查看所有日志
docker compose logs -f

# 查看特定服务日志
./scripts/logs.sh backend   # 后端日志
./scripts/logs.sh frontend  # 前端日志
```

### 停止服务
```bash
docker compose down
# 或
./scripts/stop.sh
```

### 数据备份
```bash
./scripts/backup.sh
```

## 📋 配置说明

### 数据持久化
- `backend/data/`: 数据库文件存储，重要数据请定期备份
- `backend/logs/`: 应用日志文件

### 自定义配置
- `backend/config/users.json`: 用户账号配置
- `frontend/config/nginx.conf`: 自定义nginx配置(可选)

## 🔍 故障排除

### 服务启动失败
1. 检查端口占用: `netstat -tulpn | grep :3000`
2. 检查Docker状态: `docker system df`
3. 查看详细日志: `docker compose logs`

### 数据库问题
1. 检查数据目录权限: `ls -la backend/data/`
2. 检查磁盘空间: `df -h`

### 网络问题
1. 检查容器网络: `docker network ls`
2. 检查服务连通性: `docker compose exec frontend ping backend`

## 📞 支持

如有问题，请查看主项目文档或联系维护人员。
EOF

    print_success "部署说明文档已创建: DEPLOY_README.md"
}

# 设置目录权限
set_permissions() {
    print_info "设置目录权限..."
    
    # 确保数据目录有正确的权限
    chmod 755 backend/data backend/logs
    chmod 644 backend/config/* 2>/dev/null || true
    
    print_success "权限设置完成"
}

# 显示完成信息
show_completion() {
    print_header
    print_success "🎉 EmailAlert Docker 环境初始化完成！"
    echo ""
    print_info "📋 下一步操作："
    echo "1. 检查并修改配置文件（如需要）"
    echo "2. 启动服务: docker compose up -d"
    echo "3. 访问前端: http://localhost:3000"
    echo "4. 默认账号: admin / admin"
    echo ""
    print_info "📚 更多信息请查看 DEPLOY_README.md"
    echo ""
    print_warning "💡 提示: 生产环境请务必修改默认密码！"
}

# 主函数
main() {
    print_header
    
    check_environment
    create_directories
    create_nginx_config
    create_backend_config
    create_maintenance_scripts
    create_readme
    set_permissions
    show_completion
}

# 执行主函数
main "$@" 