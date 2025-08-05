package.path = "./plugin/DDSY1829/?.lua;"

require "utils"
require "json"

-- 公共函数---------------------------------------------------------------------------------------------------------------
-- 16进制文本转10进制
function HexToDec(hex)
    local dec = 0
    local len = string.len(hex)
    for i = 1, len do
        dec = dec + tonumber(string.sub(hex, len - i + 1, len - i + 1), 16) * 2 ^ ((i - 1) * 4)
    end
    return dec
end
-- 校验和
function checkSum(buffer, startIndex, endIndex)
    local sum = 0

    for i = startIndex, endIndex do
        sum = sum + buffer[i]
    end

    return sum % 256
end

-- 地址转换 2 ~ 7 字节
function convertAddress(requestADU, sAddr)
    local addr = string.format("%012d", tonumber(sAddr))
    requestADU[2] = tonumber(string.sub(addr, 11, 12), 16)
    requestADU[3] = tonumber(string.sub(addr, 9, 10), 16)
    requestADU[4] = tonumber(string.sub(addr, 7, 8), 16)
    requestADU[5] = tonumber(string.sub(addr, 5, 6), 16)
    requestADU[6] = tonumber(string.sub(addr, 3, 4), 16)
    requestADU[7] = tonumber(string.sub(addr, 1, 2), 16)
end

-- 加密
function encode(source, startIndex, endIndex)
    local y = 13
    for i = startIndex, endIndex do
        local _, dec = math.modf(y / 2)
        if (dec == 0) then
            source[i] = (source[i] + 0x33 + 0x48 + y) % 256
        else
            source[i] = (source[i] + 0x33 + 0x54 + y) % 256
        end
        y = y + 1
    end
end

-- 解码
function decode(source, startIndex, endIndex)
    local result = {}

    local y = 1
    for i = startIndex, endIndex do
        local _, dec = math.modf(y / 2)
        if (dec == 0) then
            result[y] = source[i] - 0x33 - 0x54 - y + 1
        else
            result[y] = source[i] - 0x33 - 0x48 - y + 1
        end

        if (result[y] < 0) then
            result[y] = result[y] + 256
        end
        y = y + 1
    end

    return result
end

-- GenerateCommand
function GenerateCommand(sAddr,cmd)
    --                         |<-------------地址----------->|                                             checkSum
    --                   1     2     3     4     5     6     7     8     9     10    11    12    13    14    15    16
    local requestADU = { 0x68, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x68, 0x11, 0x04, 0x33, 0x33, 0x34, 0x33, 0x00, 0x16 }
    -- [FE FE FE FE 68 41 65 15 22 00 00 68 11 04 33 33 34 33 8F 16]
    -- 地址转换 2 ~ 7 字节
    convertAddress(requestADU, sAddr)
    -- cmd 11 ~ 14 字节
    requestADU[11] = cmd[1] + 0x33
    requestADU[12] = cmd[2] + 0x33
    requestADU[13] = cmd[3] + 0x33
    requestADU[14] = cmd[4] + 0x33
    -- 校验和
    requestADU[15] = checkSum(requestADU, 1, 14)
    -- 在requestADU前面插入 4个 0xFE
    for i = 1, 4 do
        table.insert(requestADU, 1, 0xFE)
    end
    return requestADU
end
------------------------------------------------------------------------------------------------------------------------
-- 生成获取(当前)正向有功总电能的指令
function GenerateGetConsumption(sAddr, continued)
    local requestADU = GenerateCommand(sAddr, { 0x00, 0x00, 0x01, 0x00 })
    return { Status = continued, Variable = requestADU }
end

function GenerateGetAmount(sAddr, continued)
    local requestADU = GenerateCommand(sAddr, { 0x00, 0x01, 0x90, 0x00 })
    return { Status = continued, Variable = requestADU }
end

-- 瞬时总有功功率
function GenerateGetPower(sAddr, continued)
    local requestADU = GenerateCommand(sAddr, { 0x00, 0x00, 0x03, 0x02 })
    return { Status = continued, Variable = requestADU }
end

function DeviceCustomCmd(sAddr, cmdName, cmdParam, step)
    local params = json.jsondecode(cmdParam)
    if (cmdName == "dev_consumption") then
        return GenerateGetConsumption(sAddr)
    end
    return { Status = "0", Variable = {} }
end

rxBuf = {}

