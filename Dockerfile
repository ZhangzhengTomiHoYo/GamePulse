# 构建阶段
FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o gamepulse_app .

# 运行阶段
FROM debian:bookworm-slim

# 建议：设置时区为上海（可选，方便看日志）
# RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY ./wait-for.sh /
COPY ./templates /templates
COPY ./assets /assets
COPY ./conf /conf

COPY --from=builder /build/gamepulse_app /

# 核心修正：增加 sed -i 这一行，强制把 windows 换行符转为 linux 换行符
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends netcat-openbsd; \
    sed -i 's/\r$//' /wait-for.sh; \
    chmod 755 /wait-for.sh; \
    rm -rf /var/lib/apt/lists/*

# 这里的 ENTRYPOINT 或 CMD 由 docker-compose 接管