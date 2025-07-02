#!/bin/bash
BACKUP_DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="../backups/backup_${BACKUP_DATE}"

echo "💾 开始备份 EmailAlert 数据..."
mkdir -p "$BACKUP_DIR"

# 备份数据库
if [ -f "backend/data/emailalert.db" ]; then
    cp backend/data/emailalert.db "$BACKUP_DIR/"
    echo "✅ 数据库备份完成"
fi

# 备份配置文件
cp -r backend/config "$BACKUP_DIR/"
echo "✅ 配置文件备份完成"

# 备份日志文件(最近7天)
find backend/logs -name "*.log" -mtime -7 -exec cp {} "$BACKUP_DIR/" \;
echo "✅ 日志文件备份完成"

echo "💾 备份完成: $BACKUP_DIR"
