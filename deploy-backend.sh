#!/usr/bin/env bash
set -e

cd /opt/gamepulse

echo "拉取最新代码..."
git pull

echo "重新构建并启动后端..."
docker compose up -d --build gamepulse_app

echo "后端容器状态："
docker compose ps

echo "最近后端日志："
docker logs --tail=80 gamepulse-app

echo "后端部署完成。"