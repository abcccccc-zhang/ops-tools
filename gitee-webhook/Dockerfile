# 第一阶段：构建阶段
FROM golang:1.24-alpine AS builder

# 安装依赖
#RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 拷贝源码
COPY . .

# 编译 Go 程序（关闭 CGO，生成静态链接的 Linux 可执行文件）
RUN go build -o gitee-webhook main.go

# 第二阶段：运行阶段，干净精简的镜像
FROM alpine:latest

# 安装基础运行库（如果需要 curl 等工具可以加上）
RUN apk add --no-cache ca-certificates

WORKDIR /app

# 拷贝编译好的二进制文件
COPY --from=builder /app/gitee-webhook .

# 暴露端口（比如监听 50000）
EXPOSE 7878

# 启动命令
CMD ["./gitee-webhook"]
