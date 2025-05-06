# TinyGW 物联网网关

[![Go Version](https://img.shields.io/badge/go-1.19+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

物联网网关系统，支持通过485接口和TCP网桥进行设备数据采集，提供数据上报至物联网平台（XiaoCloud）和命令下发功能。

## 功能特性
- 多协议设备接入（Modbus RTU/TCP/MQTT/Serial）
- 实时数据采集与边缘计算
- 双向通信支持命令下发
- 设备管理（型号管理、设备注册）
- 任务调度（定时采集、数据上报）
- 系统监控与远程升级

## 技术栈
### 核心组件
- **开发语言**: Go 1.19+
- **依赖注入**: Uber FX
- **Web框架**: Gin v1.9.1
- **数据库**: BoltDB 1.3.10
- **任务调度**: Cron v3.0.1
- **MQTT客户端**: Eclipse Paho v1.4.0

### 开发工具
- 脚本引擎：Go Lua 1.0.0
- 日志系统：Zap 1.24.0

## 快速开始
### 环境准备
1. 安装 Go 1.19+ 和 Node.js 16+ 和 lua5.4 环境
2. 安装 UPX 压缩工具（可选）
3. 配置 XiaoCloud 1.6.1+ 实例

### 安装依赖
```bash
# 后端依赖
cd energy-saas-gateway
go mod download

# 前端依赖
cd webapp
npm install --registry=https://registry.npmmirror.com
```

### 配置说明
复制示例配置文件并修改参数：
```bash
cp config/conf.example.yml config/conf.yml
```
主要配置项：
- `mqtt.broker`: XiaoCloud MQTT 地址
- `gateway.id`: 网关设备标识
- `collect.interval`: 默认采集间隔

## 开发指南
### 编译运行
```bash
# 启动后端服务
go run main.go

# 启动前端开发服务器
cd webapp
npm run dev
```

### 测试验证
```bash
# 运行单元测试
go test ./...

# 接口测试
curl http://localhost:8080/api/healthcheck
```

## 部署流程
### 生产环境构建
```bash
# 构建Linux可执行文件
./build-linux.sh

# 构建ARM系统可执行文件
./build-arm.bat
```

### 容器化部署
```Dockerfile
FROM golang:1.19-alpine AS builder
RUN apk add --no-cache upx
WORKDIR /app
COPY . .
RUN go build -ldflags "-s -w" -o tinyGW && \
    upx --best tinyGW

FROM alpine:3.17
COPY --from=builder /app/tinyGW /app/config.yml /app/webroot/
EXPOSE 8080
CMD ["/app/tinyGW"]
```

## 文档目录
- [系统架构文档](./doc/architecture.md)
- [API接口规范](./doc/ts004-WebApi.md)
- [设备接入指南](./doc/device-integration.md)

## 项目结构
```
energy-saas-gateway/
├── api/               # API接口层
│   ├── controller/    # HTTP控制器
│   ├── domain/        # 领域模型
│   ├── repository/    # 数据持久层
│   ├── routes/        # 路由配置
│   └── service/       # 业务服务
├── core/              # 核心模块
│   ├── config/        # 配置管理
│   ├── database/      # 数据库连接
│   ├── mqtt/          # MQTT客户端
│   ├── task/          # 定时任务
│   └── web/           # Web服务
├── doc/               # 项目文档
├── plugin/            # 设备协议插件
│   └── DDSU5886/      # 电表协议实现
├── bootstrap/         # 应用初始化
├── release/           # 构建输出
└── util/              # 通用工具

主要配置文件：
- config/conf.yml    # 应用配置
- tinyGW.service       # 系统服务配置
```

## 许可证
[MIT License](LICENSE)

