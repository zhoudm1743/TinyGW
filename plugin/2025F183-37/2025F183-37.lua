---
--- 2025F183-37.lua
--- 基于619规约的CAT1水表通信协议实现
--- 创建时间: 2023/11/08
--- 修订时间: 2023/11/08 (v2.0.0)
--- 修订说明: 
--- 1. 完全重写代码结构，使代码更清晰，符合619协议规范
--- 2. 优化阀门操作部分，确保命令格式正确
--- 3. 修正地址格式，遵循协议的"低字节在前"规范
--- 4. 优化数据解析逻辑，包括主动上报和命令响应
--- 5. 增强错误处理和调试功能
---

-- 引入必要的库
package.path = "./plugin/2025F183-37/?.lua;"
local json = require("json")
local utils = require("utils")
local bit = require("bit")

-- 常量定义
local CONSTANTS = {
    -- 帧头尾标志
    START_FLAG = 0x68,
    END_FLAG = 0x16,
    -- 控制码
    CTRL_READ_REQUEST = 0x01,  -- 服务器向表端下发读指令
    CTRL_READ_RESPONSE = 0x81, -- 表端对读指令的响应
    CTRL_WRITE_REQUEST = 0x04, -- 服务器向表端下发写指令
    CTRL_WRITE_RESPONSE = 0x84, -- 表端对写指令的响应
    CTRL_UPLOAD = 0x97,        -- 表端主动上报
    CTRL_UPLOAD_RESPONSE = 0x17, -- 服务器对上报的响应
    -- 数据标识
    DATA_ID_READ_PARAM = 0xA901, -- 读表参数
    DATA_ID_READ_FLOW = 0x901F,  -- 读取流量数据
    DATA_ID_VALVE = 0xAA05,      -- 阀门操作
    DATA_ID_SET_ADDR = 0xAA07,   -- 设置表地址
    DATA_ID_SET_VALUE = 0xAA09,  -- 设置表底数
    -- 阀门操作代码
    VALVE_OPEN = 0x55,          -- 开阀
    VALVE_CLOSE = 0x99,         -- 关阀
    VALVE_FORCE = 0x5A,         -- 强制操作
}

-- 协议工具函数
local Protocol = {}

-- 计算校验和
function Protocol.calculateChecksum(buffer, startIndex, endIndex)
    local sum = 0
    for i = startIndex, endIndex do
        sum = sum + buffer[i]
    end
    return sum % 256
end

