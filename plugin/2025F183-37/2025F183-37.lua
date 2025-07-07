---
--- 2025F183-37.lua
--- 基于619规约的CAT1水表通信协议实现
--- 创建时间: 2025/07/07
--- 修订时间: 2025/07/07 (v1.0.2)
--- 修订说明: 
--- 1. 修正地址格式，确保符合619协议"低字节在前"的规范
--- 2. 修正数据长度字段为2字节（高字节在前）
--- 3. 修正数据解析逻辑，正确处理返回的流量数据
--- 4. 添加对表端主动上报数据(0x97)的支持
--- 5. 适配网口采集器TcpServerCollector
---

-- 获取当前脚本所在的绝对路径
local current_dir = debug.getinfo(1).source:match("@?(.*/)") or ""
if current_dir == "" then current_dir = "./" end

-- 使用绝对路径设置package.path
package.path = current_dir .. "?.lua;" .. package.path

require "utils"
require "json"

-------公共函数
-- 校验和
function checkSum(buffer, startIndex, endIndex)
    local sum = 0

    for i = startIndex, endIndex do
        sum = sum + buffer[i]
    end

    return sum % 256
end

-- 地址转换 2 ~ 8 字节
function convertAddress(requestADU, sAddr)
    -- 根据619协议，表地址由A0~A6共7个字节表示，低字节在前，14位BCD码形式
    local addr = string.format("%014d", tonumber(sAddr))
    requestADU[2] = 0x10  -- 表类型固定为水表0x10
    
    -- 低字节在前的地址格式
    requestADU[9] = tonumber(string.sub(addr, 1, 2), 16)  -- A0（最低位）
    requestADU[8] = tonumber(string.sub(addr, 3, 4), 16)  -- A1
    requestADU[7] = tonumber(string.sub(addr, 5, 6), 16)  -- A2
    requestADU[6] = tonumber(string.sub(addr, 7, 8), 16)  -- A3
    requestADU[5] = tonumber(string.sub(addr, 9, 10), 16) -- A4
    requestADU[4] = tonumber(string.sub(addr, 11, 12), 16)-- A5
    requestADU[3] = tonumber(string.sub(addr, 13, 14), 16)-- A6（最高位）
end

-- GenerateCommand - 生成读取命令
function GenerateCommand(sAddr, cmd)
    -- 按照619协议构建帧结构
    -- 68 | 表类型 | 表地址(7字节) | 设备类型 | 控制码 | 指令编号(2字节) | 备用(2字节) | 数据长度(2字节) | 数据域 | 校验码 | 16
    local requestADU = { 
        0x68,  -- 起始帧
        0x10,  -- 表类型(水表)
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  -- 表地址(7字节)
        0x03,  -- 设备类型
        0x01,  -- 控制码(01H: 服务器对表端下发读指令)
        0x00, 0x00,  -- 指令编号
        0x00, 0x00,  -- 备用
        0x00, 0x0A,  -- 数据长度(2字节，高字节在前)
        0x00, 0x00,  -- 数据标识
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  -- 备用(8字节)
        0x00,  -- 校验码
        0x16   -- 结束帧
    }
    
    -- 地址转换
    convertAddress(requestADU, sAddr)
    
    -- 指令编号 - 随机生成
    requestADU[12] = math.random(0, 255)
    requestADU[13] = math.random(0, 255)
    
    -- 数据标识
    requestADU[18] = cmd[1]
    requestADU[19] = cmd[2]
    
    -- 校验和(从起始帧到校验码前所有数据)
    requestADU[28] = checkSum(requestADU, 1, 27)
    
    -- -- 在requestADU前面插入 3个 0xFE 作为唤醒符
    -- for i = 1, 3 do
    --     table.insert(requestADU, 1, 0xFE)
    -- end
    
    return requestADU
end

