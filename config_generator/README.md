# 配置生成程序

## 编译

```bash
# Linux 64位
go build -o release/generate -ldflags="-s -w" main.go

# Windows 64位
GOOS=windows GOARCH=amd64 go build -o release/generate.exe -ldflags="-s -w" main.go

# macOS 64位 (Intel)
GOOS=darwin GOARCH=amd64 go build -o release/generate -ldflags="-s -w" main.go

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o release/generate -ldflags="-s -w" main.go
```
