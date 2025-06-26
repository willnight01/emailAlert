#!/bin/bash

echo "🛑 停止邮件告警平台..."

# 读取PID文件并停止进程
if [ -f "pids.txt" ]; then
    while read pid; do
        if [ ! -z "$pid" ] && kill -0 $pid 2>/dev/null; then
            kill $pid
            echo "✅ 已停止进程 PID: $pid"
        fi
    done < pids.txt
    rm pids.txt
else
    echo "⚠️  未找到PID文件，尝试按端口停止..."
    
    # 停止占用3000端口的进程（前端）
    PID_3000=$(lsof -ti:3000)
    if [ ! -z "$PID_3000" ]; then
        kill $PID_3000
        echo "✅ 已停止前端服务 (端口3000)"
    fi
    
    # 停止占用8080端口的进程（后端）
    PID_8080=$(lsof -ti:8080)
    if [ ! -z "$PID_8080" ]; then
        kill $PID_8080
        echo "✅ 已停止后端服务 (端口8080)"
    fi
fi

echo "🎉 所有服务已停止！" 