-- 阀门控制命令生成（开关阀）
function WriteCommand(sAddr, cmd, data)
    -- 按照619协议构建帧结构
    -- 68 | 表类型 | 表地址(7字节) | 设备类型 | 控制码 | 指令编号(2字节) | 备用(2字节) | 数据长度(2字节) | 数据域 | 校验码 | 16
    local requestADU = { 
        0x68,  -- 起始帧
        0x10,  -- 表类型(水表)
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  -- 表地址(7字节)
        0x03,  -- 设备类型
        0x04,  -- 控制码(04H: 服务器对表端下发写指令)
        0x00, 0x00,  -- 指令编号
        0x00, 0x00,  -- 备用
        0x00, 0x0C,  -- 数据长度(2字节，高字节在前)
        0x00, 0x00,  -- 数据标识
        0x00, 0x00,  -- 阀门开/关和操作类型
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  -- 备用(8字节)
        0x00,  -- 校验码
        0x16   -- 结束帧
    }
    
    -- 地址转换
    convertAddress(requestADU, sAddr)
    
    -- 指令编号 - 随机生成
    requestADU[12] = math.random(0, 255)
    requestADU[13] = math.random(0, 255)
    
    -- 数据标识
    requestADU[18] = cmd[1]
    requestADU[19] = cmd[2]
    
    -- 阀门操作数据
    requestADU[20] = data[1]  -- 阀门开/关: 0x55开，0x99关
    requestADU[21] = data[2]  -- 操作类型: 0x5A强制，其他非强制
    
    -- 校验和(从起始帧到校验码前所有数据)
    requestADU[30] = checkSum(requestADU, 1, 29)
    
    -- -- 在requestADU前面插入 3个 0xFE 作为唤醒符
    -- for i = 1, 3 do
    --     table.insert(requestADU, 1, 0xFE)
    -- end
    
    return requestADU
end

-- 关阀
function GenerateCloseValve(sAddr)
    local cmd = { 0xAA, 0x05 }  -- 阀门操作数据标识
    local data = { 0x99, 0x5A }  -- 0x99表示关阀，0x5A表示强制操作
    return WriteCommand(sAddr, cmd, data)
end

-- 开阀
function GenerateOpenValve(sAddr)
    local cmd = { 0xAA, 0x05 }  -- 阀门操作数据标识
    local data = { 0x55, 0x5A }  -- 0x55表示开阀，0x5A表示强制操作
    return WriteCommand(sAddr, cmd, data)
end

-- 设置表底数
function GenerateSetValue(sAddr, value)
    -- 按照619协议构建帧结构
    -- 68 | 表类型 | 表地址(7字节) | 设备类型 | 控制码 | 指令编号(2字节) | 备用(2字节) | 数据长度(2字节) | 数据域 | 校验码 | 16
    local requestADU = { 
        0x68,  -- 起始帧
        0x10,  -- 表类型(水表)
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  -- 表地址(7字节)
        0x03,  -- 设备类型
        0x04,  -- 控制码(04H: 服务器对表端下发写指令)
        0x00, 0x00,  -- 指令编号
        0x00, 0x00,  -- 备用
        0x00, 0x0E,  -- 数据长度(2字节，高字节在前)
        0xAA, 0x09,  -- 数据标识(AA09H: 设置后付费表底数)
        0x00, 0x00, 0x00, 0x00,  -- 表底数(4字节)
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  -- 备用(8字节)
        0x00,  -- 校验码
        0x16   -- 结束帧
    }
    
    -- 地址转换
    convertAddress(requestADU, sAddr)
    
    -- 指令编号 - 随机生成
    requestADU[12] = math.random(0, 255)
    requestADU[13] = math.random(0, 255)
    
    -- 设置数值 (低字节在前)
    local v = tonumber(value) or 0
    requestADU[20] = v % 256
    requestADU[21] = math.floor(v / 256) % 256
    requestADU[22] = math.floor(v / 65536) % 256
    requestADU[23] = math.floor(v / 16777216) % 256
    
    -- 校验和(从起始帧到校验码前所有数据)
    requestADU[32] = checkSum(requestADU, 1, 31)
    
    -- -- 在requestADU前面插入 3个 0xFE 作为唤醒符
    -- for i = 1, 3 do
    --     table.insert(requestADU, 1, 0xFE)
    -- end
    
    return requestADU
