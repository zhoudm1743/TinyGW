---
--- 619系列4G CAT1远传水表通讯协议
--- 福建百悦信息科技有限公司
--- Created by: tiny-gw
--- DateTime: 2024/12/15
---

package.path = "./plugin/619-BY/?.lua;"

require "utils"
require "json"

-- 全局变量
rxBuf = {}

-------公共函数-------

-- 校验和计算：从起始帧到校验码前所有字节累加，取低8位
function checkSum(buffer, startIndex, endIndex)
    local sum = 0
    for i = startIndex, endIndex do
        sum = sum + buffer[i]
    end
    return sum % 256
end

-- 表地址转换：14位BCD码，低字节在前
function convertAddress(requestADU, sAddr)
    local addr = string.format("%014s", sAddr)
    
    -- BCD码转换，低字节在前
    requestADU[3] = tonumber(string.sub(addr, 13, 14))  -- A0
    requestADU[4] = tonumber(string.sub(addr, 11, 12))  -- A1
    requestADU[5] = tonumber(string.sub(addr, 9, 10))   -- A2
    requestADU[6] = tonumber(string.sub(addr, 7, 8))    -- A3
    requestADU[7] = tonumber(string.sub(addr, 5, 6))    -- A4
    requestADU[8] = tonumber(string.sub(addr, 3, 4))    -- A5
    requestADU[9] = tonumber(string.sub(addr, 1, 2))    -- A6
end

