# go-task-queue

## 项目说明
一个轻量级的任务队列管理服务，支持任务提交、任务状态查询、任务执行和任务重试。基于内存存储，提供简单的任务调度能力。

## 标准命令
go build ./...     # 编译
go test ./...      # 测试
go run ./cmd       # 启动

## Docker 命令
docker build -t go-task-queue -f benzhi.Dockerfile .                      # 构建镜像
docker run -it go-task-queue:latest                  # 运行容器
docker build --platform linux/arm64 -t go-task-queue -f benzhi.Dockerfile .  # 构建 arm64
