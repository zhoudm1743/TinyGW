package.path = "./plugin/Q3761-1376/?.lua;"
require "utils"
require "json"

-- Q/GDW 1376.1 协议插件
-- 根据3761-BY.md文档实现

-- 常量定义---------------------------------------------------------------------------------------------------------------
local FRAME_START = 0x68
local FRAME_END = 0x16

-- AFN 应用功能码定义
local AFN = {
    CONFIRM_DENY = 0x00,    -- 确认/否认
    RESET = 0x01,           -- 复位
    LINK_TEST = 0x02,       -- 链路接口检测
    SET_PARAM = 0x04,       -- 设置参数
    REQ_CONFIG = 0x09,      -- 请求终端配置及信息
    READ_PARAM = 0x0A,      -- 读取参数
    DATA_FORWARD = 0x10     -- 数据转发
}

-- 功能码定义 (PRM=1时，主站发送)
local FUNC_CODE = {
    RESET = 1,              -- 复位命令
    USER_DATA = 4,          -- 用户数据
    LINK_TEST = 9,          -- 链路测试
    REQ_1_DATA = 10,        -- 请求1级数据
    REQ_2_DATA = 11         -- 请求2级数据
}

-- 功能码定义 (PRM=0时，终端响应)
local RESP_CODE = {
    ACK = 0,                -- 认可
    USER_DATA = 8,          -- 用户数据
    DENY = 9,               -- 否认
    LINK_STATUS = 11        -- 链路状态
}

-- 全局变量
rxBuf = {}
local frameSeq = 0

-- 工具函数---------------------------------------------------------------------------------------------------------------

-- 计算校验和
function checkSum(buffer, startIndex, endIndex)
    local sum = 0
    for i = startIndex, endIndex do
        sum = sum + (buffer[i] or 0)
    end
    return sum % 256
end

-- 16进制转10进制
function hexToDec(hex)
    local dec = 0
    local len = string.len(hex)
    for i = 1, len do
        dec = dec + tonumber(string.sub(hex, len - i + 1, len - i + 1), 16) * 2 ^ ((i - 1) * 4)
    end
    return dec
end

-- BCD码转十进制
function bcdToDec(bcd)
    return math.floor(bcd / 16) * 10 + (bcd % 16)
end

-- 十进制转BCD码
function decToBcd(dec)
    return math.floor(dec / 10) * 16 + (dec % 10)
end

-- 地址转换
function convertAddress(frame, address)
    if not address or #address ~= 9 then
        -- 默认地址示例: 41544924
        frame[8] = 0x41
        frame[9] = 0x54
        frame[10] = 0x49
        frame[11] = 0x24
        frame[12] = 0x00
        return
    end
    
    -- 解析9位地址字符串: "053040961"
    local digits = {}
    for i = 1, 9 do
        digits[i] = tonumber(string.sub(address, i, i)) or 0
    end
    
    -- 前4位作为行政区划码 (BCD编码): 0530
    local region_d1 = digits[1]  -- 0
    local region_d2 = digits[2]  -- 5  
    local region_d3 = digits[3]  -- 3
    local region_d4 = digits[4]  -- 0
    
    -- BCD编码：每字节两个十进制数字，高4位是十位，低4位是个位
    -- 小端传输：低字节在前
    frame[8] = (region_d3 * 16) + region_d4   -- 低字节: 30H (3在高4位，0在低4位)
    frame[9] = (region_d1 * 16) + region_d2   -- 高字节: 05H (0在高4位，5在低4位)
    
    -- 后5位作为终端地址 (二进制编码): 40961  
    local terminal = digits[5] * 10000 + digits[6] * 1000 + digits[7] * 100 + digits[8] * 10 + digits[9]
    frame[10] = terminal % 256         -- 终端地址低字节
    frame[11] = math.floor(terminal / 256) % 256  -- 终端地址高字节
    
    frame[12] = 0x00  -- 主站地址
end

