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
    
    # 2. 停止所有相关的Go进程
    GO_PIDS=$(pgrep -f "go run main.go" 2>/dev/null || true)
    if [ ! -z "$GO_PIDS" ]; then
        echo "⏹️  发现Go进程，正在停止..."
        for pid in $GO_PIDS; do
            kill $pid 2>/dev/null
        done
        sleep 1
        echo "✅ Go进程已停止"
    fi
    
    # 3. 停止所有相关的npm进程
    NPM_PIDS=$(pgrep -f "npm run dev" 2>/dev/null || true)
    if [ ! -z "$NPM_PIDS" ]; then
        echo "⏹️  发现npm进程，正在停止..."
        for pid in $NPM_PIDS; do
            kill $pid 2>/dev/null
        done
        sleep 1
        echo "✅ npm进程已停止"
    fi
    
    # 4. 检查并停止占用端口的进程
    echo "🔍 检查端口占用情况..."
    
    # 检查8080端口（后端）
    BACKEND_PIDS=$(lsof -ti:8080 2>/dev/null || true)
    if [ ! -z "$BACKEND_PIDS" ]; then
        echo "⏹️  发现后端服务占用8080端口，正在停止..."
        for pid in $BACKEND_PIDS; do
            kill $pid 2>/dev/null
            sleep 1
            if kill -0 $pid 2>/dev/null; then
                echo "🔨 强制停止后端服务 (PID: $pid)"
                kill -9 $pid 2>/dev/null
            fi
        done
        echo "✅ 后端服务已停止"
    fi
    
    # 检查3000端口（前端）
    FRONTEND_PIDS=$(lsof -ti:3000 2>/dev/null || true)
    if [ ! -z "$FRONTEND_PIDS" ]; then
        echo "⏹️  发现前端服务占用3000端口，正在停止..."
        for pid in $FRONTEND_PIDS; do
            kill $pid 2>/dev/null
            sleep 1
            if kill -0 $pid 2>/dev/null; then
                echo "🔨 强制停止前端服务 (PID: $pid)"
                kill -9 $pid 2>/dev/null
            fi
        done
        echo "✅ 前端服务已停止"
    fi
    
    # 5. 停止Vite进程
    VITE_PIDS=$(pgrep -f "vite" 2>/dev/null || true)
    if [ ! -z "$VITE_PIDS" ]; then
        echo "⏹️  发现Vite进程，正在停止..."
        for pid in $VITE_PIDS; do
            kill $pid 2>/dev/null
        done
        sleep 1
        echo "✅ Vite进程已停止"
    fi
    
    echo "✨ 服务清理完成"
    echo ""
}

# 执行清理
cleanup_previous_services

# 清理旧的PID文件
rm -f pids.txt

# 启动后端
echo "📡 启动后端服务..."
cd backend
if [ -f "main.go" ]; then
    # 启动后端服务并获取PID
    nohup go run main.go > ../backend.log 2>&1 &
    BACKEND_PID=$!
    echo "✅ 后端服务已启动 (PID: $BACKEND_PID)"
    
    # 等待一下，确保进程启动
    sleep 2
    
    # 验证进程是否真的在运行
    if kill -0 $BACKEND_PID 2>/dev/null; then
        echo $BACKEND_PID > ../pids.txt
        echo "📝 后端PID已记录: $BACKEND_PID"
    else
        echo "❌ 后端服务启动失败"
    fi
else
    echo "⚠️  后端服务未找到，跳过"
fi

# 回到根目录
cd ..

# 启动前端
echo "🌐 启动前端服务..."
cd frontend

# 检查是否已安装依赖
if [ ! -d "node_modules" ]; then
    echo "📦 安装前端依赖..."
    npm install
fi

# 启动前端服务并获取PID
nohup npm run dev > ../frontend.log 2>&1 &
FRONTEND_PID=$!
echo "✅ 前端服务已启动 (PID: $FRONTEND_PID)"

# 等待一下，确保进程启动
sleep 2

# 验证进程是否真的在运行
if kill -0 $FRONTEND_PID 2>/dev/null; then
    echo $FRONTEND_PID >> ../pids.txt
    echo "📝 前端PID已记录: $FRONTEND_PID"
else
    echo "❌ 前端服务启动失败"
fi

# 回到根目录
cd ..

# 记录所有相关进程的PID
echo "📝 扫描并记录所有相关进程..."

# 查找并记录Go进程PID
GO_PIDS=$(pgrep -f "go run main.go" 2>/dev/null || true)
if [ ! -z "$GO_PIDS" ]; then
    for pid in $GO_PIDS; do
        # 避免重复记录
        if ! grep -q "^$pid$" pids.txt 2>/dev/null; then
            echo $pid >> pids.txt
            echo "📝 记录Go进程PID: $pid"
        fi
    done
fi

# 查找并记录Vite进程PID
VITE_PIDS=$(pgrep -f "vite" 2>/dev/null || true)
if [ ! -z "$VITE_PIDS" ]; then
    for pid in $VITE_PIDS; do
        # 避免重复记录
        if ! grep -q "^$pid$" pids.txt 2>/dev/null; then
            echo $pid >> pids.txt
            echo "📝 记录Vite进程PID: $pid"
        fi
    done
fi

# 等待服务启动
echo ""
echo "⏰ 等待服务完全启动..."
sleep 5

# 验证服务状态
echo "🔍 验证服务状态..."

# 检查后端服务
echo "检查后端服务 (http://localhost:8080)..."
if curl -s --connect-timeout 5 http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ 后端服务启动成功"
elif curl -s --connect-timeout 5 http://localhost:8080 > /dev/null 2>&1; then
    echo "✅ 后端服务启动成功 (health接口未响应，但主服务可用)"
else
    echo "❌ 后端服务启动可能失败，请检查日志: tail -f backend.log"
fi

# 检查前端服务
echo "检查前端服务 (http://localhost:3000)..."
for i in {1..10}; do
    if curl -s --connect-timeout 5 http://localhost:3000 > /dev/null 2>&1; then
        echo "✅ 前端服务启动成功"
        break
    else
        if [ $i -eq 10 ]; then
            echo "❌ 前端服务启动可能失败，请检查日志: tail -f frontend.log"
        else
            echo "⏳ 前端服务启动中... ($i/10)"
            sleep 2
        fi
    fi
done

# 显示PID文件内容
if [ -f "pids.txt" ]; then
    echo ""
    echo "📋 已记录的进程PID:"
    cat pids.txt | while read pid; do
        if [ ! -z "$pid" ]; then
            echo "  - PID: $pid"
        fi
    done
fi

echo ""
echo "🎉 启动完成！"
echo "📱 前端地址: http://localhost:3000"
echo "🔧 后端地址: http://localhost:8080"
echo "🔍 后端健康检查: http://localhost:8080/health"
echo ""
echo "📋 常用命令："
echo "   ⏹️  停止服务: ./stop.sh"
echo "   📊 查看进程: ps aux | grep -E '(go run|npm run|vite)'"
echo "   🔍 查看端口: lsof -i :8080,3000"
echo "   📝 查看日志: tail -f backend.log 或 tail -f frontend.log"
echo ""
echo "💡 如遇到问题，请检查："
echo "   1. 端口8080和3000是否被占用"
echo "   2. Go和Node.js环境是否正确安装"
echo "   3. 依赖包是否已安装 (go mod tidy & npm install)"
echo "   4. 查看启动日志文件" 