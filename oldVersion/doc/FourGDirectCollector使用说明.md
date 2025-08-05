# FourGDirectCollector 4G直连采集器使用说明

## 概述

`FourGDirectCollector` 是一个专门为4G设备直连设计的采集器，它能够为每个在线设备建立独立的读写通道，实现与4G设备的TCP通信。

## 主要特性

1. **自动设备管理**：自动从 `listener` 包获取在线设备列表
2. **独立读写通道**：为每个设备建立独立的TCP连接通道
3. **线程安全**：使用读写锁保证并发安全
4. **自动重连**：连接断开时自动尝试重新连接
5. **超时控制**：支持读写超时设置
6. **批量操作**：支持向所有设备或指定设备发送数据

## 架构设计

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   FourGDirect   │    │   OnlineDevice   │    │   TCP连接池     │
│   Collector     │◄──►│   管理(Listener)  │◄──►│   (设备地址映射) │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   设备连接映射   │    │   心跳检测       │    │   协议解析       │
│  (deviceConn)   │    │   (5分钟超时)    │    │   (Q3761-1376)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 核心方法

### 基础接口方法

```go
// 打开采集器
func (f *FourGDirectCollector) Open(device *models.Device) bool

// 关闭采集器
func (f *FourGDirectCollector) Close() bool

// 读取数据（从任意在线设备）
func (f *FourGDirectCollector) Read(data []byte) int

// 写入数据（向所有在线设备）
func (f *FourGDirectCollector) Write(data []byte) int

// 获取采集器名称
func (f *FourGDirectCollector) GetName() string

// 获取超时时间
func (f *FourGDirectCollector) GetTimeout() int

// 获取采集间隔
func (f *FourGDirectCollector) GetInterval() int
```

### 扩展方法

```go
// 向指定设备写入数据
func (f *FourGDirectCollector) WriteToDevice(deviceAddr string, data []byte) (int, error)

// 从指定设备读取数据
func (f *FourGDirectCollector) ReadFromDevice(deviceAddr string, data []byte) (int, error)

// 获取指定设备的连接
func (f *FourGDirectCollector) GetDeviceConnection(deviceAddr string) (net.Conn, bool)

// 获取所有设备连接
func (f *FourGDirectCollector) GetAllDeviceConnections() map[string]net.Conn

// 更新连接映射
func (f *FourGDirectCollector) UpdateConnections()

// 获取在线设备数量
func (f *FourGDirectCollector) GetOnlineDeviceCount() int

// 检查设备是否在线
func (f *FourGDirectCollector) IsDeviceOnline(deviceAddr string) bool
```

## 使用示例

### 1. 基本使用

```go
package main

import (
    "time"
    "tinyGW/app/models"
    "tinyGW/pkg/service/collect/collector"
    "go.uber.org/zap"
)

func main() {
    // 创建4G直连采集器
    fourGCollector := &collector.FourGDirectCollector{
        Collector: models.Collector{
            Name:     "4G直连采集器",
            Type:     "fouGPRS",
            Address:  "0.0.0.0:8080",
            Timeout:  30,
            Interval: 1000,
        },
    }

    // 方式1: 传入nil进行初始化（适用于Worker.Start()场景）
    if !fourGCollector.Open(nil) {
        zap.S().Error("初始化4G直连采集器失败")
        return
    }
    defer fourGCollector.Close()

    // 方式2: 传入具体设备
    device := &models.Device{
        Name:    "4G设备001",
        Address: "053040961", // 设备地址
        Type: models.DeviceType{
            Name:   "Q3761-1376设备",
            Driver: "Q3761-1376",
        },
    }

    if !fourGCollector.Open(device) {
        zap.S().Warn("设备不在线，但采集器已初始化")
    }

    zap.S().Info("4G直连采集器打开成功")

    // 定期更新连接状态
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                fourGCollector.UpdateConnections()
                onlineCount := fourGCollector.GetOnlineDeviceCount()
                zap.S().Infof("当前在线设备数量: %d", onlineCount)
            }
        }
    }()

    // 读取数据
    go func() {
        data := make([]byte, 1024)
        for {
            cnt := fourGCollector.Read(data)
            if cnt > 0 {
                zap.S().Infof("读取到数据: [% 2x]", data[:cnt])
                // 处理数据...
            }
            time.Sleep(time.Duration(fourGCollector.GetInterval()) * time.Millisecond)
        }
    }()

    // 写入数据
    go func() {
        command := []byte{0x68, 0x32, 0x00, 0x32, 0x00, 0x68, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x16}
        for {
            cnt := fourGCollector.Write(command)
            if cnt > 0 {
                zap.S().Infof("向设备发送命令，发送字节数: %d", cnt)
            }
            time.Sleep(30 * time.Second)
        }
    }()

    // 保持程序运行
    select {}
}
```

### 2. 向特定设备发送数据

```go
// 向指定设备发送数据
deviceAddr := "053040961"
command := []byte{0x68, 0x32, 0x00, 0x32, 0x00, 0x68, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x16}

if fourGCollector.IsDeviceOnline(deviceAddr) {
    cnt, err := fourGCollector.WriteToDevice(deviceAddr, command)
    if err != nil {
        zap.S().Errorf("向设备 %s 发送数据失败: %v", deviceAddr, err)
    } else {
        zap.S().Infof("向设备 %s 发送数据成功，字节数: %d", deviceAddr, cnt)
    }
} else {
    zap.S().Warnf("设备 %s 不在线", deviceAddr)
}
```

