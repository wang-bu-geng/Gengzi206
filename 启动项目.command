#!/bin/bash

cd "$(dirname "$0")"
PROJECT_ROOT="$(pwd)"
BACKEND_DIR="$PROJECT_ROOT"
FRONTEND_DIR="$PROJECT_ROOT/web"

echo "=========================================="
echo "  更子206 - 漫剧项目启动器"
echo "=========================================="
echo ""
echo "项目路径: $PROJECT_ROOT"
echo ""

echo "正在启动后端服务..."
osascript -e "tell application \"Terminal\" to do script \"cd '$BACKEND_DIR' && go run main.go\""

sleep 5

echo "正在启动前端服务..."
osascript -e "tell application \"Terminal\" to do script \"cd '$FRONTEND_DIR' && npm run dev\""

echo ""
echo "✅ 启动命令已执行！"
echo "📌 后端：Go服务"
echo "📌 前端：访问 http://localhost:5678"
echo "🔧 停止服务：关闭对应的终端窗口即可"
echo ""
echo "按回车键退出..."
read
