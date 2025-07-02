#!/bin/bash
if [ "$1" = "backend" ]; then
    echo "📋 查看后端日志..."
    docker compose logs -f backend
elif [ "$1" = "frontend" ]; then
    echo "📋 查看前端日志..."
    docker compose logs -f frontend
else
    echo "📋 查看所有日志..."
    docker compose logs -f
fi
