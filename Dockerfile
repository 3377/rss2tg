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
RUN go mod tidy
RUN go build -ldflags="-w -s" -o bot

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /build/bot /app/bot

# 创建必要的目录
RUN mkdir -p /app/config /app/data

# 复制version文件到config目录（从构建阶段复制）
COPY --from=builder /build/config/version /app/config/version

# 复制entrypoint脚本
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# 安装运行时依赖和时区数据
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/bot"]