-- 生成当前时间 (BCD格式)
function getCurrentTime()
    local now = os.date("*t")
    return {
        decToBcd(now.sec),   -- 秒
        decToBcd(now.min),   -- 分
        decToBcd(now.hour),  -- 时
        decToBcd(now.day),   -- 日
        decToBcd(now.month) + ((now.wday - 1) * 32),  -- 月+星期
        decToBcd(now.year % 100)  -- 年
    }
end

-- 帧序列号管理
function getNextSeq()
    frameSeq = (frameSeq + 1) % 16
    return frameSeq
end

-- 重置帧序号（用于测试）
function resetFrameSeq()
    frameSeq = 0
end

-- 构建基本帧结构
function buildFrame(address, afn, fn, pn, data, isUplink)
    local frame = {}
    local dataLen = data and #data or 0
    local totalLen = 8 + 4 + dataLen  -- 控制域+地址域+AFN+SEQ+DA+DT+数据
    
    -- 帧头1
    frame[1] = FRAME_START
    
    -- 长度域 (用户数据长度*4 + 协议标识2)
    local len = (totalLen * 4) + 2
    frame[2] = len % 256
    frame[3] = math.floor(len / 256) % 256
    frame[4] = frame[2]  -- 重复
    frame[5] = frame[3]  -- 重复
    
    -- 帧头2
    frame[6] = FRAME_START
    
    -- 控制域C 
    if isUplink then
        -- 上行帧：DIR=1, PRM=1, 功能码=9(链路测试)
        frame[7] = 0xC9  -- 11001001B
    else
        -- 下行帧：DIR=0, PRM=1, 功能码=4(用户数据)
        frame[7] = 0x4B  -- 01001011B
    end
    
    -- 地址域A (5字节)
    convertAddress(frame, address)
    
    -- 应用功能码AFN
    frame[13] = afn
    
    -- 帧序列域SEQ
    if isUplink then
        -- 上行帧：FIR=1, FIN=1, CON=1, PSEQ=7 (需要确认)
        frame[14] = 0x70 + 7  -- 01110111B (0x77)
    else
        -- 下行帧：FIR=1, FIN=1, CON=0, PSEQ
        frame[14] = 0xC0 + getNextSeq()  -- 11000000B + PSEQ
    end
    
    -- 数据单元标识 DA (2字节)
    frame[15] = pn % 256          -- DA1
    frame[16] = math.floor(pn / 256) % 256  -- DA2
    
    -- 数据单元标识 DT (2字节)  
    frame[17] = fn % 256          -- DT1
    frame[18] = math.floor(fn / 256) % 256  -- DT2
    
    -- 数据内容
    local dataIndex = 19
    if data then
        for i = 1, #data do
            frame[dataIndex] = data[i]
            dataIndex = dataIndex + 1
        end
    end
    
    -- 校验和
    frame[dataIndex] = checkSum(frame, 7, dataIndex - 1)
    dataIndex = dataIndex + 1
    
    -- 结束符
    frame[dataIndex] = FRAME_END
    
    return frame
end

-- 协议命令生成函数---------------------------------------------------------------------------------------------------------------

-- 生成登录命令 (AFN=02H, F1, p0)
function generateLoginCommand(address)
    return buildFrame(address, AFN.LINK_TEST, 1, 0, nil, true)  -- 上行帧
end

-- 生成心跳命令 (AFN=02H, F3, p0)
function generateHeartbeatCommand(address)
    local timeData = getCurrentTime()
    return buildFrame(address, AFN.LINK_TEST, 4, 0, timeData, true)  -- 上行帧
end

-- 生成退出登录命令 (AFN=02H, F2, p0)
function generateLogoutCommand(address)
    return buildFrame(address, AFN.LINK_TEST, 2, 0, nil, true)  -- 上行帧
end