end

-- 读取表地址
function GenerateReadAddress()
    -- 按照619协议构建帧结构
    -- 68 | 表类型 | 表地址(7字节) | 设备类型 | 控制码 | 指令编号(2字节) | 备用(2字节) | 数据长度(2字节) | 数据域 | 校验码 | 16
    local requestADU = { 
        0x68,  -- 起始帧
        0x10,  -- 表类型(水表)
        0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,  -- 表地址(7字节，广播地址)
        0x03,  -- 设备类型
        0x01,  -- 控制码(01H: 服务器对表端下发读指令)
        0x00, 0x00,  -- 指令编号
        0x00, 0x00,  -- 备用
        0x00, 0x0A,  -- 数据长度(2字节，高字节在前)
        0xA9, 0x01,  -- 数据标识(A901H: 读表参数)
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  -- 备用(8字节)
        0x00,  -- 校验码
        0x16   -- 结束帧
    }
    
    -- 指令编号 - 随机生成
    requestADU[12] = math.random(0, 255)
    requestADU[13] = math.random(0, 255)
    
    -- 校验和(从起始帧到校验码前所有数据)
    requestADU[28] = checkSum(requestADU, 1, 27)
    
    -- -- 在requestADU前面插入 3个 0xFE 作为唤醒符
    -- for i = 1, 3 do
    --     table.insert(requestADU, 1, 0xFE)
    -- end
    
    return requestADU
end

-- 设置表地址
function GenerateSetAddress(oldAddr, newAddr)
    -- 按照619协议构建帧结构
    -- 68 | 表类型 | 表地址(7字节) | 设备类型 | 控制码 | 指令编号(2字节) | 备用(2字节) | 数据长度(2字节) | 数据域 | 校验码 | 16
    local requestADU = { 
        0x68,  -- 起始帧
        0x10,  -- 表类型(水表)
        0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA,  -- 表地址(7字节)
        0x03,  -- 设备类型
        0x04,  -- 控制码(04H: 服务器对表端下发写指令)
        0x00, 0x00,  -- 指令编号
        0x00, 0x00,  -- 备用
        0x00, 0x0D,  -- 数据长度(2字节，高字节在前)
        0xAA, 0x07,  -- 数据标识(AA07H: 设置表具类型)
        0x00,  -- 计量模式
        0x00,  -- 付费模式
        0x00,  -- 到位模式
        0xC1,  -- 地址修改使能(0xC1: 允许修改表号)
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  -- 新地址(7字节)
        0x00,  -- 校验码
        0x16   -- 结束帧
    }
    
    -- 如果提供了旧地址，则转换
    if oldAddr and oldAddr ~= "" and oldAddr ~= "AAAAAAAA" then
        convertAddress(requestADU, oldAddr)
    end
    
    -- 指令编号 - 随机生成
    requestADU[12] = math.random(0, 255)
    requestADU[13] = math.random(0, 255)
    
    -- 新地址转换
    local newAddrBytes = {}
    local addr = string.format("%014d", tonumber(newAddr))
    newAddrBytes[1] = tonumber(string.sub(addr, 13, 14), 16) -- A6（最高位）
    newAddrBytes[2] = tonumber(string.sub(addr, 11, 12), 16) -- A5
    newAddrBytes[3] = tonumber(string.sub(addr, 9, 10), 16)  -- A4
    newAddrBytes[4] = tonumber(string.sub(addr, 7, 8), 16)   -- A3
    newAddrBytes[5] = tonumber(string.sub(addr, 5, 6), 16)   -- A2
    newAddrBytes[6] = tonumber(string.sub(addr, 3, 4), 16)   -- A1
    newAddrBytes[7] = tonumber(string.sub(addr, 1, 2), 16)   -- A0（最低位）
    
    -- 插入新地址
    for i = 1, 7 do
        requestADU[23 + i] = newAddrBytes[i]
    end
    
    -- 校验和(从起始帧到校验码前所有数据)
    requestADU[31] = checkSum(requestADU, 1, 30)
    
    -- -- 在requestADU前面插入 3个 0xFE 作为唤醒符
    -- for i = 1, 3 do
    --     table.insert(requestADU, 1, 0xFE)
    -- end
    
    return requestADU
