#!/bin/sh
set -eu  # 删掉 -o pipefail，其余内容完全不变

# 脚本核心功能：等待多个TCP服务端口就绪，再执行后续启动命令
# 使用方式：./wait-for.sh 服务1:端口 服务2:端口 -- 应用启动命令

# 超时时间（秒），可根据需要调整（默认60秒足够MySQL/Redis启动）
TIMEOUT=60
# 检测间隔（秒），每次检测失败后等待1秒再重试
INTERVAL=1
# 记录已等待的时间
elapsed=0

# 解析参数：-- 前面是待检测的「主机:端口」，后面是要执行的应用启动命令
while [ $# -gt 0 ]; do
    case "$1" in
        --)
            shift
            cmd="$*"
            break
            ;;
        *)
            targets="$targets $1"
            shift
            ;;
    esac
done

# 检查是否传入了待检测目标和启动命令
if [ -z "$targets" ] || [ -z "$cmd" ]; then
    echo "用法错误！正确格式：$0 主机1:端口1 主机2:端口2 -- 应用启动命令"
    echo "示例：$0 mysql:3306 redis:6379 -- ./bluebell_app ./conf/config.yaml"
    exit 1
fi

# 定义单个服务检测函数（使用netcat检测TCP端口是否通）
wait_for_target() {
    local target="$1"
    local host=$(echo "$target" | cut -d: -f1)
    local port=$(echo "$target" | cut -d: -f2)

    echo "正在等待服务就绪：$host:$port"
    # 使用netcat检测端口（-z 扫描端口不发送数据；-w1 连接超时1秒；-v 详细输出）
    if nc -zvw1 "$host" "$port"; then
        echo "服务就绪：$host:$port"
        return 0
    else
        return 1
    fi
}

# 循环检测所有目标，直到全部就绪或超时
while [ $elapsed -lt $TIMEOUT ]; do
    local all_ready=1
    # 遍历所有待检测的「主机:端口」
    for target in $targets; do
        if ! wait_for_target "$target"; then
            all_ready=0
            break
        fi
    done
    # 所有服务就绪，跳出循环执行启动命令
    if [ $all_ready -eq 1 ]; then
        echo "所有依赖服务已就绪，启动应用..."
        exec $cmd
    fi
    # 未就绪则等待1秒，累计超时时间
    sleep $INTERVAL
    elapsed=$((elapsed + INTERVAL))
done

# 超时未就绪，输出错误并退出
echo "错误：等待服务超时（超时时间：$TIMEOUT 秒），以下服务未就绪：$targets"
exit 1