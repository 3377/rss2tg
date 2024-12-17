#!/bin/sh

# 确保脚本在出错时退出
set -e

echo "正在停止所有相关容器..."
# 停止所有相关容器
docker-compose -f docker-compose.yml down 2>/dev/null || true
docker-compose -f docker-compose.dev.yml down --remove-orphans 2>/dev/null || true

echo "正在清理构建缓存..."
# 清理构建缓存
docker builder prune -f 2>/dev/null || true

echo "正在构建并启动开发容器..."
# 构建并启动开发容器
docker-compose -f docker-compose.dev.yml up --build

# 捕获Ctrl+C信号，优雅退出
trap 'docker-compose -f docker-compose.dev.yml down' INT TERM