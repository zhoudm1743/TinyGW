# 设备登录和心跳处理功能

## 功能概述

本文档描述了在 `pkg/service/listener/zdm.go` 中实现的设备登录、心跳处理和在线状态管理功能。

## 核心功能

### 1. 自动协议识别
- **Q3761-1376协议**：检测 `0x68` 起始符和第6字节的 `0x68`
- **619-BY协议**：检测 `0x68` 起始符和第2字节的 `0x10`

### 2. 设备登录处理
当收到设备登录请求时：
1. **自动识别登录帧**：通过解析变量 `LoginStatus` 包含"登录"关键字
2. **添加到在线设备表**：记录设备地址、协议类型、连接、登录时间等
3. **自动回复登录确认**：发送 AFN=00H F1 全部确认响应

### 3. 设备心跳处理  
当收到设备心跳时：
1. **自动识别心跳帧**：通过解析变量 `HeartbeatStatus` 包含"心跳"关键字
2. **更新心跳时间**：刷新设备的最后心跳时间
3. **自动回复心跳确认**：发送 AFN=00H F1 全部确认响应

### 4. 在线设备管理

#### 设备状态结构
```go
type OnlineDevice struct {
    DeviceAddr    string    // 设备地址 (如: 053040961)
    Protocol      string    // 协议类型 (Q3761-1376/619-BY)
    Conn          net.Conn  // 网络连接
    LastHeartbeat time.Time // 最后心跳时间
    LoginTime     time.Time // 登录时间
    RemoteAddr    string    // 远程IP地址
}
```

#### 设备管理API
- `GetOnlineDevices()` - 获取所有在线设备列表
- `GetDeviceStatus(deviceAddr)` - 获取指定设备状态
- `SendToDevice(deviceAddr, data)` - 向指定设备发送数据
- `isDeviceOnline(deviceAddr)` - 检查设备是否在线

### 5. 设备健康检查
- **定期检查**：每分钟检查一次设备心跳状态
- **离线判断**：超过5分钟没有心跳认为设备离线
- **自动清理**：清理离线设备并关闭连接

## 协议响应格式

### Q3761-1376 确认响应帧结构
```
起始符:     68
长度域:     32 00 32 00 (L=50)
起始符:     68
控制域:     00 (下行确认)
地址域:     [5字节设备地址]
AFN:        00 (确认/否认)
SEQ:        C0 (FIR=1,FIN=1,CON=0,PSEQ=0)
DA:         00 00 (p0)
DT:         01 00 (F1)
校验和:     [计算值]
结束符:     16
```

## 设备地址解析

### Q3761-1376 地址格式
- **完整地址**：9位数字 (如: 053040961)
- **行政区划码**：前4位BCD编码 (0530)
- **终端地址**：后5位二进制编码 (40961)

### 地址提取过程
1. 从数据包第7-11字节提取地址域
2. 解析行政区划码（BCD码，小端序）
3. 解析终端地址（二进制，小端序）
4. 组合成完整的9位设备地址

## 使用示例

### 获取在线设备
```go
import "tinyGW/pkg/service/listener"

// 获取所有在线设备
devices := listener.GetOnlineDevices()
for addr, device := range devices {
    fmt.Printf("设备: %s, 协议: %s, 登录时间: %v\n", 
        addr, device.Protocol, device.LoginTime)
}
```

### 检查设备状态
```go
device, online := listener.GetDeviceStatus("053040961")
if online {
    fmt.Printf("设备在线，最后心跳: %v\n", device.LastHeartbeat)
} else {
    fmt.Println("设备离线")
}
```

### 向设备发送数据
```go
data := []byte{0x68, 0x32, 0x00, ...} // 协议数据
err := listener.SendToDevice("053040961", data)
if err != nil {
    fmt.Printf("发送失败: %v\n", err)
}
```

## 注意事项

1. **心跳超时**：设备每3分钟发送心跳，平台5分钟超时判断离线
2. **自动回复**：所有登录和心跳请求都会自动回复确认，无需手动处理  
3. **线程安全**：所有设备管理操作都是线程安全的
4. **连接管理**：设备离线时会自动关闭TCP连接
5. **协议扩展**：可以轻松添加对新协议的支持

## 日志级别

- **Info**：设备登录、离线、清理等重要事件
- **Debug**：心跳更新、响应发送等详细信息
- **Warn**：数据解析失败、协议识别失败等警告
- **Error**：网络错误、脚本加载失败等错误信息 