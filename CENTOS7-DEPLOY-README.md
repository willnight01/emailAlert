# EmailAlert CentOS7 X86 部署指南

## 系统要求

- **操作系统**: CentOS 7.x x86_64
- **内存**: 最低 2GB，推荐 4GB+
- **磁盘**: 最低 10GB 可用空间
- **网络**: 可访问Docker Hub（或配置镜像加速）

## 1. 环境准备

### 1.1 更新系统
```bash
sudo yum update -y
sudo yum install -y yum-utils device-mapper-persistent-data lvm2
```

### 1.2 安装Docker
```bash
# 添加Docker官方仓库
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

# 安装Docker CE
sudo yum install -y docker-ce docker-ce-cli containerd.io

# 启动Docker服务
sudo systemctl start docker
sudo systemctl enable docker

# 验证安装
docker --version
```

### 1.3 安装Docker Compose
```bash
# 下载Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

# 添加执行权限
sudo chmod +x /usr/local/bin/docker-compose

# 验证安装
docker-compose --version
```

### 1.4 配置Docker镜像加速（可选）
```bash
# 创建Docker配置目录
sudo mkdir -p /etc/docker

# 配置阿里云镜像加速
sudo tee /etc/docker/daemon.json <<-'EOF'
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com"
  ],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

# 重启Docker服务
sudo systemctl daemon-reload
sudo systemctl restart docker
```

### 1.5 用户权限配置
```bash
# 将当前用户添加到docker组（避免每次都要sudo）
sudo usermod -aG docker $USER

# 重新登录或执行以下命令
newgrp docker

# 测试权限
docker ps
```

## 2. 项目部署

### 2.1 下载项目文件
```bash
# 创建项目目录
mkdir -p /opt/emailalert
cd /opt/emailalert

# 上传以下文件到服务器：
# - build-docker-x86.sh
# - docker-compose-x86.yml
# - backend/ 目录（包含Dockerfile和源代码）
# - frontend/ 目录（包含Dockerfile和源代码）
```

### 2.2 构建镜像

#### 方式一：使用构建脚本（推荐）
```bash
# 赋予执行权限
chmod +x build-docker-x86.sh

# 构建所有镜像
./build-docker-x86.sh

# 仅构建后端
./build-docker-x86.sh -b

# 仅构建前端
./build-docker-x86.sh -f

# 构建并推送到仓库
./build-docker-x86.sh -p

# 查看帮助
./build-docker-x86.sh -h
```

#### 方式二：手动构建
```bash
# 构建后端镜像
cd backend
docker build --platform linux/amd64 -t willnight1989/emailalert-backend:x86-latest .

# 构建前端镜像
cd ../frontend
docker build --platform linux/amd64 -t willnight1989/emailalert-frontend:x86-latest .

# 返回项目根目录
cd ..
```

### 2.3 启动服务
```bash
# 创建必要的目录
mkdir -p backend/data backend/logs

# 启动服务
docker-compose -f docker-compose-x86.yml up -d

# 查看服务状态
docker-compose -f docker-compose-x86.yml ps

# 查看日志
docker-compose -f docker-compose-x86.yml logs -f
```

## 3. 服务验证

### 3.1 检查容器状态
```bash
# 查看所有容器
docker ps

# 查看网络
docker network ls
docker network inspect emailelert_emailalert-network
```

### 3.2 健康检查
```bash
# 后端健康检查
curl http://localhost:8080/health

# 前端健康检查
curl http://localhost:3000/health

# 前端页面访问
curl -I http://localhost:3000
```

### 3.3 功能测试
```bash
# 测试API代理
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'
```

## 4. 服务管理

### 4.1 常用命令
```bash
# 启动服务
docker-compose -f docker-compose-x86.yml up -d

# 停止服务
docker-compose -f docker-compose-x86.yml down

# 重启服务
docker-compose -f docker-compose-x86.yml restart

# 查看日志
docker-compose -f docker-compose-x86.yml logs backend
docker-compose -f docker-compose-x86.yml logs frontend

# 进入容器
docker exec -it emailalert-backend /bin/sh
docker exec -it emailalert-frontend /bin/sh
```