-- 生成基础指令帧
function GenerateCommand(sAddr, controlCode, cmdId, dataLength, dataBytes)
    local cmdId = cmdId or 0x0001
    local dataLength = dataLength or 0
    local dataBytes = dataBytes or {}
    
    -- 基础帧结构: 68 10 [7字节地址] 03 [控制码] [2字节指令] [2字节备用] [1字节数据长度] [数据域] [校验] 16
    local requestADU = {
        0x68,           -- 1: 起始帧
        0x10,           -- 2: 表类型（水表）
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  -- 3-9: 表地址(7字节)
        0x03,           -- 10: 设备类型
        controlCode,    -- 11: 控制码
        math.floor(cmdId / 256), cmdId % 256,       -- 12-13: 指令编号
        0x00, 0x00,     -- 14-15: 备用
        dataLength      -- 16: 数据长度
    }
    
    -- 设置表地址
    convertAddress(requestADU, sAddr)
    
    -- 添加数据字节
    for i, byte in ipairs(dataBytes) do
        table.insert(requestADU, byte)
    end
    
    -- 计算校验码
    local checksum = checkSum(requestADU, 1, #requestADU)
    table.insert(requestADU, checksum)
    
    -- 结束帧
    table.insert(requestADU, 0x16)
    
    return requestADU
end

-------指令生成函数-------

-- 读取表参数
function GenerateReadParams(sAddr)
    local dataBytes = {0xA9, 0x01}  -- 数据标识：表参数
    -- 添加8字节备用数据
    for i = 1, 8 do
        table.insert(dataBytes, 0x00)
    end
    return GenerateCommand(sAddr, 0x01, 0x0001, 10, dataBytes)
end

-- 阀门操作
function GenerateValveOperation(sAddr, operation)
    local dataBytes = {
        0xAA, 0x05,    -- 数据标识：阀门操作
        operation,     -- 操作码：0x55开阀, 0x99关阀
        0x00          -- 操作类型：0x5A强制，其他非强制
    }
    -- 添加8字节备用数据
    for i = 1, 8 do
        table.insert(dataBytes, 0x00)
    end
    return GenerateCommand(sAddr, 0x04, 0x0001, 12, dataBytes)
end

-- 开阀
function GenerateOpenValve(sAddr)
    return GenerateValveOperation(sAddr, 0x55)
end

-- 关阀
function GenerateCloseValve(sAddr)
    return GenerateValveOperation(sAddr, 0x99)
end

-- 时钟校对
function GenerateClockSync(sAddr)
    local dataBytes = {
        0xAA, 0x00,    -- 数据标识：时钟校对
        0x5A,          -- 使能标识
        0x24, 0x12, 0x15, 0x14, 0x30, 0x00  -- 时间：2024-12-15 14:30:00 (BCD)
    }
    -- 添加8字节备用数据
    for i = 1, 8 do
        table.insert(dataBytes, 0x00)
    end
    return GenerateCommand(sAddr, 0x04, 0x0001, 17, dataBytes)
end

-------数据解析函数-------

-- 解析主动上报数据
function parseUploadData(buffer, dataStart, dataLen)
    local variables = {}
    local dataType = buffer[dataStart]
    
    print("数据包类型:", string.format("%02X", dataType))
    
    if dataType == 0x01 then  -- 预付费普通表上报
        -- 解析电压 (位置相对数据起始点)
        if dataStart + 28 <= #buffer then
            local voltage = buffer[dataStart + 27] * 256 + buffer[dataStart + 28]
            table.insert(variables, utils.AppendVariable(1, "voltage", "电压", "int", voltage, voltage .. "mV"))
        end
        
        -- 解析累计用量 (4字节，小端序)
        if dataStart + 48 <= #buffer then
            local totalUsage = buffer[dataStart + 45] + 
                              buffer[dataStart + 46] * 256 + 
                              buffer[dataStart + 47] * 65536 + 
                              buffer[dataStart + 48] * 16777216
            table.insert(variables, utils.AppendVariable(2, "total_usage", "累计用量", "double", totalUsage, string.format("%.2f", totalUsage / 100.0) .. "L"))
        end
        
        -- 解析剩余量
        if dataStart + 52 <= #buffer then
            local remainUsage = buffer[dataStart + 49] + 
                               buffer[dataStart + 50] * 256 + 
                               buffer[dataStart + 51] * 65536 + 
                               buffer[dataStart + 52] * 16777216
            table.insert(variables, utils.AppendVariable(3, "remain_usage", "剩余量", "double", remainUsage, string.format("%.2f", remainUsage / 100.0) .. "L"))
        end
        
    elseif dataType == 0x03 then  -- 后付费表上报
        if dataStart + 47 <= #buffer then
            local totalUsage = buffer[dataStart + 44] + 
                              buffer[dataStart + 45] * 256 + 
                              buffer[dataStart + 46] * 65536 + 
                              buffer[dataStart + 47] * 16777216
            table.insert(variables, utils.AppendVariable(2, "total_usage", "累计用量", "double", totalUsage, string.format("%.2f", totalUsage / 100.0) .. "L"))
        end
    end
    
    return variables
end

-- 主数据解析函数
function AnalysisRx(sAddr, rxBufCnt)
    if rxBufCnt < 12 then
        rxBuf = {}
        print("error: 数据包太短")
        return { Status = "1", Variable = {} }
    end
    
    -- 查找起始符0x68
    local startIndex = 1
    while startIndex <= #rxBuf and rxBuf[startIndex] ~= 0x68 do
        startIndex = startIndex + 1
    end
    
    if startIndex > #rxBuf then
        rxBuf = {}
        print("error: 起始符0x68未找到")
        return { Status = "1", Variable = {} }
    end
    
    -- 查找结束符并验证校验码
    local endIndex = 0
    for i = startIndex + 11, #rxBuf do
        if rxBuf[i] == 0x16 then
            local expectedChecksum = checkSum(rxBuf, startIndex, i - 2)
            if expectedChecksum == rxBuf[i - 1] then
                endIndex = i
                break
            end
        end
    end
    
    if endIndex == 0 then
        rxBuf = {}
        print("error: 找不到有效结束符或校验错误")
        return { Status = "1", Variable = {} }
    end
    
    -- 验证地址 (简化版本)
    local controlCode = rxBuf[startIndex + 10]
    local variables = {}
    
    print("控制码:", string.format("%02X", controlCode))
    
    if controlCode == 0x81 then
        -- 读指令响应
        table.insert(variables, utils.AppendVariable(0, "read_response", "读取响应", "string", "SUCCESS", "读取成功"))
        
    elseif controlCode == 0x84 then
        -- 写指令响应
        table.insert(variables, utils.AppendVariable(0, "write_response", "写入响应", "string", "SUCCESS", "写入成功"))
        
    elseif controlCode == 0x97 then
        -- 主动上报数据
        local dataLength = rxBuf[startIndex + 15]
        if dataLength > 0 and startIndex + 16 <= #rxBuf then
            variables = parseUploadData(rxBuf, startIndex + 16, dataLength)
        end
    end
    
    rxBuf = {}
    return { Status = "0", Variable = variables }
end

-------对外接口函数-------

-- 获取实时变量
function GenerateGetRealVariables(sAddr, step)
    print("619-BY ver 1.0", sAddr, step)
    if step == 0 then
        return { Status = "1", Variable = GenerateReadParams(sAddr) }
    end
    return { Status = "0", Variable = {} }
end

-- 自定义命令处理
function DeviceCustomCmd(sAddr, cmdName, cmdParam, step)
    print("执行命令:", cmdName, "参数:", cmdParam)
    
    if cmdName == "read_params" then
        return { Status = "1", Variable = GenerateReadParams(sAddr) }
        
    elseif cmdName == "open_valve" then
        return { Status = "1", Variable = GenerateOpenValve(sAddr) }
        
    elseif cmdName == "close_valve" then
        return { Status = "1", Variable = GenerateCloseValve(sAddr) }
        
    elseif cmdName == "sync_clock" then
        return { Status = "1", Variable = GenerateClockSync(sAddr) }
        
    else
        print("未知命令:", cmdName)
        return { Status = "0", Variable = {} }
    end
end 