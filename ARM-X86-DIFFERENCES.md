# ARM vs X86 版本差异对比

## 文件对比表

| 文件类型 | ARM版本 (Mac) | X86版本 (CentOS7) |
|---------|---------------|------------------|
| 构建脚本 | `build-docker-arm.sh` | `build-docker-x86.sh` |
| 容器编排 | `docker-compose.yml` | `docker-compose-x86.yml` |
| 部署指南 | `DOCKER-BUILD-README.md` | `CENTOS7-DEPLOY-README.md` |
| **镜像仓库** | **阿里云容器镜像服务** | **阿里云容器镜像服务** |

## 主要差异

### 1. 架构平台差异

| 配置项 | ARM版本 | X86版本 |
|--------|---------|---------|
| **平台架构** | `linux/arm64` | `linux/amd64` |
| **镜像标签** | `arm-latest` | `x86-latest` |
| **目标环境** | Mac Apple Silicon | CentOS7 x86_64 |

### 2. 构建脚本差异

#### build-docker-arm.sh
```bash
PLATFORM="linux/arm64"
VERSION="arm-latest"
# 针对Mac ARM优化的简化版本
```

#### build-docker-x86.sh
```bash
PLATFORM="linux/amd64" 
VERSION="x86-latest"
# CentOS7环境检查和优化
```

**X86版本增强功能：**
- ✅ CentOS7 Docker安装指导
- ✅ 磁盘空间检查（最低2GB）
- ✅ 系统服务状态检查
- ✅ Docker缓存清理选项 (`-c`)
- ✅ 详细的镜像信息显示

### 3. Docker Compose差异

#### docker-compose.yml (ARM)
```yaml
services:
  backend:
    image: crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-arm-latest
    # 基础配置
```

#### docker-compose-x86.yml (X86)  
```yaml
services:
  backend:
    image: crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-x86-latest
    # CentOS7优化配置
    security_opt:
      - no-new-privileges:true
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'
```

**X86版本增强配置：**
- ✅ 安全选项优化
- ✅ 资源限制配置
- ✅ 只读文件系统
- ✅ tmpfs挂载配置
- ✅ 日志轮转设置
- ✅ 网络MTU优化

### 4. 环境变量差异

#### ARM版本
```yaml
environment:
  - GIN_MODE=release
  - PORT=8080
  - DB_PATH=/app/data/emailalert.db
```

#### X86版本
```yaml
environment:
  <<: *common-variables  # 引用通用变量
  GIN_MODE: release
  PORT: 8080
  DB_PATH: /app/data/emailalert.db

# 通用变量模板
x-common-variables: &common-variables
  TZ: Asia/Shanghai
  LANG: zh_CN.UTF-8
  LC_ALL: zh_CN.UTF-8
```

### 5. 部署指南差异

| 特性 | ARM版本 | X86版本 |
|------|---------|---------|
| **系统要求** | Mac M1/M2 | CentOS7 x86_64 |
| **Docker安装** | 桌面版 | 服务器版CE |
| **镜像加速** | 无需配置 | 配置国内镜像源 |
| **防火墙配置** | 无需 | firewall-cmd配置 |
| **系统服务** | 无需 | systemd自启配置 |
| **监控工具** | Docker Desktop | 命令行工具 |

## 使用场景

### ARM版本适用于：
- 🖥️ Mac M1/M2开发环境
- 🔧 本地开发和测试
- 🏃 快速原型验证
- 📱 移动端开发适配

### X86版本适用于：
- 🏭 生产环境部署
- 🖥️ 传统x86服务器
- 📊 企业级应用
- 🔒 安全性要求高的环境

## 迁移指南

### 从ARM到X86
1. 下载X86版本文件到CentOS7服务器
2. 执行环境准备脚本
3. 使用`build-docker-x86.sh`构建镜像
4. 通过`docker-compose-x86.yml`部署

### 命令对比

| 操作 | ARM版本 | X86版本 |
|------|---------|---------|
| **构建** | `./build-docker-arm.sh` | `./build-docker-x86.sh` |
| **启动** | `docker compose up -d` | `docker-compose -f docker-compose-x86.yml up -d` |
| **查看状态** | `docker compose ps` | `docker-compose -f docker-compose-x86.yml ps` |
| **查看日志** | `docker compose logs` | `docker-compose -f docker-compose-x86.yml logs` |
| **停止** | `docker compose down` | `docker-compose -f docker-compose-x86.yml down` |

## 性能对比

### 预期差异
- **构建时间**: X86版本可能较慢（网络下载）
- **镜像大小**: 基本相同（45-65MB）
- **运行性能**: 取决于硬件配置
- **内存使用**: X86版本有限制配置（更节约）

## 注意事项

⚠️ **重要提醒**：
1. **不要混用配置文件** - ARM和X86版本的配置文件不兼容
2. **镜像标签不同** - 确保使用正确的标签（arm-latest vs x86-latest）
3. **网络环境** - X86部署可能需要配置镜像加速
4. **权限差异** - CentOS7需要sudo权限进行系统配置
5. **防火墙** - 确保开放必要端口（3000, 8080）

---

**文件创建时间**: 2025-06-30  
**适用版本**: EmailAlert v1.0+  
**维护状态**: 活跃维护中 