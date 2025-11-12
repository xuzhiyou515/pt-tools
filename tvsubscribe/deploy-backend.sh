#!/bin/bash

# TVSubscribe后端部署脚本
# 作者: 自动生成
# 描述: 构建并部署tvsubscribe后端

set -e  # 遇到错误时退出

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "开始部署TVSubscribe后端..."

# 1. 在tvsubscribe/cmd目录下构建Go程序
echo "步骤1: 构建Go后端程序..."
cd "$SCRIPT_DIR/cmd"
go build -o tvsubscribe .
if [ $? -eq 0 ]; then
    echo "✓ Go程序构建成功"
else
    echo "✗ Go程序构建失败"
    exit 1
fi

# 2. 停止tvsubscribe服务
echo "步骤2: 停止tvsubscribe服务..."
sudo systemctl stop tvsubscribe
if [ $? -eq 0 ]; then
    echo "✓ 服务已停止"
else
    echo "⚠ 服务停止失败或服务未运行，继续部署..."
fi

# 3. 复制生成的程序到目标目录
echo "步骤3: 复制后端程序..."
sudo cp "$SCRIPT_DIR/cmd/tvsubscribe" /opt/tvsubscribe/
sudo chown tvsubscribe:tvsubscribe /opt/tvsubscribe/tvsubscribe
echo "✓ 后端程序已复制并设置权限"

# 4. 启动tvsubscribe服务
echo "步骤4: 启动tvsubscribe服务..."
sudo systemctl start tvsubscribe
sleep 2  # 等待服务启动

# 5. 检查服务状态
if sudo systemctl is-active --quiet tvsubscribe; then
    echo "✓ 服务启动成功"
    echo "后端部署完成！TVSubscribe服务正在运行"
else
    echo "✗ 服务启动失败"
    echo "请检查服务日志: sudo journalctl -u tvsubscribe -f"
    exit 1
fi

echo "后端部署成功完成！"