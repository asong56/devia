@echo off
setlocal enabledelayedexpansion
if not exist build mkdir build

set CGO_ENABLED=0

echo ==^> windows/amd64 (full: CLI + serve)
set GOOS=windows
set GOARCH=amd64
go build -trimpath -ldflags="-s -w" -o build\devia-windows-amd64.exe .

echo ==^> windows/amd64 (cli-only, -tags noserve)
go build -tags noserve -trimpath -ldflags="-s -w" -o build\devia-cli-windows-amd64.exe .

echo ==^> windows/arm64 (full: CLI + serve)
set GOARCH=arm64
go build -trimpath -ldflags="-s -w" -o build\devia-windows-arm64.exe .

echo ==^> windows/arm64 (cli-only, -tags noserve)
go build -tags noserve -trimpath -ldflags="-s -w" -o build\devia-cli-windows-arm64.exe .

echo.
dir build
endlocal
