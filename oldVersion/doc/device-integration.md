# 设备接入指南

## 1. 设备型号注册
```lua
-- 示例：水表型号注册脚本
local device_type = {
    name = "DDSY1829",
    properties = {
        {name="累计流量", address=0x0000, type="float32", unit="m³"},
        {name="瞬时流量", address=0x0004, type="float32", unit="m³/h"}
    },
    driver = "DDSY1829.lua"
}
GW.set_instrument_types({device_type})
```

## 2. 协议适配开发

### 2.1 Lua驱动开发规范

#### 文件结构要求
```lua
-- 必备基础库引入
package.path = "./plugin/[型号]/?.lua;"
require "utils"
require "json"

-- 设备指令生成函数
function GenerateCommand(sAddr, cmd)
    -- 指令帧构造逻辑
end

-- 数据解析入口
function AnalysisRx(sAddr, rxBufCnt)
    -- 原始数据解析处理
end

-- 设备自定义命令入口
function DeviceCustomCmd(sAddr, cmdName, cmdParam, step)
    -- 命令路由逻辑
end
```

#### 核心函数规范
1. `GenerateCommand`: 构造设备通信指令帧
2. `AnalysisRx`: 实现原始数据到结构化数据的转换
3. `DeviceCustomCmd`: 响应平台下发的控制指令

#### 数据转换机制
```lua
-- 典型值转换示例（BCD码转十进制）
function HexToDec(hexStr)
    return tonumber(hexStr:gsub("%x", ""), 16)
end

-- 浮点数解析示例（大端序）
local value = struct.unpack(">f", data:sub(offset, offset+3))
```

### 2.2 与核心模块交互

#### 指令执行流程
```plantuml
Go核心 -> Lua脚本: 调用GenerateCommand
Lua脚本 -> 设备: 发送指令帧
设备 -> Go核心: 返回原始数据
Go核心 -> Lua脚本: 调用AnalysisRx
Lua脚本 -> 平台: 返回结构化数据
```

#### 调试验证方法
1. 单元测试：`lua test/[型号]_test.lua`
2. 指令模拟：`POST /api/debug/send {device:"WT-001", cmd:"01..."}`
3. 实时监控：`tail -f logs/lua_runtime.log`
```lua
function parse(data)
    local result = {}
    result.累计流量 = struct.unpack(">f", data:sub(1,4))
    result.瞬时流量 = struct.unpack(">f", data:sub(5,8))
    return result
end
```

## 3. 设备注册(MQTT指令示例)
```json
{
    "method": "SetDevices",
    "params": {
        "devices": [{
            "name": "WT-001",
            "type": "DDSY1829",
            "collector": {"name":"COM1", "protocol":"modbus_rtu"},
            "settings": {
                "slave_id": 1,
                "register_map": {"start":0, "quantity":8}
            }
        }]
    }
}
```

## 4. 数据采集配置
```yaml
collect_task:
  name: 数据采集
  interval: 60s
  devices:
    - WT-001
  metrics:
    - 累计流量
    - 瞬时流量
```

## 5. 接入流程时序
```
设备注册 -> 驱动部署 -> 采集配置 -> 任务启动 -> 数据上报
           ↑           |
           └─心跳监测←─┘
```

## 6. 故障排查
1. 查看设备通信日志：`tail -f logs/collect.log`
2. 使用调试接口：`POST /api/debug/send {device:"WT-001", cmd:"010300000008"}`
3. 验证驱动解析：`lua test/DDSY1829_test.lua`