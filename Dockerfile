# 构建阶段
FROM golang:alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o gamepulse_app .

# 运行阶段
FROM debian:bookworm-slim

WORKDIR /app

# 安装 ca-certificates 供 HTTPS API 调用使用，安装 netcat 供 wait-for.sh 检查依赖端口
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates netcat-openbsd; \
    rm -rf /var/lib/apt/lists/*

COPY wait-for.sh ./wait-for.sh
COPY templates ./templates
COPY assets ./assets
COPY --from=builder /src/gamepulse_app ./gamepulse_app

# conf/config.docker.yaml 由 docker-compose 运行时挂载，避免把生产密钥打进镜像
RUN mkdir -p ./conf && sed -i 's/\r$//' ./wait-for.sh && chmod 755 ./wait-for.sh

EXPOSE 8080

# 默认命令保留给直接 docker run；docker-compose 会传入 config.docker.yaml
CMD ["./gamepulse_app"]
