#!/bin/bash

# EmailAlert Docker X86_64 构建脚本 (CentOS7 x86)
# 专为 CentOS7 x86 服务器优化的版本

set -e

# 启用Docker BuildKit以支持缓存挂载
export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=auto

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
    echo "  -c           清理Docker缓存（强制重新构建）"
    echo "  -h           显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -v 1.0.0 -p    # 构建并推送 v1.0.0 版本"
    echo "  $0 -b             # 仅构建后端镜像"
    echo "  $0 -c -f          # 清理缓存并重新构建前端"
    echo "  $0 -v latest -p -c # 清理缓存，构建并推送latest版本"
    echo ""
    echo "注意："
    echo "  - 适用于 CentOS7 x86_64 服务器"
    echo "  - 需要 Docker 版本 >= 18.09"
    echo "  - 已修复npm配置废弃选项问题"
    echo "  - 默认启用Docker缓存以加速构建"
    echo "  - 使用-c参数可强制清理缓存重建"
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
    
    # 检查Docker服务状态（兼容不同操作系统）
    if ! docker info >/dev/null 2>&1; then
        print_error "❌ Docker服务未运行或无法连接"
        print_info "💡 Linux启动命令: sudo systemctl start docker"
        print_info "💡 Mac启动命令: 启动Docker Desktop应用"
        print_info "💡 Windows启动命令: 启动Docker Desktop应用"
        exit 1
    fi
    
    print_success "✅ Docker环境正常"
}

# 检查磁盘空间
check_disk_space() {
    print_info "💾 检查磁盘空间..."
    
    # 检查当前目录可用空间 (需要至少2GB)
    AVAILABLE_SPACE=$(df . | tail -1 | awk '{print $4}' 2>/dev/null || echo "0")
    REQUIRED_SPACE=$((2 * 1024 * 1024))  # 2GB in KB
    
    # 兼容性检查：如果无法获取磁盘空间，给出警告但继续
    if [[ -z "$AVAILABLE_SPACE" || "$AVAILABLE_SPACE" == "0" ]]; then
        print_info "⚠️  无法检测磁盘空间，跳过检查"
        return 0
    fi
    
    if [[ $AVAILABLE_SPACE -lt $REQUIRED_SPACE ]]; then
        print_error "❌ 磁盘空间不足 (需要至少2GB)"
        print_info "💡 当前可用: $(($AVAILABLE_SPACE / 1024 / 1024))GB"
        print_info "💡 建议清理磁盘空间或使用 -c 参数清理Docker缓存"
        exit 1
    fi
    
    print_success "✅ 磁盘空间充足 ($(($AVAILABLE_SPACE / 1024 / 1024))GB)"
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
    local start_time=$(date +%s)
    
    print_info "🔨 构建 $service 镜像..."
    print_info "   镜像: $image_name"
    print_info "   平台: $PLATFORM"
    print_info "   系统: CentOS7 x86_64"
    print_info "   缓存: $([ "$CLEANUP" == "true" ] && echo "已清理，强制重建" || echo "启用 (BuildKit + Docker缓存)")"
    
    # 检查Dockerfile是否存在
    if [[ ! -f "$service/Dockerfile" ]]; then
        print_error "❌ 找不到 $service/Dockerfile"
        exit 1
    fi
    
    # 构建命令（根据CLEANUP参数决定是否使用缓存）
    local build_args="--platform $PLATFORM"
    if [[ "$CLEANUP" == "true" ]]; then
        build_args="$build_args --no-cache"
    fi
    
    if docker build \
        $build_args \
        -t "$image_name" \
        -f "$service/Dockerfile" \
        "$service/"; then
        
        local end_time=$(date +%s)
        local build_time=$((end_time - start_time))
        print_success "✅ $service 镜像构建成功"
        
        # 显示镜像信息
        IMAGE_SIZE=$(docker images "$image_name" --format "table {{.Size}}" | tail -1)
        print_info "   镜像大小: $IMAGE_SIZE"
        print_info "   构建耗时: ${build_time}秒"
        
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