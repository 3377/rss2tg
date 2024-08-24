#!/bin/sh

if [ ! -f /app/config/config.yaml ]; then
    echo "未找到配置文件。使用环境变量进行初始配置。"
fi

exec "/app/bot"
