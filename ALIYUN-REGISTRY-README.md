# 阿里云容器镜像服务配置指南

## 仓库信息

- **仓库地址**: `crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com`
- **命名空间**: `willnight`  
- **项目名称**: `emailalert`
- **登录用户**: `willnightzhanglixia@126.com`

## 镜像命名规范

### ARM版本 (Mac M1/M2)
```bash
# 后端镜像
crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-arm-latest

# 前端镜像  
crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:frontend-arm-latest
```

### X86版本 (CentOS7)
```bash
# 后端镜像
crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-x86-latest

# 前端镜像
crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:frontend-x86-latest
```

## 登录方式

### 自动登录（推荐）
构建脚本已内置自动登录功能，使用 `-p` 参数时会自动登录：

```bash
# ARM版本
./build-docker-arm.sh -p

# X86版本  
./build-docker-x86.sh -p
```

### 手动登录
```bash
# 登录命令
docker login --username=willnightzhanglixia@126.com crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com

# 输入密码: Aa56764009
```

### 使用环境变量登录
```bash
# 设置环境变量
export DOCKER_REGISTRY_PASSWORD="Aa56764009"

# 登录
echo $DOCKER_REGISTRY_PASSWORD | docker login --username=willnightzhanglixia@126.com --password-stdin crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com
```

## 构建和推送流程

### 1. ARM版本构建推送
```bash
# 构建并推送所有镜像
./build-docker-arm.sh -p

# 构建并推送指定版本
./build-docker-arm.sh -v 1.0.0 -p

# 仅构建推送后端
./build-docker-arm.sh -b -p

# 仅构建推送前端
./build-docker-arm.sh -f -p
```

### 2. X86版本构建推送
```bash
# 构建并推送所有镜像
./build-docker-x86.sh -p

# 构建并推送指定版本
./build-docker-x86.sh -v 1.0.0 -p

# 仅构建推送后端
./build-docker-x86.sh -b -p

# 仅构建推送前端
./build-docker-x86.sh -f -p
```

## 拉取镜像

### 从阿里云仓库拉取
```bash
# 登录（如果还未登录）
docker login --username=willnightzhanglixia@126.com crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com

# 拉取ARM版本镜像
docker pull crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-arm-latest
docker pull crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:frontend-arm-latest

# 拉取X86版本镜像
docker pull crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-x86-latest
docker pull crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:frontend-x86-latest
```

### 使用docker-compose拉取
```bash
# ARM版本
docker-compose pull

# X86版本
docker-compose -f docker-compose-x86.yml pull
```

## 镜像管理

### 查看本地镜像
```bash
# 查看所有EmailAlert镜像
docker images | grep emailalert

# 查看指定仓库的镜像
docker images crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert
```

### 删除本地镜像
```bash
# 删除指定镜像
docker rmi crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-arm-latest

# 删除所有EmailAlert镜像
docker images | grep emailalert | awk '{print $1":"$2}' | xargs docker rmi
```

### 镜像重命名（标签）
```bash
# 为镜像添加新标签
docker tag crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-arm-latest \
           crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-v1.0.0
```

## 自动化部署配置

### 使用阿里云镜像的docker-compose
```yaml
# ARM版本 (docker-compose.yml)
services:
  backend:
    image: crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-arm-latest
  frontend:
    image: crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:frontend-arm-latest

# X86版本 (docker-compose-x86.yml)  
services:
  backend:
    image: crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-x86-latest
  frontend:
    image: crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:frontend-x86-latest
```

## 镜像加速配置

### 在CentOS7上配置阿里云镜像加速
```bash
# 创建Docker配置目录
sudo mkdir -p /etc/docker

# 配置阿里云镜像加速
sudo tee /etc/docker/daemon.json <<-'EOF'
{
  "registry-mirrors": [
    "https://crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com",
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com"
  ]
}
EOF

# 重启Docker服务
sudo systemctl daemon-reload
sudo systemctl restart docker
```

## 故障排查

### 登录失败
```bash
# 检查网络连接
ping crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com

# 检查用户名和密码
docker login --username=willnightzhanglixia@126.com crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com

# 清除登录信息重新登录
docker logout crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com
```

### 推送失败
```bash
# 检查镜像是否存在
docker images | grep emailalert

# 检查推送权限
docker push crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-arm-latest

# 重新构建镜像
./build-docker-arm.sh -b
```

### 拉取失败
```bash
# 检查镜像是否存在于仓库
docker search crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert

# 手动拉取测试
docker pull crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com/willnight/emailalert:backend-arm-latest
```

## 版本管理建议

### 版本标签规范
- `latest` - 最新稳定版本
- `arm-latest` - ARM平台最新版本
- `x86-latest` - X86平台最新版本  
- `v1.0.0` - 具体版本号
- `backend-v1.0.0` - 后端具体版本
- `frontend-v1.0.0` - 前端具体版本

### 发布流程建议
1. **开发阶段**: 使用 `dev` 标签
2. **测试阶段**: 使用 `test` 标签  
3. **预发布**: 使用 `rc` 标签
4. **正式发布**: 使用版本号和 `latest` 标签

---

**更新时间**: 2025-06-30  
**维护者**: willnight  
**仓库状态**: 活跃 