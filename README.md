

# TinyGW 物联网网关系统

TinyGW 是一个专注于物联网设备接入与管理的轻量级网关系统。通过灵活的插件机制与多协议支持，TinyGW 能够适应多种物联网场景，提供设备通信、数据采集、远程控制与系统管理等功能。

## 程序架构图
系统采用模块化设计，分为以下几个核心层次：
- **协议适配层**：负责支持多种物联网通信协议（如 Q1376.1、619-BY、Lua 驱动等）。
- **任务调度引擎**：管理采集任务与上报任务的调度与执行。
- **设备管理服务**：提供设备生命周期管理、状态监控、在线管理等功能。
- **系统服务层**：包含 HTTP 服务、MQTT 通信、日志、NTP、数据库等基础服务。
- **插件系统**：支持 Lua 脚本扩展，实现协议解析、数据处理等功能。

## 程序流程图（简略）
1. 设备接入：通过 TCP、MQTT、串口等方式连接设备。
2. 自动协议识别：根据设备特性自动匹配协议插件。
3. 数据采集与转发：采集设备数据，支持本地存储与云端转发。
4. 远程控制：接收云端指令，下发到设备。
5. 任务调度：定时采集、定时上报任务的触发与执行。
6. 状态管理：心跳处理、设备在线状态维护、健康检查。

## 组件关系及数据流向
- **采集器（Collector）**：管理设备连接，负责数据收发。
- **采集任务（CollectTask）**：定时触发采集器任务。
- **上报任务（ReportTask）**：定时将设备数据上报至云端。
- **事件总线（EventBus）**：发布与订阅设备状态、采集数据等事件。
- **插件系统**：通过 Lua �ance 数据解析、命令生成等。

## 功能特性

### 设备接入
- 支持 TCP、MQTT、串口、4G �--等多种通信方式。
- 设备自动登录、心跳检测与地址识别。
- 支持多种通信协议（如 Q1376.1、619-BY、DLT645、自定义 Lua 脚本等）。

### 数据处理
- 支持 Lua 脚本进行数据解析与转换。
- 提供数据缓存、队列管理与并发控制。
- 支持数据持久化与内存中缓存。

### 远程管理
- 通过 MQTT 接收远程指令（如升级、重启、采集任务控制等）。
- 支持远程配置设备参数、采集任务与上报任务。
- 提供 NTP 时间同步与日志上传功能。

## 系统功能
- **HTTP API 接口**：提供 RESTful 接口管理设备、采集任务、上报任务、用户等。
- **MQTT 通信**：实现与云端平台的双向通信，支持指令下发与数据上报。
- **协议适配**：支持多种协议插件，可通过 Lua 脚本扩展。
- **任务调度**：采集与上报任务的定时触发与动态管理。
- **设备状态管理**：实时维护设备在线状态，支持心跳与健康检查。

## 技术架构
- **后端框架**：Golang + Gin + GORM。
- **插件系统**：Lua 脚本驱动协议解析与数据处理。
- **消息通信**：基于 Eclipse Paho 的 MQTT 客户端。
- **任务调度**：使用 fx、cron �--等实现定时任务。
- **日志系统**：Zap 日志框架。
- **协议解析**：基于 AST 结构与 Lua 脚本的组合解析引擎。

## 快速开始

### 环境要求
- Golang 1.20+
- Lua 5.4+
- MySQL 8.0 或 SQLite
- MQTT Broker
- Linux / Windows / ARM 嵌入式环境

### 安装与配置
1. 下载源码：
```bash
git clone https://gitee.com/illusoryNone/tiny-gw.git
```

2. 安装依赖：
```bash
go mod download
```

3. 修改配置：
- `config/conf.yml`：配置数据库、MQTT、HTTP 服务、日志等。
- `config/ntp.conf`：时间同步配置。

### 运行与调试
- 启动服务：
```bash
go run main.go
```

- 查看日志：
```bash
tail -f public/logs/energy.log
```

- 使用 HTTP API 调试：
  - 登录：`POST /api/account/login`
  - 查看设备：`GET /api/device`
  - 添加采集任务：`POST /api/collect-task`

## 开发指南

### 目录结构
- `app/api/`：HTTP 接口定义。
- `pkg/service/`：核心服务组件（MQTT、HTTP、定时任务、采集器等）。
- `pkg/plugin/`：插件系统，包含 Lua 解析与通用工具。
- `plugin/`：设备协议插件，支持 Lua 驱动。
- `config/`：配置文件。
- `public/`：静态资源（如 Web 页面）。

### 添加新设备类型
1. 在 `plugin/` 目录下创建设备协议 Lua 文件。
2. 实现 `GenerateCommand`, `AnalysisRx`, `DeviceCustomCmd` 等核心函数。
3. 在 Web API 中使用 `POST /api/device-type` 注册设备类型。
4. 配置采集任务与上报任务。

### 扩展功能模块
- **采集器插件**：实现 `Collector` 接口，支持新连接方式。
- **协议解析插件**：在 `pkg/plugin/` 中实现 Lua 解析函数。
- **定时任务**：扩展 `CollectTaskServer` 或 `ReportTaskServer`。
- **事件处理**：在 `pkg/service/event/` 添加自定义事件处理器。

## 部署流程

### 编译打包
- 编译 Linux 版本：
```bash
go build -o tinygw main.go
```

- 编译 ARM 版本（用于嵌入式设备）：
```bash
GOARCH=arm GOARM=7 go build -o tinygw_arm main.go
```

### 容器部署
- 构建 Docker 镜像：
```bash
docker build -t tinygw .
```

- 运行容器：
```bash
docker run -d -p 8080:8080 -v ./config:/app/config tinygw
```

### 系统服务部署
- 将 `install.sh` 脚本复制到目标系统：
```bash
cp install.sh /etc/init.d/tinygw
```

- 设置开机启动：
```bash
systemctl enable tinygw
systemctl start tinygw
```

## 常见问题

### 设备连接问题
- 检查设备地址与协议是否匹配。
- 检查采集器配置（串口、IP、MQTT 主题等）。
- 查看日志：`public/logs/energy.log`

### 数据采集问题
- 检查采集任务是否启用。
- 检查 Lua 解析脚本是否正确。
- 检查设备心跳是否正常。

### 系统性能优化
- 调整 `config/conf.yml` 中的采集频率与并发配置。
- 使用缓存机制：`pkg/service/cache/`
- 优化 Lua 脚本性能，减少不必要的数据处理。

## 联系方式
如有问题或需要技术支持，请联系：
- 邮箱：illusoryNone@example.com
- 电话：+86 123 4567 8901
- GitHub Issues：[提交问题](https://gitee.com/illusoryNone/tiny-gw/issues)

## 许可证
本项目采用 MIT 许可证，请查看 `LICENSE` 文件以获取详细信息。