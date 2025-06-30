# EmailAlert Docker 构建和部署指南

## 🏗️ 项目架构优化

### 优化内容
- ✅ **优化Dockerfile**: 支持多平台构建（ARM64 + AMD64）
- ✅ **简化脚本**: 移除 docker-start.sh 和 docker-stop.sh
- ✅ **统一环境**: 使用生产环境配置
- ✅ **多平台构建**: 创建 build-docker.sh 脚本
- ✅ **镜像仓库**: 推送到 willnight1989/emailalert

## 📁 文件结构

```
emailElert/
├── backend/
│   └── Dockerfile          # 后端多平台构建文件
├── frontend/
│   └── Dockerfile          # 前端多平台构建文件
├── docker-compose.yml      # 容器编排配置
├── build-docker.sh         # 多平台构建脚本
└── DOCKER-BUILD-README.md  # 本文档
```

## 🔧 Dockerfile 优化详情

### 后端 Dockerfile 特性
- **多平台支持**: 使用 `--platform=$BUILDPLATFORM`
- **多阶段构建**: 构建阶段 + 运行阶段
- **安全性**: 非root用户运行
- **时区设置**: 统一使用 Asia/Shanghai
- **健康检查**: 使用 `/health` 端点
- **镜像优化**: 使用 `-ldflags="-w -s"` 减小体积

### 前端 Dockerfile 特性
- **多平台支持**: 支持 ARM64 和 AMD64
- **构建优化**: 使用 npm 国内镜像源
- **Nginx配置**: 自定义配置和权限设置
- **健康检查**: 独立的 health 端点
- **缓存优化**: 分层构建提高构建效率

## 🚀 使用指南

### 1. 快速开始

```bash
# 拉取并启动服务
docker compose up -d

# 停止服务
docker compose down
```

### 2. 构建镜像

```bash
# 显示帮助信息
./build-docker.sh --help

# 构建测试版本（本地）
./build-docker.sh -t

# 构建指定版本
./build-docker.sh -v 1.0.0

# 构建并推送到仓库
./build-docker.sh -v 1.0.0 -p

# 仅构建后端
./build-docker.sh -b -v 1.0.0

# 仅构建前端
./build-docker.sh -f -v 1.0.0
```

### 3. 镜像推送

```bash
# 登录Docker Hub
docker login

# 构建并推送最新版本
./build-docker.sh -p

# 构建并推送指定版本
./build-docker.sh -v 2.0.0 -p
```

## 📋 构建脚本功能

### build-docker.sh 特性
- **多平台构建**: 支持 linux/amd64 和 linux/arm64
- **智能模式**: 测试模式自动优化构建参数
- **环境检查**: 自动检查 Docker 和 Buildx 环境
- **登录验证**: 推送前自动检查登录状态
- **详细日志**: 彩色输出和详细构建信息
- **错误处理**: 完善的错误处理和退出机制

### 支持的参数
| 参数 | 描述 | 示例 |
|------|------|------|
| `-v, --version` | 指定镜像版本 | `-v 1.0.0` |
| `-p, --push` | 构建后推送到仓库 | `-p` |
| `-b, --backend-only` | 仅构建后端镜像 | `-b` |
| `-f, --frontend-only` | 仅构建前端镜像 | `-f` |
| `-t, --test` | 构建测试版本 | `-t` |
| `-h, --help` | 显示帮助信息 | `-h` |

## 🎯 部署方式

### 方式1: 使用预构建镜像（推荐）
```bash
# 直接使用docker-compose
docker compose up -d
```

### 方式2: 本地构建
```bash
# 构建本地镜像
./build-docker.sh -t

# 更新docker-compose.yml使用本地镜像
# 然后启动
docker compose up -d
```

## 🔍 故障排除

### 常见问题

1. **网络超时问题**
   ```bash
   # 解决方案: 使用代理或国内镜像源
   docker login --help
   ```

2. **多平台构建失败**
   ```bash
   # 检查buildx状态
   docker buildx ls
   
   # 重新创建构建器
   docker buildx create --name mybuilder --use
   ```

3. **权限问题**
   ```bash
   # 给脚本添加执行权限
   chmod +x build-docker.sh
   ```

4. **Docker Hub登录**
   ```bash
   # 登录Docker Hub
   docker login
   
   # 验证登录状态
   docker info | grep Username
   ```

## 📊 性能指标

### 镜像大小
- **后端镜像**: ~60MB (优化后)
- **前端镜像**: ~65MB (优化后)

### 构建时间
- **后端构建**: ~2-3分钟
- **前端构建**: ~1-2分钟
- **多平台构建**: ~5-8分钟

### 支持平台
- ✅ linux/amd64 (Intel/AMD)
- ✅ linux/arm64 (Apple Silicon/ARM)

## 🎨 镜像信息

### Docker Hub 仓库
- **后端镜像**: `willnight1989/emailalert-backend`
- **前端镜像**: `willnight1989/emailalert-frontend`

### 镜像标签
- `latest`: 最新稳定版本
- `v1.0.0`: 特定版本号
- `test`: 测试版本

## 📝 版本说明

### v1.0.0 (当前版本)
- 多平台Docker支持
- 优化构建脚本
- 统一生产环境配置
- 完善的健康检查机制

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交代码更改
4. 推送到分支
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

---

**EmailAlert** - 企业级邮件告警管理系统 