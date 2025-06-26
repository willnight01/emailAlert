#!/bin/bash

echo "🚀 启动邮件告警平台..."

# 函数：停止之前的服务
cleanup_previous_services() {
    echo "🧹 检查并清理之前的服务..."
    
    # 1. 通过PID文件停止服务
    if [ -f "pids.txt" ]; then
        echo "📄 发现PID文件，正在停止之前的服务..."
        while read pid; do
            if [ ! -z "$pid" ] && kill -0 $pid 2>/dev/null; then
                echo "⏹️  停止进程 PID: $pid"
                kill $pid 2>/dev/null
                sleep 1
                # 如果进程仍然存在，强制杀死
                if kill -0 $pid 2>/dev/null; then
                    echo "🔨 强制停止进程 PID: $pid"
                    kill -9 $pid 2>/dev/null
                fi
            fi
        done < pids.txt
        rm -f pids.txt
        echo "✅ PID文件清理完成"
    fi
    
    # 2. 检查并停止占用端口的进程
    echo "🔍 检查端口占用情况..."
    
    # 检查8080端口（后端）
    BACKEND_PID=$(lsof -ti:8080 2>/dev/null)
    if [ ! -z "$BACKEND_PID" ]; then
        echo "⏹️  发现后端服务占用8080端口 (PID: $BACKEND_PID)，正在停止..."
        kill $BACKEND_PID 2>/dev/null
        sleep 2
        # 检查是否还在运行，强制杀死
        if kill -0 $BACKEND_PID 2>/dev/null; then
            echo "🔨 强制停止后端服务 (PID: $BACKEND_PID)"
            kill -9 $BACKEND_PID 2>/dev/null
        fi
        echo "✅ 后端服务已停止"
    fi
    
    # 检查3000端口（前端）
    FRONTEND_PID=$(lsof -ti:3000 2>/dev/null)
    if [ ! -z "$FRONTEND_PID" ]; then
        echo "⏹️  发现前端服务占用3000端口 (PID: $FRONTEND_PID)，正在停止..."
        kill $FRONTEND_PID 2>/dev/null
        sleep 2
        # 检查是否还在运行，强制杀死
        if kill -0 $FRONTEND_PID 2>/dev/null; then
            echo "🔨 强制停止前端服务 (PID: $FRONTEND_PID)"
            kill -9 $FRONTEND_PID 2>/dev/null
        fi
        echo "✅ 前端服务已停止"
    fi
    
    # 3. 额外检查go run main.go进程
    GO_PIDS=$(pgrep -f "go run main.go" 2>/dev/null)
    if [ ! -z "$GO_PIDS" ]; then
        echo "⏹️  发现go run进程，正在停止..."
        echo "$GO_PIDS" | xargs kill 2>/dev/null
        sleep 1
        echo "✅ go run进程已停止"
    fi
    
    echo "✨ 服务清理完成"
    echo ""
}

# 执行清理
cleanup_previous_services

# 启动后端
echo "📡 启动后端服务..."
cd backend
if [ -f "main.go" ]; then
    go run main.go &
    BACKEND_PID=$!
    echo "✅ 后端服务已启动 (PID: $BACKEND_PID)"
else
    echo "⚠️  后端服务未找到，跳过"
fi

# 回到根目录
cd ..

# 启动前端
echo "🌐 启动前端服务..."
cd frontend
npm run dev &
FRONTEND_PID=$!
echo "✅ 前端服务已启动 (PID: $FRONTEND_PID)"

# 保存PID到文件
echo $BACKEND_PID > ../pids.txt
echo $FRONTEND_PID >> ../pids.txt

# 等待服务启动
echo ""
echo "⏰ 等待服务启动..."
sleep 3

# 验证服务状态
echo "🔍 验证服务状态..."

# 检查后端服务
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ 后端服务启动成功"
else
    echo "❌ 后端服务启动失败，请检查日志"
fi

# 检查前端服务
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    echo "✅ 前端服务启动成功"
else
    echo "⏳ 前端服务正在启动中..."
fi

echo ""
echo "🎉 启动完成！"
echo "📱 前端地址: http://localhost:3000"
echo "🔧 后端地址: http://localhost:8080"
echo "🔍 后端健康检查: http://localhost:8080/health"
echo ""
echo "📋 常用命令："
echo "   ⏹️  停止服务: ./stop.sh"
echo "   📊 查看进程: ps aux | grep -E '(go run|npm run)'"
echo "   🔍 查看端口: lsof -i :8080,3000"
echo ""
echo "💡 如遇到问题，请检查："
echo "   1. 端口8080和3000是否被占用"
echo "   2. Go和Node.js环境是否正确安装"
echo "   3. 依赖包是否已安装 (go mod tidy & npm install)" 