### 3. 从特定设备读取数据

```go
// 从指定设备读取数据
deviceAddr := "053040961"
data := make([]byte, 1024)

if fourGCollector.IsDeviceOnline(deviceAddr) {
    cnt, err := fourGCollector.ReadFromDevice(deviceAddr, data)
    if err != nil {
        zap.S().Debugf("从设备 %s 读取数据失败: %v", deviceAddr, err)
    } else if cnt > 0 {
        zap.S().Infof("从设备 %s 读取数据: [% 2x]", deviceAddr, data[:cnt])
    }
}
```

### 4. 批量设备管理

```go
// 批量发送命令到所有设备
command := []byte{0x68, 0x32, 0x00, 0x32, 0x00, 0x68, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x16}

totalWritten := fourGCollector.Write(command)
if totalWritten > 0 {
    zap.S().Infof("批量发送命令，总发送字节数: %d", totalWritten)
}

// 获取所有设备连接
connections := fourGCollector.GetAllDeviceConnections()
zap.S().Infof("当前管理的设备连接数量: %d", len(connections))

for deviceAddr, conn := range connections {
    if conn != nil {
        zap.S().Debugf("设备 %s 连接状态: 正常", deviceAddr)
    } else {
        zap.S().Warnf("设备 %s 连接状态: 异常", deviceAddr)
    }
}
```

### 5. 使用工厂函数创建

```go
// 通过工厂函数创建
collectorConfig := models.Collector{
    Name:     "4G直连采集器",
    Type:     "fouGPRS",
    Address:  "0.0.0.0:8080",
    Timeout:  30,
    Interval: 1000,
}

collector := collector.ConnectorFactory(collectorConfig)
if collector == nil {
    zap.S().Error("创建4G直连采集器失败")
    return
}

// 类型断言
if fourGCollector, ok := collector.(*collector.FourGDirectCollector); ok {
    zap.S().Info("成功创建4G直连采集器")
    
    device := &models.Device{
        Name:    "4G设备001",
        Address: "053040961",
    }
    
    if fourGCollector.Open(device) {
        zap.S().Info("4G直连采集器打开成功")
        defer fourGCollector.Close()
        
        // 进行数据读写操作...
    }
}
```

## 配置参数

### Collector配置

```go
type Collector struct {
    Name      string    // 采集器名称
    Type      string    // 类型: "fouGPRS" 或 "FourGPRS"
    Address   string    // 地址 (可选，主要用于标识)
    Timeout   int       // 超时时间(秒)，默认30秒
    Interval  int       // 采集间隔(毫秒)，默认1000毫秒
}
```

### 设备配置

```go
type Device struct {
    Name    string      // 设备名称
    Address string      // 设备地址 (如: "053040961")
    Type    DeviceType  // 设备类型
}
```

## 工作流程

1. **初始化**：创建FourGDirectCollector实例
2. **打开**：调用Open()方法，检查设备是否在线
3. **连接管理**：自动从listener包获取在线设备连接
4. **数据读写**：
   - Read(): 从任意在线设备读取数据
   - Write(): 向所有在线设备发送数据
   - WriteToDevice(): 向指定设备发送数据
   - ReadFromDevice(): 从指定设备读取数据
5. **状态监控**：定期调用UpdateConnections()更新连接状态
6. **关闭**：调用Close()方法清理所有连接

## 注意事项

1. **设备在线检查**：使用前应检查设备是否在线
2. **连接更新**：定期调用UpdateConnections()同步连接状态
3. **错误处理**：读写操作可能返回错误，需要适当处理
4. **超时设置**：根据网络环境调整超时时间
5. **并发安全**：所有操作都是线程安全的
6. **资源清理**：使用完毕后调用Close()方法
7. **空指针处理**：Open()方法支持传入nil设备，此时仅进行初始化
8. **Worker集成**：与Worker.Start()方法兼容，不会出现空指针异常

## 性能优化

1. **连接池管理**：自动管理设备连接，避免重复创建
2. **读写超时**：设置合理的超时时间，避免阻塞
3. **批量操作**：使用Write()方法向所有设备发送数据
4. **状态缓存**：缓存设备在线状态，减少查询开销

## 故障排除

### 常见问题

1. **设备连接失败**
   - 检查设备是否在线
   - 确认设备地址是否正确
   - 检查网络连接

2. **读写超时**
   - 增加超时时间
   - 检查网络延迟
   - 确认设备响应正常

3. **连接断开**
   - 自动重连机制会处理
   - 检查设备心跳状态
   - 确认网络稳定性

4. **空指针异常**
   - 已修复：Open()方法现在支持传入nil设备
   - 与Worker.Start()方法完全兼容
   - 传入nil时仅进行初始化，不会报错

### 日志级别

- **Info**：重要操作信息
- **Debug**：详细调试信息
- **Warn**：警告信息
- **Error**：错误信息

## 扩展功能

1. **协议支持**：支持Q3761-1376、619-BY等协议
2. **设备管理**：自动管理设备上线/下线
3. **数据缓存**：支持数据缓存和重发
4. **监控告警**：设备状态监控和告警
5. **性能统计**：连接数、读写次数等统计信息 