#!/usr/bin/env bash
set -e

cd /opt/gamepulse

echo "拉取最新代码..."
git pull

echo "打包前端..."
cd /opt/gamepulse/frontend

docker run --rm \
  -v "$PWD":/app \
  -w /app \
  node:20-alpine \
  sh -c "npm config set registry https://registry.npmmirror.com && npm install -g pnpm && pnpm config set registry https://registry.npmmirror.com && pnpm install --frozen-lockfile && pnpm run build"

echo "复制 dist 到 Nginx 目录..."
sudo rm -rf /var/www/gamepulse/*
sudo cp -r dist/* /var/www/gamepulse/

echo "重载 Nginx..."
sudo systemctl reload nginx

echo "前端部署完成。"