-- 生成透明转发命令 (AFN=10H, F1, p0)
function generateTransparentForward(address, port, timeout, data)
    local forwardData = {}
    
    -- 终端通信端口号
    table.insert(forwardData, port or 1)
    
    -- 透明转发通信控制字 (示例: 9600bps, 8位数据位, 1停止位, 无校验)
    table.insert(forwardData, 0x6B)  -- 01101011B
    
    -- 超时时间 (D7=1表示秒, D6-D0表示数值)
    table.insert(forwardData, (timeout or 30) + 0x80)
    
    -- 字节间超时 (10ms)
    table.insert(forwardData, 10)
    
    -- 数据长度 (2字节小端)
    local dataLen = data and #data or 0
    table.insert(forwardData, dataLen % 256)
    table.insert(forwardData, math.floor(dataLen / 256) % 256)
    
    -- 数据内容
    if data then
        for i = 1, #data do
            table.insert(forwardData, data[i])
        end
    end
    
    return buildFrame(address, AFN.DATA_FORWARD, 1, 0, forwardData, false)  -- 下行帧
end

-- 生成读取终端版本信息命令 (AFN=09H, F1, p0)
function generateReadVersionCommand(address)
    return buildFrame(address, AFN.REQ_CONFIG, 1, 0, nil, false)  -- 下行帧
end

-- 生成读取通信模块版本信息命令 (AFN=09H, F9, p0)
function generateReadModuleVersionCommand(address)
    return buildFrame(address, AFN.REQ_CONFIG, 256, 0, nil, false)  -- 下行帧，F9对应bit8，即256
end

-- 数据解析函数---------------------------------------------------------------------------------------------------------------

-- 解析接收帧
function parseFrame(buffer, bufLen)
    if bufLen < 16 then
        return nil, "帧长度不足"
    end
    
    -- 查找帧头
    local startIndex = 1
    for i = 1, bufLen - 15 do
        if buffer[i] == FRAME_START and i + 5 <= bufLen and buffer[i + 5] == FRAME_START then
            startIndex = i
            break
        end
    end
    
    if startIndex >= bufLen - 15 then
        return nil, "未找到有效帧头"
    end
    
    -- 提取长度
    local len1 = buffer[startIndex + 1] or 0
    local len2 = buffer[startIndex + 2] or 0
    local frameLen = math.floor(((len2 * 256 + len1) - 2) / 4)  -- 用户数据长度
    
    -- 完整帧长度 = 起始符(1) + 长度域(4) + 起始符(1) + 用户数据区(frameLen) + 校验(1) + 结束符(1) = frameLen + 8
    local totalFrameLen = startIndex + frameLen + 7  -- 修正：应该是+7而不是+8
    if totalFrameLen > bufLen then
        return nil, "帧不完整，预期长度: " .. totalFrameLen .. ", 实际长度: " .. bufLen
    end
    
    -- 校验和验证 (从控制域到数据区末尾)
    local calcCS = checkSum(buffer, startIndex + 6, startIndex + frameLen + 5)
    local frameCS = buffer[startIndex + frameLen + 6] or 0
    
    if calcCS ~= frameCS then
        return nil, "校验和错误，计算值: " .. calcCS .. ", 帧值: " .. frameCS
    end
    
    -- 解析帧内容
    local frame = {
        startIndex = startIndex,
        totalLen = frameLen + 9,
        ctrl = buffer[startIndex + 6] or 0,
        address = (buffer[startIndex + 10] or 0) * 256 + (buffer[startIndex + 11] or 0),
        masterAddr = buffer[startIndex + 12] or 0,
        afn = buffer[startIndex + 13] or 0,
        seq = buffer[startIndex + 14] or 0,
        dataStart = startIndex + 15,
        dataLen = frameLen - 8,
        checksum = frameCS
    }
    
    return frame, "成功"
end

-- 解析数据单元标识
function parseDataUnitID(frame, buffer)
    if frame.dataLen < 4 then
        return nil, "数据长度不足"
    end
    
    local da1 = buffer[frame.dataStart] or 0
    local da2 = buffer[frame.dataStart + 1] or 0
    local dt1 = buffer[frame.dataStart + 2] or 0
    local dt2 = buffer[frame.dataStart + 3] or 0
    
    local pn = da2 * 256 + da1
    local fn = dt2 * 256 + dt1
    
    return {
        pn = pn,
        fn = fn,
        dataStart = frame.dataStart + 4
    }