end

-- go 程序接收到表具数据，放到rxBuf中，从而供AnalysisRx解析
rxBuf = {}
function AnalysisRx(sAddr, rxBufCnt)
    -- 最短指令16个字节
    if (rxBufCnt < 16) then
        rxBuf = {}
        -- Status = "1" 错误
        print("error:rxBufCnt < 16")
        return { Status = "1", Variable = {} }
    end
    
    -- 查找起始符0x68
    local startIndex = 1
    while startIndex <= #rxBuf and rxBuf[startIndex] ~= 0x68 do
        startIndex = startIndex + 1
        if startIndex > #rxBuf then
            print("error: 0x68起始符未找到")
            rxBuf = {}
            return { Status = "1", Variable = {} }
        end
    end
    
    -- 移除0xFE唤醒符
    while startIndex > 1 and rxBuf[startIndex - 1] == 0xFE do
        table.remove(rxBuf, startIndex - 1)
        startIndex = startIndex - 1
    end

    local endFlag
    local endIndex = 0
    local csFlag
    for i = 1, #rxBuf do
        if (rxBuf[i] == 0x16) then
            local cs = checkSum(rxBuf, 1, i - 2)
            if (cs == rxBuf[i-1]) then
                endFlag = rxBuf[i]
                csFlag = rxBuf[i-1]
                endIndex = i
                break
            end
        end
    end
    
    if endIndex == 0 then
        print("error: 未找到有效的结束符或校验和错误")
        rxBuf = {}
        return { Status = "1", Variable = {} }
    end
    
    -- 提取表地址 (位置2-8)
    local addr = ""
    for i = 2, 8 do
        if i <= #rxBuf then
            addr = addr .. string.format("%02X", rxBuf[i])
        end
    end
    print("接收数据表地址:", addr)
    
    -- 控制码
    local ctrlCode = rxBuf[10] or 0
    print("控制码:", string.format("0x%02X", ctrlCode))
    
    -- 处理表端主动上报 (97h)
    if ctrlCode == 0x97 then
        -- 获取上报类型
        local reportType = 0
        if rxBuf[13] and rxBuf[14] then
            reportType = (rxBuf[13] * 256) + rxBuf[14]
        end
        print("上报类型:", string.format("0x%04X", reportType))
        
        -- 获取数据包类型
        local dataPackType = rxBuf[16] or 0
        print("数据包类型:", dataPackType)
        
        -- 解析不同类型的上报数据
        if dataPackType == 0x01 or dataPackType == 0x02 or dataPackType == 0x03 then
            -- 普通表、阶梯表或后付费表上报
            local variables = {}
            
            -- 解析状态字
            local status = 0
            if rxBuf[58] and rxBuf[59] then
                status = (rxBuf[58] * 256) + rxBuf[59]
            end
            table.insert(variables, utils.AppendVariable(0, "dev_status", "状态", "int", status, tostring(status)))
            
            -- 解析累计用量/金额
            local totalValue = 0
            if dataPackType == 0x01 or dataPackType == 0x02 then
                -- 预付费表在位置61-64
                if rxBuf[61] and rxBuf[62] and rxBuf[63] and rxBuf[64] then
                    totalValue = rxBuf[61] + 
                                (rxBuf[62] * 256) + 
                                (rxBuf[63] * 65536) + 
                                (rxBuf[64] * 16777216)
                end
            elseif dataPackType == 0x03 then
                -- 后付费表在位置60-63
                if rxBuf[60] and rxBuf[61] and rxBuf[62] and rxBuf[63] then
                    totalValue = rxBuf[60] + 
                                (rxBuf[61] * 256) + 
                                (rxBuf[62] * 65536) + 
                                (rxBuf[63] * 16777216)
                end
            end
            table.insert(variables, utils.AppendVariable(1, "dev_flow", "总流量", "double", totalValue/10.0, tostring(totalValue/10.0)))
            
            -- 获取电池电压
            local voltage = 0
            if rxBuf[43] and rxBuf[44] then
                voltage = (rxBuf[43] * 256) + rxBuf[44]
                voltage = voltage / 100.0  -- 单位V
            end
            table.insert(variables, utils.AppendVariable(2, "dev_voltage", "电池电压", "double", voltage, tostring(voltage)))
            
            -- 获取信号强度
            local signal = 0
            if rxBuf[45] then
                signal = rxBuf[45]
            end
            table.insert(variables, utils.AppendVariable(3, "dev_signal", "信号强度", "int", signal, tostring(signal)))
            
            rxBuf = {}
            return {Status = "0", Variable = variables}
        end
        
        -- 其他上报类型
        print("未处理的上报类型:", reportType, "数据包类型:", dataPackType)
        rxBuf = {}
        return {Status = "1", Variable = {}}
    end
    
    -- 读表计数据响应 (81h)
    if ctrlCode == 0x81 then
        -- 根据619协议，数据标识在16-17位置
        local dataId = string.format("%02X%02X", rxBuf[16], rxBuf[17])
        print("数据标识:", dataId)
        
        -- 总流量读取 (901F)
        if dataId == "901F" then
            print("接收到总流量数据")
            
            -- 根据619协议，流量数据的位置是固定的
            -- 数据域应该从第18个字节开始
            local data = {}
            
            -- 提取4字节流量数据
            for i = 18, 21 do
                if i <= #rxBuf then
                    table.insert(data, rxBuf[i])
                end
            end
            
            if #data == 4 then
                -- 按照619协议，水表数据是低字节在前
                local flowValue = data[1] + 
                                 (data[2] * 256) + 
                                 (data[3] * 65536) + 
                                 (data[4] * 16777216)
                
                print("解析的流量值:", flowValue)
                
                rxBuf = {}
                return {Status = "0", Variable = {
                    utils.AppendVariable(0, "dev_flow", "总流量", "double", flowValue/10.0,
                            tostring(flowValue/10.0))
                }}
            else
                print("警告: 无法解析流量数据，字节数不足")
                rxBuf = {}
                return {Status = "1", Variable = {}}
            end
            
        -- 读表参数 (A901)
        elseif dataId == "A901" then
            print("接收到表参数数据")
            
            -- 根据619协议，表地址在2-8位置
            local addr = ""
            for i = 2, 8 do
                addr = addr .. string.format("%02X", rxBuf[i])
            end
            
            rxBuf = {}
            return {Status = "0", Variable = {
                utils.AppendVariable(0, "meter_address", "表地址", "string", addr, "")
            }}
        else
            print("未识别的数据标识:", dataId)
        end
        
    -- 阀门控制响应 (84h)
    elseif ctrlCode == 0x84 then
        -- 根据619协议，数据标识在16-17位置
        local dataId = string.format("%02X%02X", rxBuf[16], rxBuf[17])
        
        if dataId == "AA05" then
            -- 根据619协议，成功标志在18位置
            local status = rxBuf[18] or 0xFF
            rxBuf = {}
            
            if status == 0 then
                return {Status = "0", Variable = {
                    utils.AppendVariable(0, "dev_status", "阀门状态", "string", "成功", "成功")
                }}
            else
                return {Status = "0", Variable = {
                    utils.AppendVariable(0, "dev_status", "阀门状态", "string", "失败", "失败")
                }}
            end
        elseif dataId == "AA07" then
            -- 设置表地址响应
            -- 根据619协议，成功标志在18位置
            local status = rxBuf[18] or 0xFF
            rxBuf = {}
            
            if status == 0 then
                return {Status = "0", Variable = {
                    utils.AppendVariable(0, "new_address", "新表地址", "string", "设置成功", "设置成功")
                }}
            else
                return {Status = "0", Variable = {
                    utils.AppendVariable(0, "new_address", "新表地址", "string", "设置失败", "设置失败")
                }}
            end
        elseif dataId == "AA09" then
            -- 设置表底数响应
            -- 根据619协议，成功标志在18位置
            local status = rxBuf[18] or 0xFF
            rxBuf = {}
            
            if status == 0 then
                return {Status = "0", Variable = {
                    utils.AppendVariable(0, "set_value", "设置底数", "string", "成功", "")
                }}
            else
                return {Status = "0", Variable = {
                    utils.AppendVariable(0, "set_value", "设置底数", "string", "失败", "")
                }}
            end
        end
    end
    
    -- 未知的响应或解析失败
    print("未能识别的响应或解析失败")
    -- 打印接收到的完整数据以便调试
    local dataHex = ""
    for i = 1, #rxBuf do
        dataHex = dataHex .. string.format("%02X ", rxBuf[i])
    end
    print("接收数据:", dataHex)
    
    rxBuf = {}
    return {Status = "1", Variable = {}}
