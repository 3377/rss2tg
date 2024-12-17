#!/bin/sh

# 确保配置目录存在
mkdir -p /app/config /app/data

# 检查配置文件
if [ ! -f /app/config/config.yaml ]; then
    echo "未找到配置文件。使用环境变量进行初始配置。"
fi

# 确保bot有执行权限
chmod +x /app/bot

# 执行bot
exec "/app/bot"
