@echo off
chcp 65001 >nul
title Meow Music Server - 启动中...

echo.
echo ╔═══════════════════════════════════════════════╗
echo ║   🎵 Meow Embedded Music Server v2.0         ║
echo ║   喵波音律 - 专为ESP32设计的音乐服务器          ║
echo ╚═══════════════════════════════════════════════╝
echo.

REM 检查Go是否安装
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未找到Go语言环境！
    echo.
    echo 请先安装 Go: https://golang.org/dl/
    echo.
    pause
    exit /b 1
)

echo [✓] Go环境检测成功
go version
echo.

echo [步骤 1/3] 检查配置文件...
if not exist .env (
    if exist .env.example (
        echo [提示] 未找到.env文件，正在从.env.example创建...
        copy .env.example .env >nul
        echo [✓] 配置文件已创建
    ) else (
        echo [提示] 将使用默认配置
    )
) else (
    echo [✓] 配置文件已存在
)
echo.

echo [步骤 2/3] 安装/更新依赖...
go mod tidy
if %errorlevel% neq 0 (
    echo [错误] 依赖安装失败！
    echo.
    echo 可能需要配置Go代理（中国大陆用户）：
    echo go env -w GOPROXY=https://goproxy.cn,direct
    echo.
    pause
    exit /b 1
)
echo [✓] 依赖安装完成
echo.

echo [步骤 3/3] 启动服务器...
echo.
echo ╔═══════════════════════════════════════════════╗
echo ║                 访问地址                       ║
echo ╠═══════════════════════════════════════════════╣
echo ║  新版应用: http://localhost:2233/app          ║
echo ║  经典界面: http://localhost:2233/             ║
echo ╚═══════════════════════════════════════════════╝
echo.
echo [✓] 服务器正在启动，请稍候...
echo.
echo ┌───────────────────────────────────────────────┐
echo │  提示：按 Ctrl+C 可以停止服务器                │
echo │  首次使用请访问 /app 注册账户                 │
echo └───────────────────────────────────────────────┘
echo.

title Meow Music Server - 运行中
go run .