-- 生成命令头部
function Protocol.generateCommandHeader(address, controlCode)
    local cmd = {
        CONSTANTS.START_FLAG,  -- 起始帧
        0x10,                  -- 表类型(水表)
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, -- 表地址(7字节)
        0x03,                  -- 设备类型
        controlCode,           -- 控制码
        0x00, 0x00,           -- 指令编号
        0x00, 0x00            -- 备用字段
    }
    
    -- 确保地址格式正确
    local addr = string.format("%014s", address):gsub("[^0-9]", "0")
    if #addr ~= 14 then
        addr = string.rep("0", 14 - #addr) .. addr
    end

    -- 表类型固定为水表0x10
    cmd[2] = 0x10
    
    -- 按照619协议格式填充地址字段
    -- 619协议表地址格式：A6(高字节) ~ A0(低字节)
    for i = 0, 6 do
        local pos = 13 - i * 2
        local byteValue = tonumber(addr:sub(pos-1, pos), 10) or 0
        cmd[9-i] = byteValue
    end
    
    -- 设置随机指令编号
    cmd[12] = math.random(0, 255)
    cmd[13] = math.random(0, 255)
    
    return cmd
end

-- 设置数据长度和数据标识
function Protocol.setDataField(cmd, dataId, dataLength)
    -- 设置数据长度(1字节，符合表格第8项规范)
    cmd[16] = bit.band(dataLength, 0xFF)
    
    -- 设置数据标识(2字节)
    cmd[17] = math.floor(dataId / 256)
    cmd[18] = dataId % 256
    
    return cmd
end

-- 完成命令生成，添加校验码和结束标志
function Protocol.finalizeCommand(cmd)
    -- 计算并添加校验码
    table.insert(cmd, Protocol.calculateChecksum(cmd, 1, #cmd))
    -- 添加结束标志
    table.insert(cmd, CONSTANTS.END_FLAG)
    return cmd
end

-- 生成读取命令
function Protocol.generateReadCommand(address, dataId, extraData)
    extraData = extraData or {}
    
    -- 创建命令头
    local cmd = Protocol.generateCommandHeader(address, CONSTANTS.CTRL_READ_REQUEST)
    
    -- 设置数据长度和数据标识
    -- 默认数据长度：2字节数据标识 + 8字节备用
    Protocol.setDataField(cmd, dataId, 10)
    
    -- 添加额外数据(如果有)
    for _, value in ipairs(extraData) do
        table.insert(cmd, value)
    end
    
    -- 如果需要8字节备用字段且未添加
    local extraDataLen = #extraData
    if extraDataLen < 8 then
        for i = 1, 8 - extraDataLen do
            table.insert(cmd, 0x00)
        end
    end
    
    -- 完成命令
    return Protocol.finalizeCommand(cmd)
end

-- 生成写入命令
function Protocol.generateWriteCommand(address, dataId, data)
    data = data or {}
    
    -- 创建命令头
    local cmd = Protocol.generateCommandHeader(address, CONSTANTS.CTRL_WRITE_REQUEST)
    
    -- 设置数据长度和数据标识
    -- 数据长度：2字节数据标识 + 实际数据长度
    Protocol.setDataField(cmd, dataId, 2 + #data)
    
    -- 添加数据
    for _, value in ipairs(data) do
        table.insert(cmd, value)
    end
    
    -- 完成命令
    return Protocol.finalizeCommand(cmd)
end

-- 根据表格规范重写阀门操作命令生成函数
function Protocol.generateValveCommand(address, isOpen, isForce)
    -- 创建一个精确30字节的阀门操作命令
    local cmd = {
        0x68,  -- 1. 起始帧
        0x10,  -- 2. 表类型(水表)
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  -- 3. 表地址(7字节，将在后面设置)
        0x03,  -- 4. 设备类型
        0x04,  -- 5. 控制码(04H: 服务器对表端下发写指令)
        0xBE, 0x6E,  -- 6. 指令编号(固定为BE6E)
        0x00, 0x00,  -- 7. 备用(2字节)
        0x0C,  -- 8. 数据长度(1字节，表示12字节)
        0xAA, 0x05,  -- 9. 数据标识(AA05H: 阀门控制)
        isOpen and 0x55 or 0x99,  -- 10. 阀门开(0x55)/关(0x99)
        isForce and 0x5A or 0x00,  -- 11. 操作类型(0x5A强制，其他非强制)
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  -- 12. 备用(8字节)
        0x00,  -- 13. 校验码(将在后面计算)
        0x16   -- 14. 结束帧
    }
    
    -- 解析表地址并设置到命令中
    if type(address) == "string" and #address > 0 then
        -- 处理地址
        if #address == 14 then
            -- 如果是14位BCD码格式，需要反转为01001906255282格式
            -- 将82522506190001转换为01001906255282
            local reversedAddr = ""
            for i = 7, 1, -1 do
                reversedAddr = reversedAddr .. address:sub(i*2-1, i*2)
            end
            
            -- 按照619协议格式填充地址字段
            for i = 1, 7 do
                local byteValue = tonumber(reversedAddr:sub(i*2-1, i*2), 16) or 0
                cmd[i+2] = byteValue
            end
        else
            -- 处理其他长度的地址，直接按照字节顺序填充
            local addrBytes = {}
            
            -- 解析地址字符串为字节数组
            for i = 1, #address, 2 do
                if i+1 <= #address then
                    local byteStr = address:sub(i, i+1)
                    local byteVal = tonumber(byteStr, 16) or 0
                    table.insert(addrBytes, byteVal)
                end
            end
            
            -- 填充地址字段，最多7个字节
            for i = 1, math.min(#addrBytes, 7) do
                cmd[i+2] = addrBytes[i]
            end
        end
    end
    
    -- 修改校验和计算逻辑
    -- 根据文档，校验码是起始帧0x68至校验码前所有数据的和取模256
    local checksum = 0
    for i = 1, #cmd - 2 do  -- 从起始帧(第1字节)到校验码前(倒数第2字节)
        checksum = checksum + cmd[i]
    end
    cmd[#cmd - 1] = checksum % 256
    
    return cmd
end

-- 辅助函数：将命令转换为十六进制字符串
function toHexString(cmd)
    local hexStr = ""
    for i = 1, #cmd do
        hexStr = hexStr .. string.format("%02X", cmd[i])
    end
    return hexStr
end

-- 辅助函数：打印命令内容
function printCommand(cmd)
    local hexStr = ""
    for i = 1, #cmd do
        hexStr = hexStr .. string.format("%02X ", cmd[i])
    end
    print("命令: " .. hexStr)
    return toHexString(cmd)
end

-- 修改GenerateOpenValve函数，使用通用的阀门命令生成函数
function GenerateOpenValve(address)
    -- 使用通用的阀门命令生成函数生成开阀命令
    local cmd = Protocol.generateValveCommand(address, true, true)
    return cmd
end

-- 修改GenerateCloseValve函数，使用通用的阀门命令生成函数
function GenerateCloseValve(address)
    -- 使用通用的阀门命令生成函数生成关阀命令
    local cmd = Protocol.generateValveCommand(address, false, true)
    return cmd
end

-- 生成读取流量命令
function GenerateGetFlow(address, continued)
    local cmd = Protocol.generateReadCommand(address, CONSTANTS.DATA_ID_READ_FLOW)
    return {Status = continued, Variable = cmd}
end

-- 生成设置表底数命令
function GenerateSetValue(address, value)
    -- 将值转换为4字节数据，低字节在前
    local valueBytes = {}
    local v = tonumber(value) or 0
    
    table.insert(valueBytes, v % 256)
    table.insert(valueBytes, math.floor(v / 256) % 256)
    table.insert(valueBytes, math.floor(v / 65536) % 256)
    table.insert(valueBytes, math.floor(v / 16777216) % 256)
    
    -- 添加8字节备用字段
    for i = 1, 8 do
        table.insert(valueBytes, 0x00)
    end
    
    return Protocol.generateWriteCommand(address, CONSTANTS.DATA_ID_SET_VALUE, valueBytes)
end

-- 生成读取表地址命令
function GenerateReadAddress()
    -- 使用广播地址读取表参数
    local broadcastAddr = "AAAAAAAAAAAAAA"  -- 14个A表示广播地址
    return Protocol.generateReadCommand(broadcastAddr, CONSTANTS.DATA_ID_READ_PARAM)
end

-- 生成设置表地址命令
function GenerateSetAddress(oldAddr, newAddr)
    -- 创建命令头
    local cmd = Protocol.generateCommandHeader(oldAddr, CONSTANTS.CTRL_WRITE_REQUEST)
    
    -- 准备数据部分
    local data = {
        0xAA, 0x07,  -- 数据标识(AA07H: 设置表具类型)
        0x00,        -- 计量模式
        0x00,        -- 付费模式
        0x00,        -- 到位模式
        0xC1         -- 地址修改使能(0xC1: 允许修改表号)
    }
    
    -- 添加新地址
    local addr = string.format("%014s", newAddr):gsub("[^0-9]", "0")
    for i = 0, 6 do
        local pos = 13 - i * 2
        local byteValue = tonumber(addr:sub(pos-1, pos), 10) or 0
        table.insert(data, byteValue)
    end
    
    -- 设置数据长度
    Protocol.setDataField(cmd, 0, #data)
    
    -- 添加数据
    for _, value in ipairs(data) do
        table.insert(cmd, value)
    end
    
    -- 完成命令
    return Protocol.finalizeCommand(cmd)
end

-- 生成周期读取命令
function GenerateGetRealVariables(address, step)
    if step == 0 then
        return GenerateGetFlow(address, "0")
    end
    return {Status = "0", Variable = {}}
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

-- 修改DeviceCustomCmd函数中的Status返回值
function DeviceCustomCmd(address, cmd, params, step)
    -- 处理开阀命令
    if cmd == "valve_open" or cmd == "OpenValve" then
        local openValveCmd = GenerateOpenValve(address)
        return {Status = "0", Variable = openValveCmd}
    
    -- 处理关阀命令
    elseif cmd == "valve_close" or cmd == "CloseValve" then
        local closeValveCmd = GenerateCloseValve(address)
        return {Status = "0", Variable = closeValveCmd}
    
    -- 处理读取流量命令
    elseif cmd == "dev_flow" then
        return GenerateGetFlow(address, "1")
    
    -- 处理读取表地址命令
    elseif cmd == "ReadAddress" then
        return {Status = "0", Variable = GenerateReadAddress()}
    
    -- 处理设置表地址命令
    elseif cmd == "SetAddress" then
        -- 解析参数
        local oldAddr = address
        local newAddr = ""
        if params and params ~= "" then
            local parts = {}
            for part in string.gmatch(params, "[^,]+") do
                table.insert(parts, part)
            end
            newAddr = parts[1] or "01001906255283"
        end
        return {Status = "0", Variable = GenerateSetAddress(oldAddr, newAddr)}
    
    -- 处理设置表底数命令
    elseif cmd == "SetValue" then
        -- 解析参数
        local value = "0"
        if params and params ~= "" then
            local parts = {}
            for part in string.gmatch(params, "[^,]+") do
                table.insert(parts, part)
            end
            value = parts[1] or "0"
        end
        return {Status = "0", Variable = GenerateSetValue(address, value)}
    
    -- 未知命令
    else
        print("未知命令:", cmd)
        return {Status = "0", Variable = {}}
    end
end

-- 生成设备命令
function GenerateDeviceCommand(address, cmdName, cmdParams)
    local result = {
        Status = "0",
        Variable = {}
    }
    
    if cmdName == "valve_open" then
        result.Variable = GenerateOpenValve(address)
    elseif cmdName == "valve_close" then
        result.Variable = GenerateCloseValve(address)
    else
        result.Status = "1"
    end
    
    return result
end

-- 数据接收和解析
rxBuf = {}
function AnalysisRx(address, rxBufCnt)
    -- 调试信息
    print("AnalysisRx调用: 地址=", address, "数据长度=", rxBufCnt)
    
    -- 验证接收数据有效性
    if rxBufCnt < 16 then
        print("错误: 数据长度不足")
        rxBuf = {}
        return {Status = "1", Variable = {}}
    end
    
    -- 查找起始符
    local startIndex = 1
    while startIndex <= #rxBuf and rxBuf[startIndex] ~= CONSTANTS.START_FLAG do
        startIndex = startIndex + 1
    end
    
        if startIndex > #rxBuf then
        print("错误: 未找到起始符")
            rxBuf = {}
        return {Status = "1", Variable = {}}
    end
    
    -- 移除唤醒符
    while startIndex > 1 and rxBuf[startIndex - 1] == 0xFE do
        table.remove(rxBuf, startIndex - 1)
        startIndex = startIndex - 1
    end

    -- 查找结束符和校验校验和
    local endIndex = 0
    local csFlag, endFlag
    
    for i = startIndex + 15, #rxBuf do  -- 至少16字节的帧
        if rxBuf[i] == CONSTANTS.END_FLAG then
            local cs = Protocol.calculateChecksum(rxBuf, startIndex, i - 2)
            if cs == rxBuf[i-1] then
                endFlag = rxBuf[i]
                csFlag = rxBuf[i-1]
                endIndex = i
                break
            end
        end
    end
    
    if endIndex == 0 then
        print("错误: 未找到有效结束符或校验和错误")
        rxBuf = {}
        return {Status = "1", Variable = {}}
    end
    
    -- 提取表地址
    local addr = ""
    for i = startIndex + 1, startIndex + 8 do
        if i <= #rxBuf then
            addr = addr .. string.format("%02X", rxBuf[i])
        end
    end
    print("接收数据表地址:", addr)
    
    -- 提取控制码
    local ctrlCode = rxBuf[startIndex + 10] or 0
    print("控制码:", string.format("0x%02X", ctrlCode))
    
    -- 打印完整报文内容(调试用)
    local dataHex = ""
    for i = startIndex, endIndex do
        dataHex = dataHex .. string.format("%02X ", rxBuf[i])
    end
    print("完整报文:", dataHex)
    
    -- 处理不同类型的响应
    local variables = {}
    
    -- 表端主动上报数据(0x97)
    if ctrlCode == CONSTANTS.CTRL_UPLOAD then
        -- 获取上报类型和数据包类型
        local reportType = 0
        if rxBuf[startIndex + 13] and rxBuf[startIndex + 14] then
            reportType = (rxBuf[startIndex + 13] * 256) + rxBuf[startIndex + 14]
        end
        print("上报类型:", string.format("0x%04X", reportType))
        
        local dataPackType = rxBuf[startIndex + 16] or 0
        print("数据包类型:", string.format("0x%02X", dataPackType))
        
        -- 解析主动上报数据
        if dataPackType == 0x01 or dataPackType == 0x02 or dataPackType == 0x03 then
            -- 解析状态字
            local status = 0
            if rxBuf[startIndex + 58] and rxBuf[startIndex + 59] then
                status = (rxBuf[startIndex + 58] * 256) + rxBuf[startIndex + 59]
            end
            table.insert(variables, utils.AppendVariable(0, "dev_status", "状态", "int", status, tostring(status)))
            
            -- 解析累计用量/金额
            local totalValue = 0
            if dataPackType == 0x01 or dataPackType == 0x02 then
                -- 预付费表在位置61-64
                if rxBuf[startIndex + 61] and rxBuf[startIndex + 64] then
                    totalValue = rxBuf[startIndex + 61] + 
                                (rxBuf[startIndex + 62] * 256) + 
                                (rxBuf[startIndex + 63] * 65536) + 
                                (rxBuf[startIndex + 64] * 16777216)
                end
            elseif dataPackType == 0x03 then
                -- 后付费表在位置60-63
                if rxBuf[startIndex + 60] and rxBuf[startIndex + 63] then
                    totalValue = rxBuf[startIndex + 60] + 
                                (rxBuf[startIndex + 61] * 256) + 
                                (rxBuf[startIndex + 62] * 65536) + 
                                (rxBuf[startIndex + 63] * 16777216)
                end
            end
            table.insert(variables, utils.AppendVariable(1, "dev_flow", "总流量", "double", totalValue/10.0, tostring(totalValue/10.0)))
            
            -- 获取电池电压
            local voltage = 0
            if rxBuf[startIndex + 43] and rxBuf[startIndex + 44] then
                voltage = (rxBuf[startIndex + 43] * 256) + rxBuf[startIndex + 44]
                voltage = voltage / 100.0  -- 单位V
            end
            table.insert(variables, utils.AppendVariable(2, "dev_voltage", "电池电压", "double", voltage, tostring(voltage)))
            
            -- 获取信号强度
            local signal = 0
            if rxBuf[startIndex + 45] then
                signal = rxBuf[startIndex + 45]
            end
            table.insert(variables, utils.AppendVariable(3, "dev_signal", "信号强度", "int", signal, tostring(signal)))
        elseif dataPackType == 0x3E then
            -- 特殊数据包类型0x3E的处理
            -- 解析水表示数(位置51-52)
            local waterReading = 0
            if rxBuf[startIndex + 51] and rxBuf[startIndex + 52] then
                waterReading = (rxBuf[startIndex + 51] * 256) + rxBuf[startIndex + 52]
                waterReading = waterReading / 50.0  -- 比例因子1/50
            end
            table.insert(variables, utils.AppendVariable(0, "dev_flow", "总流量", "double", waterReading, tostring(waterReading)))
            
            -- 解析电池电压
            local voltage = 0
            if rxBuf[startIndex + 43] and rxBuf[startIndex + 44] then
                voltage = (rxBuf[startIndex + 43] * 256) + rxBuf[startIndex + 44]
                voltage = voltage / 100.0  -- 单位V
            end
            table.insert(variables, utils.AppendVariable(1, "dev_voltage", "电池电压", "double", voltage, tostring(voltage)))
            
            -- 解析信号强度
            local signal = 0
            if rxBuf[startIndex + 45] then
                signal = rxBuf[startIndex + 45]
            end
            table.insert(variables, utils.AppendVariable(2, "dev_signal", "信号强度", "int", signal, tostring(signal)))
            
            -- 解析状态字
            local status = 0
            if rxBuf[startIndex + 46] and rxBuf[startIndex + 47] then
                status = (rxBuf[startIndex + 46] * 256) + rxBuf[startIndex + 47]
            end
            table.insert(variables, utils.AppendVariable(3, "dev_status", "状态", "int", status, tostring(status)))
            
            -- 解析阀门状态
            local valveStatus = "未知"
            if status % 2 == 1 then
                valveStatus = "关闭"
            else
                valveStatus = "开启"
            end
            table.insert(variables, utils.AppendVariable(4, "valve_status", "阀门状态", "string", valveStatus, valveStatus))
            
            -- 解析电池状态
            local batteryStatus = "未知"
            if math.floor(status / 2) % 2 == 1 then
                batteryStatus = "电量低"
            else
                batteryStatus = "正常"
            end
            table.insert(variables, utils.AppendVariable(5, "battery_status", "电池状态", "string", batteryStatus, batteryStatus))
        else
            print("未处理的数据包类型:", dataPackType)
            rxBuf = {}
        return {Status = "1", Variable = {}}
    end
    -- 读数据响应(0x81)
    elseif ctrlCode == CONSTANTS.CTRL_READ_RESPONSE then
        -- 获取数据标识
        local dataId = 0
        if rxBuf[startIndex + 16] and rxBuf[startIndex + 17] then
            dataId = (rxBuf[startIndex + 16] * 256) + rxBuf[startIndex + 17]
        end
        print("数据标识:", string.format("0x%04X", dataId))
        
        -- 读取流量数据(901F)
        if dataId == CONSTANTS.DATA_ID_READ_FLOW then
            -- 提取4字节流量数据
            if startIndex + 21 <= endIndex then
                local flow = rxBuf[startIndex + 18] + 
                            (rxBuf[startIndex + 19] * 256) + 
                            (rxBuf[startIndex + 20] * 65536) + 
                            (rxBuf[startIndex + 21] * 16777216)
                
                table.insert(variables, utils.AppendVariable(0, "dev_flow", "总流量", "double", flow/10.0, tostring(flow/10.0)))
            end
        -- 读取表参数(A901)
        elseif dataId == CONSTANTS.DATA_ID_READ_PARAM then
            -- 提取表地址
            local meterAddr = ""
            for i = startIndex + 2, startIndex + 8 do
                meterAddr = meterAddr .. string.format("%02X", rxBuf[i])
            end
            table.insert(variables, utils.AppendVariable(0, "meter_address", "表地址", "string", meterAddr, ""))
        else
            print("未处理的数据标识:", string.format("0x%04X", dataId))
        end
    -- 写数据响应(0x84)
    elseif ctrlCode == CONSTANTS.CTRL_WRITE_RESPONSE then
        -- 获取数据标识
        local dataId = 0
        if rxBuf[startIndex + 16] and rxBuf[startIndex + 17] then
            dataId = (rxBuf[startIndex + 16] * 256) + rxBuf[startIndex + 17]
        end
        
        -- 获取成功标志
        local status = rxBuf[startIndex + 18] or 0xFF
        
        -- 阀门操作响应
        if dataId == CONSTANTS.DATA_ID_VALVE then
            if status == 0 then
                table.insert(variables, utils.AppendVariable(0, "dev_status", "阀门状态", "string", "成功", "成功"))
            else
                table.insert(variables, utils.AppendVariable(0, "dev_status", "阀门状态", "string", "失败", "失败"))
            end
            -- 设置表地址响应
        elseif dataId == CONSTANTS.DATA_ID_SET_ADDR then
            if status == 0 then
                table.insert(variables, utils.AppendVariable(0, "new_address", "新表地址", "string", "设置成功", "设置成功"))
            else
                table.insert(variables, utils.AppendVariable(0, "new_address", "新表地址", "string", "设置失败", "设置失败"))
            end
            -- 设置表底数响应
        elseif dataId == CONSTANTS.DATA_ID_SET_VALUE then
            if status == 0 then
                table.insert(variables, utils.AppendVariable(0, "set_value", "设置底数", "string", "成功", ""))
            else
                table.insert(variables, utils.AppendVariable(0, "set_value", "设置底数", "string", "失败", ""))
            end
        else
            print("未处理的写操作数据标识:", string.format("0x%04X", dataId))
        end
    else
        print("未处理的控制码:", string.format("0x%02X", ctrlCode))
    end
    
    -- 清空接收缓冲区并返回解析结果
    rxBuf = {}
    return {Status = "0", Variable = variables}
end

-- 获取支持的命令列表
function GetSupportedCommands()
    return {
        {name = "dev_flow", desc = "读取总流量"},
        {name = "valve_open", desc = "开阀控制"},
        {name = "valve_close", desc = "关阀控制"},
        {name = "OpenValve", desc = "开阀控制(兼容旧版)"},
        {name = "CloseValve", desc = "关阀控制(兼容旧版)"},
        {name = "ReadAddress", desc = "读取表地址"},
        {name = "SetAddress", desc = "设置表地址", params = "newAddress"},
        {name = "SetValue", desc = "设置表底数", params = "value"}
    }
end

-- 获取插件信息
function GetPluginInfo()
    return {
        name = "2025F183-37",
        version = "2.0.0",
        author = "Claude",
        description = "基于619规约的CAT1水表通信协议插件，优化阀门操作，支持主动上报数据解析",
        protocol = "2025F183-37"
    }
end 