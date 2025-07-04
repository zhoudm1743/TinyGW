-- 测试数据分析脚本
package.path = "./plugin/Q3761-1376/?.lua;"
require "utils"
require "json"

-- 模拟接收到的数据
local testData = {
    0x68, 0x8E, 0x00, 0x8E, 0x00, 0x68, 0x84, 0x30, 0x05, 0x01, 0xA0, 0xDE, 0x10, 0x60, 0x00, 0x00, 
    0x01, 0x00, 0x01, 0x14, 0x00, 0x68, 0x01, 0x00, 0x30, 0x05, 0x25, 0x20, 0x68, 0x91, 0x08, 0x33, 
    0x33, 0x34, 0x33, 0x53, 0x3A, 0x33, 0x33, 0xA4, 0x16, 0x1C, 0x16
}

-- 设置全局rxBuf
rxBuf = testData

-- 手动分析数据
print("=== Manual Analysis ===")
print("Data length: " .. #testData)

-- 查找帧头
local startIndex = 1
for i = 1, #testData - 15 do
    if testData[i] == 0x68 and i + 5 <= #testData and testData[i + 5] == 0x68 then
        startIndex = i
        break
    end
end

print("Frame start: " .. startIndex)

-- 提取长度
local len1 = testData[startIndex + 1] or 0
local len2 = testData[startIndex + 2] or 0
local frameLen = math.floor(((len2 * 256 + len1) - 2) / 4)
print("User data length: " .. frameLen)

-- 控制域
local ctrl = testData[startIndex + 6] or 0
print("Control: 0x" .. string.format("%02X", ctrl))

-- AFN
local afn = testData[startIndex + 12] or 0
print("AFN: 0x" .. string.format("%02X", afn))

-- SEQ
local seq = testData[startIndex + 13] or 0
print("SEQ: 0x" .. string.format("%02X", seq))

-- 数据单元标识
local da1 = testData[startIndex + 14] or 0
local da2 = testData[startIndex + 15] or 0
local dt1 = testData[startIndex + 16] or 0
local dt2 = testData[startIndex + 17] or 0

local pn = da2 * 256 + da1
local fn = dt2 * 256 + dt1
print("Pn: " .. pn .. ", Fn: 0x" .. string.format("%04X", fn))

-- 数据内容
local dataStart = startIndex + 18
print("Data start: " .. dataStart)

-- 分析DLT645数据
if afn == 0x10 and fn == 0x0001 then
    print("=== Analyze DLT645 transparent forward data ===")
    
    -- 跳过透明转发头部 (6字节)
    local dltStart = dataStart + 6
    print("DLT645 data start: " .. dltStart)
    
    -- 获取透明转发内容长度（2字节，小端序）
    local contentLen = (testData[dltStart + 1] or 0) * 256 + (testData[dltStart] or 0)
    print("Transparent forward content length: " .. contentLen)
    
    -- 透明转发内容起始位置（跳过长度字段2字节）
    local contentStart = dltStart + 2
    print("Transparent forward content start: " .. contentStart)
    
    -- 打印透明转发内容
    local contentHex = ""
    for i = contentStart, math.min(contentStart + contentLen - 1, #testData) do
        contentHex = contentHex .. string.format("%02X ", testData[i] or 0)
    end
    print("Transparent forward content: " .. contentHex)
    
    -- 在透明转发内容中查找DLT645帧
    local dltStartIndex = -1
    for i = contentStart, contentStart + contentLen - 8 do
        if testData[i] == 0x68 and testData[i+7] == 0x68 then
            dltStartIndex = i
            print("Found DLT645 frame at transparent content position " .. (i - contentStart))
            break
        end
    end
    
    if dltStartIndex > 0 then
        print("Found DLT645 frame at: " .. dltStartIndex)
        
        -- 打印DLT645帧的前几个字节
        local dltHex = ""
        for i = dltStartIndex, math.min(dltStartIndex + 15, #testData) do
            dltHex = dltHex .. string.format("%02X ", testData[i] or 0)
        end
        print("DLT645 frame data: " .. dltHex)
        
        -- DLT645地址
        local addr = ""
        for i = dltStartIndex + 1, dltStartIndex + 6 do
            addr = addr .. string.format("%02X", testData[i] or 0)
        end
        print("DLT645 address: " .. addr)
        
        -- 检查第二个68
        local second68Pos = dltStartIndex + 7
        print("Check second 68 at position " .. second68Pos .. ", value 0x" .. string.format("%02X", testData[second68Pos] or 0))
        
        if testData[second68Pos] == 0x68 then
            -- 控制码
            local ctrlCode = testData[dltStartIndex + 8] or 0
            print("DLT645 control code: 0x" .. string.format("%02X", ctrlCode))
            
            -- 数据长度
            local dltDataLen = testData[dltStartIndex + 9] or 0
            print("DLT645 data length: " .. dltDataLen)
            
            if ctrlCode == 0x91 then
                print("=== Parse read data response ===")
                
                -- 数据标识
                local dataIdStart = dltStartIndex + 10
                local dataIdStr = string.format("%02X%02X%02X%02X",
                        (testData[dataIdStart] or 0) - 0x33, 
                        (testData[dataIdStart + 1] or 0) - 0x33,
                        (testData[dataIdStart + 2] or 0) - 0x33, 
                        (testData[dataIdStart + 3] or 0) - 0x33)
                print("Data ID: " .. dataIdStr)
                
                -- 数据值
                local valueStart = dataIdStart + 4
                local valueData = {}
                local valueEnd = valueStart + dltDataLen - 5
                print("Data value start: " .. valueStart .. ", end: " .. valueEnd)
                
                for i = valueStart, valueEnd do
                    if i <= #testData then
                        local rawValue = testData[i] or 0
                        local decodedValue = rawValue - 0x33
                        table.insert(valueData, decodedValue)
                        print("Data[" .. i .. "] raw=0x" .. string.format("%02X", rawValue) .. ", decoded=" .. decodedValue)
                    else
                        print("Data index " .. i .. " out of range")
                        break
                    end
                end
                
                print("Parsed " .. #valueData .. " bytes of data")
                
                if dataIdStr == "00000100" then
                    -- 正向总有功电能
                    local n = ""
                    for i = #valueData, 1, -1 do
                        n = n .. string.format("%02X", valueData[i])
                    end
                    
                    print("Hex string: " .. n)
                    local consumption = tonumber(n, 10) or 0
                    local actualConsumption = consumption / 100.0
                    print("Forward total active energy: " .. actualConsumption .. " kWh")
                    print("Decimal value: " .. consumption)
                end
            end
        else
            print("DLT645 frame format error, second 68 not found")
            -- 查找其他可能的68位置
            for i = dltStartIndex + 1, math.min(dltStartIndex + 10, #testData) do
                if testData[i] == 0x68 then
                    print("Found 0x68 at position " .. i)
                end
            end
        end
    else
        print("No valid DLT645 frame found")
        -- 打印当前位置附近的数据
        local debugHex = ""
        for i = contentStart, math.min(contentStart + 10, #testData) do
            debugHex = debugHex .. string.format("%02X ", testData[i] or 0)
        end
        print("Data near current position: " .. debugHex)
    end
end

print("=== Analysis complete ===") 