end

-- 主要接口函数---------------------------------------------------------------------------------------------------------------

-- 设备自定义命令
function DeviceCustomCmd(sAddr, cmdName, cmdParam, step)
    local params = {}
    if cmdParam and cmdParam ~= "" then
        local success, result = pcall(json.jsondecode, cmdParam)
        if success then
            params = result
        end
    end
    
    if cmdName == "login" then
        local cmd = generateLoginCommand(sAddr)
        return {Status = "0", Variable = cmd}
        
    elseif cmdName == "heartbeat" then
        local cmd = generateHeartbeatCommand(sAddr)
        return {Status = "0", Variable = cmd}
        
    elseif cmdName == "logout" then
        local cmd = generateLogoutCommand(sAddr)
        return {Status = "0", Variable = cmd}
        
    elseif cmdName == "transparent" then
        local port = params.port or 1
        local timeout = params.timeout or 30
        local data = params.data or {}
        local cmd = generateTransparentForward(sAddr, port, timeout, data)
        return {Status = "0", Variable = cmd}
        
    elseif cmdName == "version" then
        local cmd = generateReadVersionCommand(sAddr)
        return {Status = "0", Variable = cmd}
        
    elseif cmdName == "module_version" then
        local cmd = generateReadModuleVersionCommand(sAddr)
        return {Status = "0", Variable = cmd}
    end
    
    -- 默认返回心跳命令
    local cmd = generateHeartbeatCommand(sAddr)
    return {Status = "0", Variable = cmd}
end

