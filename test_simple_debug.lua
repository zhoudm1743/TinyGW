-- 简单调试脚本
package.path = "./plugin/Q3761-1376/?.lua;"

-- 加载真正的utils
dofile("./plugin/Q3761-1376/utils.lua")
dofile("./plugin/Q3761-1376/Q3761-1376.lua")

-- 测试数据包 (F254)
local testData = {
    0x68, 0xea, 0x01, 0xea, 0x01, 0x68, 0xca, 0x30, 0x05, 0x01, 0xa0, 0x01, 0x10, 0x73, 0x00, 0x00,
    0x20, 0x1f, 0x00, 0x15, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0xff, 0x00, 0x62, 0x00, 0xfd, 0x00,
    0x5e, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x01, 0x20, 0x25, 0x05, 0x30, 0x00, 0x01, 0x00, 0x02,
    0x25, 0x07, 0x01, 0x02, 0x13, 0x26, 0x00, 0x00, 0x03, 0x25, 0x07, 0x01, 0x13, 0x00, 0x00, 0x04,
    0x00, 0x00, 0x07, 0x15, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x14, 0x00, 0x00, 0x00, 0x15,
    0x21, 0x72, 0x00, 0x18, 0x00, 0x00, 0x37, 0x00, 0x1b, 0x00, 0x00, 0x46, 0x00, 0x1c, 0x00, 0x00,
    0x46, 0x00, 0x1f, 0x00, 0x00, 0x80, 0x00, 0x20, 0x00, 0x00, 0x80, 0x00, 0x23, 0x05, 0x75, 0x00,
    0x24, 0x05, 0x75, 0x00, 0x49, 0x00, 0x00, 0x00, 0x78, 0x00, 0x4e, 0x00, 0x00, 0x00, 0x00, 0x10,
    0xfa, 0x16
}

print("=== 调试F254解析问题 ===")

-- 测试utils.AppendVariable函数
local testVar = utils.AppendVariable(0, "TestName", "测试标签", "String", "测试值", "测试描述")
print("utils.AppendVariable返回结构:")
for k, v in pairs(testVar) do
    print("  " .. k .. ": " .. tostring(v))
end

-- 设置全局变量并调用解析
rxBuf = testData
local result = AnalysisRx("041549240", #testData)

print("\n=== 解析调试信息 ===")
print("状态: " .. result.Status)
print("变量数量: " .. #result.Variable)

-- 详细检查每个变量
for i = 1, math.min(3, #result.Variable) do
    local var = result.Variable[i]
    print(string.format("\n变量%d结构:", i))
    if type(var) == "table" then
        for k, v in pairs(var) do
            print(string.format("  %s: %s", k, tostring(v)))
        end
    else
        print("  不是表格: " .. tostring(var))
    end
end

-- 手动检查关键解析步骤
print("\n=== 手动验证关键步骤 ===")
local frame = parseFrame(testData, #testData)
if frame then
    print("✓ 帧解析成功，AFN: 0x" .. string.format("%02X", frame.afn))
    
    local dataUnit = parseDataUnitID(frame, testData)
    if dataUnit then
        local realFn = convertDTToFn(dataUnit.fn)
        print("✓ DT转换成功，realFn: " .. realFn)
        
        if frame.afn == AFN.DATA_FORWARD and realFn == 254 then
            print("✓ 进入F254解析分支")
            
            -- 手动检查F254数据格式
            local dataStart = dataUnit.dataStart
            local reserved = testData[dataStart]
            local protocol = testData[dataStart + 9]
            local dataLen = testData[dataStart + 10] + testData[dataStart + 11] * 256
            
            print(string.format("  预留: %d", reserved))
            print(string.format("  协议类型: %d", protocol))
            print(string.format("  数据长度: %d", dataLen))
            
            local fdPos = dataStart + 12
            local fdHeader = testData[fdPos]
            print(string.format("  FD头位置%d: 0x%02X", fdPos, fdHeader))
            
            if fdHeader == 0xFD then
                print("✓ 找到FD头标识")
            else
                print("✗ FD头标识错误")
            end
        else
            print("✗ 未进入F254解析分支")
            print("  frame.afn: " .. frame.afn .. " (期望: " .. AFN.DATA_FORWARD .. ")")
            print("  realFn: " .. realFn .. " (期望: 254)")
        end
    else
        print("✗ 数据单元解析失败")
    end
else
    print("✗ 帧解析失败")
end 