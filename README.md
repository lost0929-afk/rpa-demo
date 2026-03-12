# RPA Demo 说明
这是一个最小可用的RPA Demo，已经实现了核心的流程执行和一键打包EXE功能。
## 功能特性
✅ 支持JSON格式定义自动化流程
✅ 内置常用操作节点：日志、等待、执行命令、打开网址、弹窗提示
✅ ✨ 核心功能：一键打包流程为独立EXE文件，无任何依赖
✅ 打包后的EXE体积 < 10MB，双击即可运行
## 使用方法
### 1. 编译Demo主程序
```bash
go build -ldflags "-s -w" -o rpa-demo.exe main.go
```
### 2. 执行流程
```bash
rpa-demo.exe -flow demo-flow.json
```
### 3. 打包流程为独立EXE
```bash
rpa-demo.exe -build demo-flow.json -output my-automation.exe
```
## 支持的节点类型
### log - 打印日志
```json
{
  "type": "log",
  "params": {
    "message": "要打印的日志内容"
  }
}
```
### delay - 等待
```json
{
  "type": "delay",
  "params": {
    "seconds": 3
  }
}
```
### run_command - 执行系统命令
```json
{
  "type": "run_command",
  "params": {
    "command": "dir /b"
  }
}
```
### open_url - 打开浏览器访问网址
```json
{
  "type": "open_url",
  "params": {
    "url": "https://www.baidu.com"
  }
}
```
### message_box - 弹出提示框
```json
{
  "type": "message_box",
  "params": {
    "text": "提示内容"
  }
}
```
## 扩展开发
你可以很容易地添加更多节点类型，只需要在`executeNode`函数中添加新的case即可。
