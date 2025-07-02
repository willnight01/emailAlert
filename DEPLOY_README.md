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
