#!/bin/bash

echo "🛑 停止邮件告警平台..."

# 函数：强制停止进程
force_kill_process() {
    local pid=$1
    if [ ! -z "$pid" ] && kill -0 $pid 2>/dev/null; then
        echo "⏹️  正在停止进程 PID: $pid"
        kill $pid 2>/dev/null
        sleep 2
        
        # 如果进程仍然存在，强制杀死
        if kill -0 $pid 2>/dev/null; then
            echo "🔨 强制停止进程 PID: $pid"
            kill -9 $pid 2>/dev/null
            sleep 1
        fi
        
        # 最终检查
        if kill -0 $pid 2>/dev/null; then
            echo "❌ 无法停止进程 PID: $pid"
            return 1
        else
            echo "✅ 已停止进程 PID: $pid"
            return 0
        fi
    fi
    return 0
}

# 1. 通过PID文件停止服务
if [ -f "pids.txt" ]; then
    echo "📄 发现PID文件，正在停止服务..."
    while read pid; do
        if [ ! -z "$pid" ]; then
            force_kill_process $pid
        fi
    done < pids.txt
    rm -f pids.txt
    echo "✅ PID文件处理完成"
else
    echo "⚠️  未找到PID文件"
fi

# 2. 停止所有相关的Go进程
echo "🔍 查找并停止Go相关进程..."
GO_PIDS=$(pgrep -f "go run main.go" 2>/dev/null || true)
if [ ! -z "$GO_PIDS" ]; then
    echo "发现Go进程: $GO_PIDS"
    for pid in $GO_PIDS; do
        force_kill_process $pid
    done
else
    echo "未发现Go进程"
fi

# 3. 停止所有相关的emailAlert进程
EMAILALERT_PIDS=$(pgrep -f "emailAlert" 2>/dev/null || true)
if [ ! -z "$EMAILALERT_PIDS" ]; then
    echo "🔍 发现emailAlert进程: $EMAILALERT_PIDS"
    for pid in $EMAILALERT_PIDS; do
        force_kill_process $pid
    done
fi

# 4. 停止所有相关的npm进程
echo "🔍 查找并停止npm相关进程..."
NPM_PIDS=$(pgrep -f "npm run dev" 2>/dev/null || true)
if [ ! -z "$NPM_PIDS" ]; then
    echo "发现npm进程: $NPM_PIDS"
    for pid in $NPM_PIDS; do
        force_kill_process $pid
    done
else
    echo "未发现npm进程"
fi

# 5. 停止所有相关的node进程（Vite开发服务器）
NODE_PIDS=$(pgrep -f "vite" 2>/dev/null || true)
if [ ! -z "$NODE_PIDS" ]; then
    echo "🔍 发现Vite进程: $NODE_PIDS"
    for pid in $NODE_PIDS; do
        force_kill_process $pid
    done
fi

# 6. 按端口停止进程（最终保障）
echo "🔍 检查端口占用情况..."

# 检查8080端口（后端）
BACKEND_PIDS=$(lsof -ti:8080 2>/dev/null || true)
if [ ! -z "$BACKEND_PIDS" ]; then
    echo "发现占用8080端口的进程: $BACKEND_PIDS"
    for pid in $BACKEND_PIDS; do
        force_kill_process $pid
    done
else
    echo "8080端口未被占用"
fi

# 检查3000端口（前端）
FRONTEND_PIDS=$(lsof -ti:3000 2>/dev/null || true)
if [ ! -z "$FRONTEND_PIDS" ]; then
    echo "发现占用3000端口的进程: $FRONTEND_PIDS"
    for pid in $FRONTEND_PIDS; do
        force_kill_process $pid
    done
else
    echo "3000端口未被占用"
fi

# 7. 最终验证
echo ""
echo "🔍 最终验证服务状态..."

# 检查端口是否还被占用
if lsof -i:8080 >/dev/null 2>&1; then
    echo "❌ 8080端口仍被占用"
    lsof -i:8080
else
    echo "✅ 8080端口已释放"
fi

if lsof -i:3000 >/dev/null 2>&1; then
    echo "❌ 3000端口仍被占用"
    lsof -i:3000
else
    echo "✅ 3000端口已释放"
fi

# 检查相关进程是否还在运行
REMAINING_PIDS=$(pgrep -f "(go run main.go|npm run dev|vite|emailAlert)" 2>/dev/null || true)
if [ ! -z "$REMAINING_PIDS" ]; then
    echo "⚠️  发现残留进程: $REMAINING_PIDS"
    echo "可以手动执行: kill -9 $REMAINING_PIDS"
else
    echo "✅ 所有相关进程已停止"
fi

echo ""
echo "🎉 停止脚本执行完成！"
echo "💡 如果仍有问题，可以手动检查："
echo "   📋 查看进程: ps aux | grep -E '(go run|npm run|vite|emailAlert)'"
echo "   🔍 查看端口: lsof -i :8080,3000" 