end

function GenerateGetFlow(sAddr, continued)
    local cmd = { 0x90, 0x1F }  -- 读取流量数据标识
    local requestADU = GenerateCommand(sAddr, cmd)
    return {Status = continued, Variable = requestADU}
end

function GenerateGetRealVariables(sAddr, step)
    print("2025F183-37 ver 1.0", sAddr, step)
    if (step == 0) then
        return GenerateGetFlow(sAddr, "0")
    end
end

-- 解析命令参数
function parseCommandParams(cmdParam)
    local params = {}
    if cmdParam and cmdParam ~= "" then
        local success, result = pcall(json.jsondecode, cmdParam)
        if success and result and type(result) == "table" then
            params = result
        end
    end
    return params
end

function DeviceCustomCmd(sAddr, cmdName, cmdParam, step)
    local params = parseCommandParams(cmdParam)
    
    if (cmdName == "dev_flow" or cmdName == "dev_low") then
        return GenerateGetFlow(sAddr, "0")
    elseif (cmdName == "OpenValve") then
        return {Status = "0", Variable = GenerateOpenValve(sAddr)}
    elseif (cmdName == "CloseValve") then
        return {Status = "0", Variable = GenerateCloseValve(sAddr)}
    elseif (cmdName == "ReadAddress") then
        return {Status = "0", Variable = GenerateReadAddress()}
    elseif (cmdName == "SetAddress") then
        local oldAddr = params.oldAddress or "AAAAAAAA"
        local newAddr = params.newAddress
        if not newAddr then
            print("错误: 未提供新地址")
            return {Status = "1", Variable = {}}
        end
        return {Status = "0", Variable = GenerateSetAddress(oldAddr, newAddr)}
    elseif (cmdName == "SetValue") then
        local value = params.value
        if not value then
            print("错误: 未提供设置值")
            return {Status = "1", Variable = {}}
        end
        return {Status = "0", Variable = GenerateSetValue(sAddr, value)}
    end
    
    return {Status = "0", Variable = {}}
end

-- 获取支持的命令列表
function GetSupportedCommands()
    return {
        {name = "dev_flow", desc = "读取总流量"},
        {name = "OpenValve", desc = "开阀控制"},
        {name = "CloseValve", desc = "关阀控制"},
        {name = "ReadAddress", desc = "读取表地址"},
        {name = "SetAddress", desc = "设置表地址", params = "oldAddress,newAddress"},
        {name = "SetValue", desc = "设置表底数", params = "value"}
    }
end

-- 获取插件信息
function GetPluginInfo()
    return {
        name = "2025F183-37",
        version = "1.0.2",
        author = "Claude",
        description = "基于619规约的CAT1水表通信协议插件，支持网口采集器，可处理主动上报数据",
        protocol = "2025F183-37"
    }
end 