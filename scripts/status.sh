#!/bin/bash
echo "📊 EmailAlert 服务状态:"
echo "================================"
docker compose ps
echo ""
echo "📊 资源使用情况:"
docker compose top
