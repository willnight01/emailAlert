#!/bin/bash

# EmailAlert Docker X86_64 构建脚本 (CentOS7 x86)
# 专为 CentOS7 x86 服务器优化的版本

set -e

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置变量
DOCKER_REGISTRY="crpi-0vtsukduyebtna5k.cn-hangzhou.personal.cr.aliyuncs.com"
PROJECT_NAMESPACE="willnight"
PROJECT_NAME="emailalert"
PLATFORM="linux/amd64"  # 固定为AMD64平台

# 打印消息函数
print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_error() {
    echo -e "${RED}$1${NC}"
}

print_info() {
    echo -e "${BLUE}$1${NC}"
}

# 显示帮助信息
show_help() {
    echo "EmailAlert Docker X86_64 构建脚本 (CentOS7)"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -v VERSION    指定镜像版本 (默认: x86-latest)"
    echo "  -p           构建完成后推送到镜像仓库"
    echo "  -b           仅构建后端镜像"
    echo "  -f           仅构建前端镜像"
    echo "  -h           显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -v 1.0.0 -p    # 构建并推送 v1.0.0 版本"
    echo "  $0 -b             # 仅构建后端镜像"
    echo ""
    echo "注意："
    echo "  - 适用于 CentOS7 x86_64 服务器"
    echo "  - 需要 Docker 版本 >= 18.09"
}

