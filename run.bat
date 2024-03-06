@echo off
REM 清理工作和退出时的处理
trap "del server.exe" EXIT

REM 编译 Go 代码并启动服务器实例
go build -o server.exe
start "" server.exe -port=8001
start "" server.exe -port=8002
start "" server.exe -port=8003 -api=1

REM 等待 2 秒以确保服务器已启动
ping 127.0.0.1 -n 2 > nul

echo ">>> start test"
REM 使用 curl 发送请求
curl "http://localhost:8080/api?key=Tom"
curl "http://localhost:8080/api?key=Tom"
curl "http://localhost:8080/api?key=Tom"

REM 等待所有子进程结束
echo "Waiting for all subprocesses to finish"
pause
