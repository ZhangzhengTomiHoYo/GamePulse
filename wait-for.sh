#!/bin/sh
set -eu

TIMEOUT="${WAIT_FOR_TIMEOUT:-60}"
INTERVAL="${WAIT_FOR_INTERVAL:-1}"

targets=""
cmd=""

while [ "$#" -gt 0 ]; do
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

if [ -z "$targets" ] || [ -z "$cmd" ]; then
    echo "Usage: $0 host1:port1 host2:port2 -- command [args...]"
    exit 1
fi

wait_for_target() {
    target="$1"
    host="$(echo "$target" | cut -d: -f1)"
    port="$(echo "$target" | cut -d: -f2)"

    echo "Waiting for $host:$port ..."
    nc -z -w 1 "$host" "$port"
}

elapsed=0
while [ "$elapsed" -lt "$TIMEOUT" ]; do
    all_ready=1

    for target in $targets; do
        if ! wait_for_target "$target"; then
            all_ready=0
            break
        fi
    done

    if [ "$all_ready" -eq 1 ]; then
        echo "Dependencies are ready. Starting application..."
        exec $cmd
    fi

    sleep "$INTERVAL"
    elapsed=$((elapsed + INTERVAL))
done

echo "Timed out waiting for dependencies:$targets"
exit 1
