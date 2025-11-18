#!/bin/bash

# Mattermost 服务器初始化脚本
# 用途: 第一次部署时，在服务器上设置完整的环境

set -e  # 遇到错误立即退出

# 配置
SERVER_IP="8.218.215.103"
SERVER_USER="root"
SERVER_PATH="/opt/mattermost"

echo "=========================================="
echo "Mattermost 服务器初始化脚本"
echo "=========================================="
echo ""

# 1. 创建目录结构
echo "📁 步骤 1/4: 创建目录结构..."
ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
mkdir -p /opt/mattermost/server
mkdir -p /opt/mattermost/enterprise
mkdir -p /opt/mattermost/webapp
echo "✅ 目录结构已创建"
EOF

# 2. 创建 enterprise placeholder
echo ""
echo "🏢 步骤 2/4: 创建 Enterprise placeholder..."
ssh ${SERVER_USER}@${SERVER_IP} << 'ENTERPRISE_EOF'
cd /opt/mattermost/enterprise

# 初始化 git 仓库
git init

# 创建 placeholder.go
cat > placeholder.go << 'EOF'
// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.enterprise for license information.

// Ensure this is a valid package even when build tags preclude building anything in it.
package enterprise

EOF

# 创建 go.mod
cat > go.mod << 'EOF'
module github.com/mattermost/mattermost/server/v8/enterprise

go 1.24.6

require github.com/mattermost/mattermost/server/v8 v0.0.0

replace github.com/mattermost/mattermost/server/v8 => ../server

EOF

echo "✅ Enterprise placeholder 已创建"
ls -la
ENTERPRISE_EOF

# 3. 同步代码
echo ""
echo "📦 步骤 3/4: 同步代码到服务器..."
echo "正在同步 server 目录..."
rsync -avz --progress \
  --exclude 'node_modules' \
  --exclude '.git' \
  --exclude 'bin' \
  --exclude 'logs' \
  --exclude 'data' \
  --exclude 'plugins' \
  /Users/arthur/RustroverProjects/mattermost/server/ ${SERVER_USER}@${SERVER_IP}:${SERVER_PATH}/server/

echo ""
echo "正在同步 webapp 目录..."
rsync -avz --progress \
  --exclude 'node_modules' \
  --exclude '.git' \
  --exclude 'dist' \
  --exclude 'build' \
  /Users/arthur/RustroverProjects/mattermost/webapp/ ${SERVER_USER}@${SERVER_IP}:${SERVER_PATH}/webapp/

# 4. 设置 Go workspace
echo ""
echo "🔨 步骤 4/4: 设置 Go workspace..."
ssh ${SERVER_USER}@${SERVER_IP} << 'EOF'
cd /opt/mattermost/server

# 创建 go.work 文件
echo "创建 go.work 文件..."
go work init
go work use .
go work use ./public
go work use ../enterprise

echo ""
echo "✅ go.work 文件内容:"
cat go.work

echo ""
echo "验证 Go 模块..."
go mod download || echo "⚠️  某些依赖可能需要在容器中下载"
EOF

echo ""
echo "=========================================="
echo "✅ 服务器初始化完成!"
echo "=========================================="
echo ""
echo "下一步:"
echo "  1. 运行 ./deploy-to-server.sh 启动服务"
echo "  2. 或者手动 SSH 到服务器: ssh ${SERVER_USER}@${SERVER_IP}"
echo "  3. 进入目录: cd ${SERVER_PATH}/server"
echo "  4. 启动服务: docker compose up -d"
echo ""