-- 数据接收分析
function AnalysisRx(sAddr, rxBufCnt)
    if rxBufCnt < 16 then
        rxBuf = {}
        print("错误: 接收数据长度不足，长度: " .. rxBufCnt)
        return {Status = "1", Variable = {}}
    end
    
    local frame, err = parseFrame(rxBuf, rxBufCnt)
    if not frame then
        rxBuf = {}
        print("帧解析错误: " .. err)
        return {Status = "1", Variable = {}}
    end
    
    local variables = {}
    
    -- 解析数据单元标识
    local dataUnit = parseDataUnitID(frame, rxBuf)
    if not dataUnit then
        rxBuf = {}
        print("数据单元解析错误")
        return {Status = "1", Variable = {}}
    end
    
    -- 根据AFN类型解析数据
    if frame.afn == AFN.LINK_TEST then
        -- 链路检测响应
        if dataUnit.fn == 1 then
            table.insert(variables, utils.AppendVariable(0, "LoginStatus", "登录状态", "String", "登录响应", ""))
        elseif dataUnit.fn == 4 then
            table.insert(variables, utils.AppendVariable(0, "HeartbeatStatus", "心跳状态", "String", "心跳响应", ""))
            -- 解析时间数据
            if frame.dataLen > 10 then
                local timeBytes = {}
                for i = 1, 6 do
                    timeBytes[i] = rxBuf[dataUnit.dataStart + i - 1] or 0
                end
                local timeStr = utils.FormatTime(timeBytes)
                table.insert(variables, utils.AppendVariable(0, "DeviceTime", "设备时间", "String", timeStr, ""))
            end
        end
        
    elseif frame.afn == AFN.DATA_FORWARD then
        -- 数据转发响应
        if frame.dataLen > 7 then
            local port = rxBuf[dataUnit.dataStart] or 0
            local dataLen = (rxBuf[dataUnit.dataStart + 2] or 0) * 256 + (rxBuf[dataUnit.dataStart + 1] or 0)
            
            table.insert(variables, utils.AppendVariable(0, "Port", "端口", "Number", port, ""))
            table.insert(variables, utils.AppendVariable(0, "DataLen", "数据长度", "Number", dataLen, ""))
            
            if dataLen > 0 and frame.dataLen >= dataLen + 7 then
                local responseData = {}
                for i = 1, math.min(dataLen, 50) do  -- 限制显示长度
                    responseData[i] = rxBuf[dataUnit.dataStart + 2 + i] or 0
                end
                local hexStr = utils.BytesToHex(responseData)
                table.insert(variables, utils.AppendVariable(0, "Data", "数据内容", "String", hexStr, ""))
            end
        end
        
    elseif frame.afn == AFN.REQ_CONFIG then
        -- 配置信息响应
        if dataUnit.fn == 1 then  -- F1: 终端版本信息
            if frame.dataLen > 4 then
                local dataStart = dataUnit.dataStart
                -- 尝试解析ASCII字符串
                local version = ""
                for i = 1, math.min(frame.dataLen - 4, 20) do
                    local char = rxBuf[dataStart + i - 1] or 0
                    if char >= 32 and char <= 126 then
                        version = version .. string.char(char)
                    elseif char > 0 then
                        version = version .. string.format("\\x%02X", char)
                    end
                end
                table.insert(variables, utils.AppendVariable(0, "TerminalVersion", "终端版本", "String", version, ""))
            end
        elseif dataUnit.fn == 256 then  -- F9: 通信模块版本信息
            if frame.dataLen > 4 then
                local dataStart = dataUnit.dataStart
                local imei = ""
                for i = 1, math.min(15, frame.dataLen - 4) do
                    local char = rxBuf[dataStart + i - 1] or 0
                    if char >= 32 and char <= 126 then
                        imei = imei .. string.char(char)
                    end
                end
                table.insert(variables, utils.AppendVariable(0, "IMEI", "IMEI号", "String", imei, ""))
            end
        end
        
    elseif frame.afn == AFN.CONFIRM_DENY then
        -- 确认/否认响应
        if dataUnit.fn == 1 then
            table.insert(variables, utils.AppendVariable(0, "Confirm", "确认状态", "String", "全部确认", ""))
        elseif dataUnit.fn == 2 then
            table.insert(variables, utils.AppendVariable(0, "Confirm", "确认状态", "String", "全部否认", ""))
        elseif dataUnit.fn == 4 then
            table.insert(variables, utils.AppendVariable(0, "Confirm", "确认状态", "String", "按数据单元确认/否认", ""))
        end
    end
    
    -- 添加基本帧信息
    table.insert(variables, utils.AppendVariable(100, "AFN", "应用功能码", "Number", frame.afn, string.format("0x%02X", frame.afn)))
    table.insert(variables, utils.AppendVariable(101, "SEQ", "帧序列", "Number", frame.seq % 16, ""))
    table.insert(variables, utils.AppendVariable(102, "Address", "终端地址", "Number", frame.address, ""))
    table.insert(variables, utils.AppendVariable(103, "Fn", "信息类", "Number", dataUnit.fn, ""))
    table.insert(variables, utils.AppendVariable(104, "Pn", "信息点", "Number", dataUnit.pn, ""))
    
    rxBuf = {}
    return {Status = "0", Variable = variables}
end

-- 插件信息函数---------------------------------------------------------------------------------------------------------------

-- 获取支持的命令列表
function GetSupportedCommands()
    return {
        {name = "login", desc = "登录命令"},
        {name = "heartbeat", desc = "心跳命令"},  
        {name = "logout", desc = "退出登录"},
        {name = "transparent", desc = "透明转发", params = "port,timeout,data"},
        {name = "version", desc = "读取终端版本"},
        {name = "module_version", desc = "读取模块版本"}
    }
end

-- 获取插件信息
function GetPluginInfo()
    return {
        name = "Q/GDW 1376.1",
        version = "1.0.0",
        author = "zdm",
        description = "Q/GDW 1376.1 电力通信协议插件",
        protocol = "Q3761-1376"
    }
end

-- 初始化函数
function Initialize()
    rxBuf = {}
    frameSeq = 0
    print("Q/GDW 1376.1 协议插件已初始化")
    return true
end

-- 清理函数
function Cleanup()
    rxBuf = {}
    print("Q/GDW 1376.1 协议插件已清理")
end

print("Q/GDW 1376.1 协议插件已加载") 