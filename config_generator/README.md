# 使用说明

## 操作步骤

1.在 config.yaml 中设置好可用的 tianwan1、tianwan2 推理服务器资源
2.在 config.yaml 中设置告警服务器 url
3.在 config.yaml 中设置摄像头信息的的 excel 文件路径 (需要按照固定格式的 excel 文件)
4.(可选) 定义需要过滤掉的摄像头信息, 生成的配置文件将不包括这些摄像头
5.运行如下指令生成配置文件

  ```bash
    # Windows 环境下
    ./generate.exe -c config.yaml
  ```

6.将生成的配置文件 `tianwan_config.json` 导入摄像头平台
7.(可选) 在摄像头平台上按照需要设置阈值以提升识别输出效果

## 其他

1.dist 包里已经有编译好的不同平台的配置生成程序以及输出的摄像头平台配置文件
2.如果有需要可以自行编译这个配置生成程序

```bash
# 如果在 Linux/macOS 上请使用 GOOS=xxx GOARCH=xxx 指示环境变量

# Linux 64位
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o dist/generate_x86_64_linux -ldflags="-s -w" main.go

# Windows 64位
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o dist/generate_x86_64_win.exe -ldflags="-s -w" main.go

# macOS 64位 (Intel)
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o dist/generate_x86_64_darwin -ldflags="-s -w" main.go

# macOS (Apple Silicon)
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o dist/generate_arm64_darwin -ldflags="-s -w" main.go
```
