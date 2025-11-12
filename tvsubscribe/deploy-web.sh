#!/bin/bash

# TVSubscribe前端部署脚本
# 作者: 自动生成
# 描述: 构建并部署tvsubscribe前端

set -e  # 遇到错误时退出

# 用户配置目录
USER_CONFIG_DIR="$HOME/.config/tvsubscribe"

# 确保用户配置目录存在
mkdir -p "$USER_CONFIG_DIR"

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "开始部署TVSubscribe前端..."

# 1. 在tvsubscribe/web目录下构建前端
echo "步骤1: 构建前端程序..."
cd "$SCRIPT_DIR/web"
npm run build
if [ $? -eq 0 ]; then
    echo "✓ 前端构建成功"
else
    echo "✗ 前端构建失败"
    exit 1
fi

# 2. 删除历史网页文件
echo "步骤2: 清理旧的网页文件..."
rm -rf "$USER_CONFIG_DIR/web"/*
echo "✓ 旧网页文件已清理"

# 3. 复制dist目录到用户目录
echo "步骤3: 复制前端文件到用户目录..."
mkdir -p "$USER_CONFIG_DIR/web"
cp -r "$SCRIPT_DIR/web/dist" "$USER_CONFIG_DIR/web/"
echo "✓ 前端文件已复制到用户目录"

# 4. 检查文件是否正确复制
echo "步骤4: 验证部署..."
if [ -f "$USER_CONFIG_DIR/web/dist/index.html" ]; then
    echo "✓ 前端部署验证成功"
    echo "前端部署完成！网页文件位于 $USER_CONFIG_DIR/web/dist/"
else
    echo "✗ 前端部署验证失败"
    exit 1
fi

echo "前端部署成功完成！"