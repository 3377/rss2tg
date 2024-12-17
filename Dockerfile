# 构建阶段
FROM golang:1.19-alpine AS builder

WORKDIR /build

# 安装构建依赖
RUN apk add --no-cache gcc musl-dev

# 复制并下载依赖
COPY go.mod ./
RUN go mod download

# 复制源代码并构建
COPY . .
# 确保version文件存在并复制到正确位置
RUN mkdir -p config && \
    if [ -f version ]; then \
    cp version config/version; \
    else \
    echo "1.0.0" > config/version; \
    fi && \
    go mod tidy && \
    go build -ldflags="-w -s" -o bot

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 创建必要的目录并复制文件
RUN mkdir -p /app/config /app/data

# 从builder阶段复制文件
COPY --from=builder /build/bot /app/bot
COPY --from=builder /build/config/version /app/config/version
COPY entrypoint.sh /app/entrypoint.sh

# 设置文件权限
RUN chmod +x /app/entrypoint.sh /app/bot && \
    # 确保version文件存在且有正确的权限
    touch /app/config/version && \
    chmod 644 /app/config/version

# 安装运行时依赖和时区数据
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 使用shell形式的ENTRYPOINT以确保正确的权限和路径
ENTRYPOINT ["/bin/sh", "-c", "exec /app/entrypoint.sh"]
CMD ["/app/bot"]
