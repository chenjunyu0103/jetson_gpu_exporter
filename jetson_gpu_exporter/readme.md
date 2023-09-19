# 软件说明
## 环境依赖
- go1.17 版本以上
- git 客户端
- GoLand 2021.3.3
- make 命令安装
## go 用到的框架
 - viper 主要提供命令模式
 - cobra 主要提供参数解析和绑定
 - logrus 日志框架
 - prometheus/client_golang prometheus客户端
 - robfig/cron 定时任务框架
## 关键类说明
 - 启动类 main.go
 - Tegrastats.go 调用Tegrastats 命令先关类和 Tegrastats 生成的文件解析
 - exporter.go 提供http 服务,并调用prometheus客户端 实现 指标上报
 - cobra.go 参数解析并调用exporter 启动http 服务
## 程序编译
 make build