# 检查Docker环境
check_docker() {
    print_info "🔍 检查Docker环境..."
    
    if ! command -v docker &> /dev/null; then
        print_error "❌ Docker 未安装"
        print_info "💡 CentOS7 安装命令："
        echo "   sudo yum install -y yum-utils"
        echo "   sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo"
        echo "   sudo yum install docker-ce docker-ce-cli containerd.io"
        echo "   sudo systemctl start docker"
        echo "   sudo systemctl enable docker"
        exit 1
    fi
    
    # 检查Docker版本
    DOCKER_VERSION=$(docker --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    print_info "   Docker版本: $DOCKER_VERSION"
    
    # 检查Docker服务状态
    if ! systemctl is-active --quiet docker; then
        print_error "❌ Docker服务未运行"
        print_info "💡 启动命令: sudo systemctl start docker"
        exit 1
    fi
    
    print_success "✅ Docker环境正常"
}

# 检查磁盘空间
check_disk_space() {
    print_info "💾 检查磁盘空间..."
    
    # 检查当前目录可用空间 (需要至少2GB)
    AVAILABLE_SPACE=$(df . | tail -1 | awk '{print $4}')
    REQUIRED_SPACE=$((2 * 1024 * 1024))  # 2GB in KB
    
    if [[ $AVAILABLE_SPACE -lt $REQUIRED_SPACE ]]; then
        print_error "❌ 磁盘空间不足 (需要至少2GB)"
        print_info "💡 当前可用: $(($AVAILABLE_SPACE / 1024 / 1024))GB"
        exit 1
    fi
    
    print_success "✅ 磁盘空间充足"
}

# 清理Docker缓存
cleanup_docker() {
    print_info "🧹 清理Docker缓存..."
    docker system prune -f > /dev/null 2>&1 || true
    print_success "✅ Docker缓存清理完成"
}

# 检查Docker登录状态
check_docker_login() {
    if [[ "$PUSH_IMAGE" == "true" ]]; then
        print_info "🔐 检查阿里云仓库登录状态..."
        
        # 检查是否已登录阿里云仓库
        if ! docker info | grep -q "$DOCKER_REGISTRY"; then
            print_info "📝 登录阿里云容器镜像服务..."
            echo "Aa56764009" | docker login --username=willnightzhanglixia@126.com --password-stdin $DOCKER_REGISTRY
            
            if [[ $? -eq 0 ]]; then
                print_success "✅ 阿里云仓库登录成功"
            else
                print_error "❌ 阿里云仓库登录失败"
                exit 1
            fi
        else
            print_success "✅ 已登录阿里云仓库"
        fi
    fi
}

# 构建镜像
build_image() {
    local service=$1
    local image_name="$DOCKER_REGISTRY/$PROJECT_NAMESPACE/$PROJECT_NAME:$service-$VERSION"
    
    print_info "🔨 构建 $service 镜像..."
    print_info "   镜像: $image_name"
    print_info "   平台: $PLATFORM"
    print_info "   系统: CentOS7 x86_64"
    
    # 检查Dockerfile是否存在
    if [[ ! -f "$service/Dockerfile" ]]; then
        print_error "❌ 找不到 $service/Dockerfile"
        exit 1
    fi
    
    # 构建命令
    if docker build \
        --platform "$PLATFORM" \
        --no-cache \
        -t "$image_name" \
        -f "$service/Dockerfile" \
        "$service/"; then
        
        print_success "✅ $service 镜像构建成功"
        
        # 显示镜像信息
        IMAGE_SIZE=$(docker images "$image_name" --format "table {{.Size}}" | tail -1)
        print_info "   镜像大小: $IMAGE_SIZE"
        
        # 如果需要推送
        if [[ "$PUSH_IMAGE" == "true" ]]; then
            print_info "📤 推送 $service 镜像..."
            if docker push "$image_name"; then
                print_success "✅ $service 镜像推送成功"
            else
                print_error "❌ $service 镜像推送失败"
                exit 1
            fi
        fi
    else
        print_error "❌ $service 镜像构建失败"
        exit 1
    fi
}

# 显示构建结果
show_results() {
    print_success "🎉 构建完成！"
    echo ""
    print_info "📋 构建结果："
    
    if [[ "$BUILD_BACKEND" == "true" ]]; then
        echo "  🔧 后端: $DOCKER_REGISTRY/$PROJECT_NAMESPACE/$PROJECT_NAME:backend-$VERSION"
    fi
    
    if [[ "$BUILD_FRONTEND" == "true" ]]; then
        echo "  🌐 前端: $DOCKER_REGISTRY/$PROJECT_NAMESPACE/$PROJECT_NAME:frontend-$VERSION"
    fi
    
    echo "  🏗️ 平台: $PLATFORM"
    echo "  💻 系统: CentOS7 x86_64"
    echo "  📤 推送: $([ "$PUSH_IMAGE" == "true" ] && echo "已推送" || echo "仅本地")"
    echo ""
    
    # 显示使用说明
    print_info "🚀 使用说明："
    echo "  1. 启动服务: docker-compose -f docker-compose-x86.yml up -d"
    echo "  2. 查看状态: docker-compose -f docker-compose-x86.yml ps"
    echo "  3. 查看日志: docker-compose -f docker-compose-x86.yml logs"
    echo "  4. 停止服务: docker-compose -f docker-compose-x86.yml down"
}

# 主函数
main() {
    print_info "🚀 开始构建 X86_64 镜像 (CentOS7)..."
    
    # 检查环境
    check_docker
    check_docker_login
    check_disk_space
    
    # 清理缓存
    if [[ "$CLEANUP" == "true" ]]; then
        cleanup_docker
    fi
    
    # 构建镜像
    if [[ "$BUILD_BACKEND" == "true" ]]; then
        build_image "backend"
    fi
    
    if [[ "$BUILD_FRONTEND" == "true" ]]; then
        build_image "frontend"
    fi
    
    # 显示结果
    show_results
}

# 默认参数
VERSION="x86-latest"
PUSH_IMAGE="false"
BUILD_BACKEND="true"
BUILD_FRONTEND="true"
CLEANUP="false"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -v)
            VERSION="$2"
            shift 2
            ;;
        -p)
            PUSH_IMAGE="true"
            shift
            ;;
        -b)
            BUILD_BACKEND="true"
            BUILD_FRONTEND="false"
            shift
            ;;
        -f)
            BUILD_BACKEND="false"
            BUILD_FRONTEND="true"
            shift
            ;;
        -c)
            CLEANUP="true"
            shift
            ;;
        -h)
            show_help
            exit 0
            ;;
        *)
            print_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 验证参数
if [[ "$BUILD_BACKEND" == "false" && "$BUILD_FRONTEND" == "false" ]]; then
    print_error "❌ 必须至少构建一个服务"
    exit 1
fi

# 显示配置
print_info "📋 构建配置："
echo "  版本: $VERSION"
echo "  推送: $PUSH_IMAGE"
echo "  后端: $BUILD_BACKEND"
echo "  前端: $BUILD_FRONTEND"
echo "  平台: $PLATFORM (CentOS7 x86)"
echo "  清理: $CLEANUP"
echo ""

# 执行构建
main 