### 4.2 数据备份
```bash
# 备份数据库
tar -czf emailalert-backup-$(date +%Y%m%d).tar.gz backend/data

# 备份日志
tar -czf emailalert-logs-$(date +%Y%m%d).tar.gz backend/logs
```

### 4.3 更新服务
```bash
# 停止服务
docker-compose -f docker-compose-x86.yml down

# 重新构建镜像
./build-docker-x86.sh

# 启动服务
docker-compose -f docker-compose-x86.yml up -d
```

## 5. 防火墙配置

### 5.1 开放端口
```bash
# 开放3000端口（前端）
sudo firewall-cmd --permanent --add-port=3000/tcp

# 开放8080端口（后端，可选）
sudo firewall-cmd --permanent --add-port=8080/tcp

# 重载防火墙配置
sudo firewall-cmd --reload

# 查看开放的端口
sudo firewall-cmd --list-ports
```

## 6. 系统服务配置（开机自启）

### 6.1 创建systemd服务
```bash
sudo tee /etc/systemd/system/emailalert.service <<-'EOF'
[Unit]
Description=EmailAlert Service
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/emailalert
ExecStart=/usr/local/bin/docker-compose -f docker-compose-x86.yml up -d
ExecStop=/usr/local/bin/docker-compose -f docker-compose-x86.yml down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

# 启用服务
sudo systemctl enable emailalert.service

# 启动服务
sudo systemctl start emailalert.service

# 检查状态
sudo systemctl status emailalert.service
```

## 7. 故障排查

### 7.1 常见问题

#### 镜像构建失败
```bash
# 检查Docker版本
docker --version

# 清理Docker缓存
docker system prune -f

# 检查磁盘空间
df -h

# 重新构建
./build-docker-x86.sh -c
```

#### 容器启动失败
```bash
# 查看详细日志
docker-compose -f docker-compose-x86.yml logs

# 检查端口占用
netstat -tlnp | grep :3000
netstat -tlnp | grep :8080

# 检查目录权限
ls -la backend/
```

#### 网络连接问题
```bash
# 检查Docker网络
docker network ls
docker network inspect emailelert_emailalert-network

# 测试容器间连通性
docker exec emailalert-frontend ping emailalert-backend
```

### 7.2 日志位置
- **容器日志**: `docker-compose logs`
- **应用日志**: `./backend/logs/`
- **系统日志**: `/var/log/messages`
- **Docker日志**: `journalctl -u docker`

## 8. 性能优化

### 8.1 系统级优化
```bash
# 增加文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# 优化内核参数
echo "vm.max_map_count=262144" >> /etc/sysctl.conf
sysctl -p
```

### 8.2 Docker优化
```bash
# 配置Docker存储驱动
sudo tee -a /etc/docker/daemon.json <<-'EOF'
{
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.override_kernel_check=true"
  ]
}
EOF
```

## 9. 安全建议

1. **定期更新系统和Docker**
2. **使用非root用户运行容器**
3. **限制容器资源使用**
4. **定期备份数据**
5. **监控系统资源使用情况**
6. **配置防火墙规则**
7. **使用SSL/TLS加密**

## 10. 监控和维护

### 10.1 资源监控
```bash
# 查看容器资源使用
docker stats

# 查看系统资源
htop
free -h
df -h
```

### 10.2 定期维护
```bash
# 清理未使用的镜像和容器
docker system prune -f

# 日志轮转
logrotate -f /etc/logrotate.conf
```

---

## 联系信息

如遇到问题，请检查：
1. 系统日志：`journalctl -xe`
2. Docker日志：`docker-compose logs`
3. 容器状态：`docker ps -a`

**部署完成后访问地址**：
- 前端页面：http://服务器IP:3000
- 后端API：http://服务器IP:8080（可选） 