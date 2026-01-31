FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    COOS=linux \
    GOARCH=amd64

# 移动到工作目录
WORKDIR /build

# 复制项目中的 go.mod 和 go.sum文件并下载依赖
COPY go.mod .
COPY go.sum .
RUN go mod download

# 将代码复制到容器中
COPY . .

# 将代码编译成可执行的二进制文件
RUN go build -o bluebell_app .

#########
# 接下来创建一个小镜像
#########
FROM scratch

COPY ./templates /templates
COPY ./assets /assets
COPY ./conf /conf

# 从builder 镜像中把/dist/app拷贝到当前目录
COPY --from=builder /build/bluebell_app /

# 需要运行的命令
ENTRYPOINT ["/bluebell_app"]