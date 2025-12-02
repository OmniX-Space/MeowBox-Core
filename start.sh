#!/bin/bash

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo ""
echo "╔═══════════════════════════════════════════════╗"
echo "║   🎵 Meow Embedded Music Server v2.0         ║"
echo "║   喵波音律 - 专为ESP32设计的音乐服务器          ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo -e "${RED}[错误]${NC} 未找到Go语言环境！"
    echo ""
    echo "请先安装 Go:"
    echo "  Ubuntu/Debian: sudo apt install golang-go"
    echo "  macOS:         brew install go"
    echo "  CentOS/RHEL:   sudo yum install golang"
    echo ""
    exit 1
fi

echo -e "${GREEN}[✓]${NC} Go环境检测成功"
go version
echo ""

# 检查配置文件
echo -e "${BLUE}[步骤 1/3]${NC} 检查配置文件..."
if [ ! -f .env ]; then
    if [ -f .env.example ]; then
        echo -e "${YELLOW}[提示]${NC} 未找到.env文件，正在从.env.example创建..."
        cp .env.example .env
        echo -e "${GREEN}[✓]${NC} 配置文件已创建"
    else
        echo -e "${YELLOW}[提示]${NC} 将使用默认配置"
    fi
else
    echo -e "${GREEN}[✓]${NC} 配置文件已存在"
fi
echo ""

# 安装依赖
echo -e "${BLUE}[步骤 2/3]${NC} 安装/更新依赖..."
if ! go mod tidy; then
    echo -e "${RED}[错误]${NC} 依赖安装失败！"
    echo ""
    echo "可能需要配置Go代理（中国大陆用户）："
    echo "  export GOPROXY=https://goproxy.cn,direct"
    echo "或永久设置："
    echo "  go env -w GOPROXY=https://goproxy.cn,direct"
    echo ""
    exit 1
fi
echo -e "${GREEN}[✓]${NC} 依赖安装完成"
echo ""

# 启动服务器
echo -e "${BLUE}[步骤 3/3]${NC} 启动服务器..."
echo ""
echo "╔═══════════════════════════════════════════════╗"
echo "║                 访问地址                       ║"
echo "╠═══════════════════════════════════════════════╣"
echo "║  新版应用: http://localhost:2233/app          ║"
echo "║  经典界面: http://localhost:2233/             ║"
echo "╚═══════════════════════════════════════════════╝"
echo ""
echo -e "${GREEN}[✓]${NC} 服务器正在启动，请稍候..."
echo ""
echo "┌───────────────────────────────────────────────┐"
echo "│  提示：按 Ctrl+C 可以停止服务器                │"
echo "│  首次使用请访问 /app 注册账户                 │"
echo "└───────────────────────────────────────────────┘"
echo ""

go run .
