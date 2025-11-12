#!/bin/bash

echo "=== TVSubscribe 管理界面启动脚本 ==="

# 检查是否已安装依赖
if [ ! -d "node_modules" ]; then
    echo "正在安装依赖..."
    npm install
    if [ $? -ne 0 ]; then
        echo "依赖安装失败，请检查网络连接和 Node.js 环境"
        exit 1
    fi
fi

echo "启动开发服务器..."
echo "管理界面将在 http://localhost:3000 启动"
echo "API请求将代理到 http://localhost:8443"
echo "请确保 tvsubscribe 后端服务已启动"
echo ""
echo "按 Ctrl+C 停止服务器"
echo ""

npm run dev