function AnalysisRx(sAddr, rxBufCnt)
    -- 最短指令12个字节
    if (rxBufCnt < 16) then
        rxBuf = {}
        -- Status = "1" 错误
        print("error:rxBufCnt < 16")
        return { Status = "1", Variable = {} }
    end
    local startIndex = 1
    while rxBuf[startIndex] ~= 0x68 do
        startIndex = startIndex + 1
        if startIndex > #rxBuf then
            error("0x68起始符未找到")
            return
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
            print("i:".. i, string.format("%X", rxBuf[i-1]), string.format("%X", rxBuf[i]))
            local cs = checkSum(rxBuf, 1, i - 2)
            if (cs == rxBuf[i-1]) then
                endFlag = rxBuf[i]
                csFlag = rxBuf[i-1]
                endIndex = i

                break
            end
        end
    end
    print("endFlag:".. string.format("%X", endFlag), endIndex)
    print("csFlag:".. string.format("%X", csFlag))
    if (endFlag ~= 0x16) then
        rxBuf = {}
        -- Status = "1" 错误
        print("error:endFlag ~= 0x16")
        return { Status = "1", Variable = {} }
    end

    local index = 1
    -- 地址域 6字节
    local mAddr = string.format("%02X%02X%02X%02X%02X%02X",
                    rxBuf[index + 6], rxBuf[index + 5], rxBuf[index + 4], rxBuf[index + 3], rxBuf[index + 2], rxBuf[index + 1])
    -- 如果地址域前后不是0x68 则错误
    if (rxBuf[index] ~= 0x68 or rxBuf[index + 7] ~= 0x68) then
        rxBuf = {}
        -- Status = "1" 错误
        print("error:rxBuf[index] ~= 0x68 or rxBuf[index + 7] ~= 0x68")
        return { Status = "1", Variable = {} }
    end
    local addr = string.format("%012d", tonumber(sAddr))
    if (mAddr ~= addr) then
        rxBuf = {}
        print("error:mAddr ~= sAddr")
        return { Status = "1", Variable = {} }
    end

    index = index + 8
    -- 91
    local Cflag = index
    -- 控制码
    local flag = rxBuf[Cflag]
    -- 数据长度
    local dataLen = rxBuf[Cflag+1]

    -- 数据标识
    local dfStr = string.format("%02X%02X%02X%02X",
                    rxBuf[Cflag+2]-0x33, rxBuf[Cflag+3]-0x33, rxBuf[Cflag+4]-0x33, rxBuf[Cflag+5]-0x33)
    dataIndex = Cflag + 6
    -- 数据域 从dataIndex开始,到 cs 前
    local data = {}
    for i = dataIndex, endIndex - 2 do
        table.insert(data, rxBuf[i]-0x33)
    end
    if dfStr == "00000100" then
        local n = ""
        for i = #data, 1, -1 do
            n = n.. string.format("%02X", data[i])
        end

        dev_consumption = tonumber(n, 10)
        print("dev_consumption:".. dev_consumption/100.0)
        rxBuf = {}
        return { Status = "0", Variable = {
            utils.AppendVariable(0, "dev_consumption", "总电能", "double", dev_consumption,
                string.format("%.2f", dev_consumption / 100.0))
            }
        }
    elseif dfStr == "00019000" then
        local n = ""
        for i = #data, 1, -1 do
            n = n.. string.format("%02X", data[i])
        end
        dev_amount = tonumber(n, 10)
        rxBuf = {}
        return { Status = "0", Variable = {
            utils.AppendVariable(1, "dev_remain_amt", "剩余电量", "double", dev_amount,
                string.format("%.2f", dev_amount / 100.0))
            }
        }
    elseif dfStr == "01019000" then
       local n = ""
        for i = #data, 1, -1 do
            n = n.. string.format("%02X", data[i])
        end
        dev_remain_amt = tonumber(n, 10)
        rxBuf = {}
        return { Status = "0", Variable = {
            utils.AppendVariable(2, "dev_remain_amt", "剩余电量", "double", dev_remain_amt,
                string.format("%.2f", dev_remain_amt / 100.0))
            }
        }
    elseif dfStr == "00000302" then
        for i = #data, 1, -1 do
            n = n.. string.format("%02X", data[i])
        end
        dev_power = tonumber(n, 10)
        print("dev_power:".. dev_power/10000.0)
        rxBuf = {}
        return { Status = "0", Variable = {
            utils.AppendVariable(3, "dev_power", "剩余电量", "double", dev_power,
                    string.format("%.2f", dev_power / 10000.0))
        }
        }
    end
    rxBuf = {}
    return { Status = "1", Variable = {} }
end

function GenerateGetRealVariables(sAddr, step)
    print("DDSY1829", sAddr, step)
    if (step == 0) then
        return GenerateGetConsumption(sAddr, "0")
    end
end
