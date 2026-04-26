FROM golang:alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=sum.golang.google.cn

WORKDIR /src

RUN sed -i 's|https://dl-cdn.alpinelinux.org/alpine|https://mirrors.cloud.tencent.com/alpine|g' /etc/apk/repositories \
    && apk add --no-cache git ca-certificates tzdata \
    && update-ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o gamepulse_app .

FROM debian:bookworm-slim

WORKDIR /app

RUN set -eux; \
    sed -i 's|http://deb.debian.org/debian|http://mirrors.cloud.tencent.com/debian|g; s|http://security.debian.org/debian-security|http://mirrors.cloud.tencent.com/debian-security|g' /etc/apt/sources.list.d/debian.sources || true; \
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates netcat-openbsd; \
    rm -rf /var/lib/apt/lists/*

COPY wait-for.sh ./wait-for.sh
COPY templates ./templates
COPY assets ./assets
COPY conf ./conf
COPY --from=builder /src/gamepulse_app ./gamepulse_app

RUN sed -i 's/\r$//' ./wait-for.sh && chmod 755 ./wait-for.sh

EXPOSE 8080

CMD ["./gamepulse_app"]