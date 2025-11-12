#!/bin/bash

# TVSubscribe部署脚本
# 作者: 自动生成
# 描述: 构建并部署tvsubscribe服务

set -e  # 遇到错误时退出

# 用户配置目录
USER_CONFIG_DIR="$HOME/.config/tvsubscribe"

# 确保用户配置目录存在
mkdir -p "$USER_CONFIG_DIR"

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "开始部署TVSubscribe..."
echo "脚本目录: $SCRIPT_DIR"

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

# 2. 在tvsubscribe/web目录下构建前端
echo "步骤2: 构建前端程序..."
cd "$SCRIPT_DIR/web"
npm run build
if [ $? -eq 0 ]; then
    echo "✓ 前端构建成功"
else
    echo "✗ 前端构建失败"
    exit 1
fi

# 3. 停止tvsubscribe服务
echo "步骤3: 停止tvsubscribe服务..."
systemctl --user stop tvsubscribe
if [ $? -eq 0 ]; then
    echo "✓ 服务已停止"
else
    echo "⚠ 服务停止失败或服务未运行，继续部署..."
fi

# 4. 删除历史网页文件
echo "步骤4: 清理旧的网页文件..."
rm -rf "$USER_CONFIG_DIR/web"/*
echo "✓ 旧网页文件已清理"

# 5. 复制生成的文件到用户目录
echo "步骤5: 复制新文件到用户目录..."
# 复制Go程序
cp "$SCRIPT_DIR/cmd/tvsubscribe" "$USER_CONFIG_DIR/"
echo "✓ 后端程序已复制到用户目录"

# 复制前端文件到用户目录
mkdir -p "$USER_CONFIG_DIR/web"
cp -r "$SCRIPT_DIR/web/dist" "$USER_CONFIG_DIR/web/"
echo "✓ 前端文件已复制到用户目录"

# 6. 启动tvsubscribe服务
echo "步骤6: 启动tvsubscribe服务..."
systemctl --user start tvsubscribe
sleep 2  # 等待服务启动

# 检查服务状态
if systemctl --user is-active --quiet tvsubscribe; then
    echo "✓ 服务启动成功"
    echo "部署完成！TVSubscribe服务正在运行"
else
    echo "✗ 服务启动失败"
    echo "请检查服务日志: journalctl --user -u tvsubscribe -f"
    exit 1
fi

echo "部署成功完成！"