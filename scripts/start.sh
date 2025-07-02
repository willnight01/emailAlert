#!/bin/bash
echo "🚀 启动 EmailAlert 服务..."
docker compose up -d
echo "✅ 服务启动完成"
echo "📱 前端地址: http://localhost:3000"
echo "🔧 后端API: http://localhost:8080"
