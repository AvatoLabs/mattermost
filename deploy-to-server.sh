#!/bin/bash

# Mattermost 服务器部署脚本
# 用途: 将本地代码同步到远程服务器并启动 HA 集群
# 仓库: https://github.com/AvatoLabs/mattermost

set -e  # 遇到错误立即退出

# 配置
SERVER_IP="8.218.215.103"
SERVER_USER="root"
SERVER_PATH="/opt/mattermost"
LOCAL_PATH="/Users/arthur/RustroverProjects/mattermost"

echo "=========================================="
echo "Mattermost 服务器部署脚本"
echo "=========================================="
echo ""

# 1. 同步代码到服务器
echo "📦 步骤 1/5: 同步代码到服务器..."
echo "正在同步 server 目录..."
rsync -avz --progress \
  --exclude 'node_modules' \
  --exclude '.git' \
  --exclude 'bin' \
  --exclude 'logs' \
  --exclude 'data' \
  --exclude 'plugins' \
  "${LOCAL_PATH}/server/" "${SERVER_USER}@${SERVER_IP}:${SERVER_PATH}/server/"

echo ""
echo "正在同步 enterprise 目录..."
rsync -avz --progress \
  --exclude 'node_modules' \
  --exclude '.git' \
  "${LOCAL_PATH}/enterprise/" "${SERVER_USER}@${SERVER_IP}:${SERVER_PATH}/enterprise/"

echo ""
echo "正在同步 webapp 目录..."
rsync -avz --progress \
  --exclude 'node_modules' \
  --exclude '.git' \
  --exclude 'dist' \
  --exclude 'build' \
  "${LOCAL_PATH}/webapp/" "${SERVER_USER}@${SERVER_IP}:${SERVER_PATH}/webapp/"

# 2. 修复 docker-compose.yaml 中的命令
echo ""
echo "🔧 步骤 2/5: 修复 docker-compose.yaml 配置..."
ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd /opt/mattermost/server

# 将 build-server 替换为 run-server
if grep -q "build-server" docker-compose.yaml; then
  echo "修复 docker-compose.yaml 中的命令..."
  sed -i "s/command: \['make', 'build-server'\]/command: ['make', 'run-server']/g" docker-compose.yaml
  echo "✅ 已将 build-server 替换为 run-server"
else
  echo "✅ docker-compose.yaml 已经是正确的配置"
fi
EOF

# 3. 设置 go.work 文件
echo ""
echo "🔨 步骤 3/5: 设置 Go workspace..."
ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd /opt/mattermost/server

# 检查 go.work 是否存在
if [ ! -f go.work ]; then
  echo "创建 go.work 文件..."
  go work init
  go work use .
  go work use ./public
  go work use ../enterprise
  echo "✅ go.work 文件已创建"
else
  echo "✅ go.work 文件已存在"
fi
EOF

# 4. 停止现有容器
echo ""
echo "🛑 步骤 4/5: 停止现有容器..."
ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd /opt/mattermost/server
docker compose down
echo "✅ 容器已停止"
EOF

# 5. 启动服务
echo ""
echo "🚀 步骤 5/5: 启动 Mattermost HA 集群..."
ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd /opt/mattermost/server

# 设置 CURRENT_UID 环境变量
export CURRENT_UID=$(id -u):$(id -g)

# 启动服务
echo "启动 docker compose..."
docker compose up -d

echo ""
echo "等待 10 秒让服务启动..."
sleep 10

echo ""
echo "📊 容器状态:"
docker compose ps

echo ""
echo "📝 查看 leader 容器日志 (最后 20 行):"
docker logs server-leader-1 --tail 20 2>&1 || echo "leader 容器尚未创建"
EOF

echo ""
echo "=========================================="
echo "✅ 部署完成!"
echo "=========================================="
echo ""
echo "访问地址: http://${SERVER_IP}:8065"
echo ""
echo "常用命令:"
echo "  查看日志: ssh ${SERVER_USER}@${SERVER_IP} 'cd ${SERVER_PATH}/server && docker compose logs -f'"
echo "  查看状态: ssh ${SERVER_USER}@${SERVER_IP} 'cd ${SERVER_PATH}/server && docker compose ps'"
echo "  重启服务: ssh ${SERVER_USER}@${SERVER_IP} 'cd ${SERVER_PATH}/server && docker compose restart'"
echo "  停止服务: ssh ${SERVER_USER}@${SERVER_IP} 'cd ${SERVER_PATH}/server && docker compose down'"
echo ""
