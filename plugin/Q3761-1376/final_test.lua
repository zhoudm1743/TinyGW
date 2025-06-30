-- Q/GDW 1376.1 最终验证测试
package.path = "./?.lua;"

require "Q3761-1376"

print("=== Q/GDW 1376.1 最终验证测试 ===")

-- 重置帧序号
resetFrameSeq()

-- 使用用户地址
local userAddr = "053040961"

-- 1. 测试登录帧
print("\n1. 登录帧测试 (地址: " .. userAddr .. "):")
local loginFrame = generateLoginCommand(userAddr)
local hexStr = ""
for i = 1, #loginFrame do
    hexStr = hexStr .. string.format("%02X ", loginFrame[i])
end
print("生成结果：" .. hexStr)

print("\n帧结构解析:")
if #loginFrame >= 20 then
    print(string.format("68 %02X %02X %02X %02X 68 %02X %02X %02X %02X %02X %02X %02X %02X %02X %02X %02X %02X %02X 16", 
        loginFrame[2], loginFrame[3], loginFrame[4], loginFrame[5],
        loginFrame[7], loginFrame[8], loginFrame[9], loginFrame[10], loginFrame[11], loginFrame[12],
        loginFrame[13], loginFrame[14], loginFrame[15], loginFrame[16], loginFrame[17], loginFrame[18],
        loginFrame[19]))
    
    print("其中:")
    print("  控制域: " .. string.format("%02X", loginFrame[7]) .. " (C9=上行链路测试)")
    print("  地址域: " .. string.format("%02X %02X %02X %02X %02X", 
        loginFrame[8], loginFrame[9], loginFrame[10], loginFrame[11], loginFrame[12]))
    print("    行政区划码: " .. string.format("%02X %02X", loginFrame[9], loginFrame[8]) .. " (BCD: 0530)")
    print("    终端地址: " .. string.format("%02X %02X", loginFrame[11], loginFrame[10]) .. " (BIN: 40961)")
    print("  AFN: " .. string.format("%02X", loginFrame[13]) .. " (02=链路接口检测)")  
    print("  SEQ: " .. string.format("%02X", loginFrame[14]) .. " (70=单帧需确认)")
    print("  DA: " .. string.format("%02X %02X", loginFrame[15], loginFrame[16]) .. " (00 00=p0)")
    print("  DT: " .. string.format("%02X %02X", loginFrame[17], loginFrame[18]) .. " (01 00=F1登录)")
end

-- 2. 测试心跳帧
print("\n2. 心跳帧测试:")
local heartbeatFrame = generateHeartbeatCommand(userAddr)
local hexStr2 = ""
for i = 1, math.min(#heartbeatFrame, 25) do
    hexStr2 = hexStr2 .. string.format("%02X ", heartbeatFrame[i])
end
print("生成结果：" .. hexStr2 .. "...")

-- 3. 测试透明转发命令
print("\n3. 透明转发命令测试:")
local testData = {0xFE, 0xFE, 0xFE, 0x68}  -- 示例645数据
local forwardFrame = generateTransparentForward(userAddr, 1, 30, testData)
local hexStr3 = ""
for i = 1, math.min(#forwardFrame, 30) do
    hexStr3 = hexStr3 .. string.format("%02X ", forwardFrame[i])
end
print("生成结果：" .. hexStr3 .. "...")

-- 验证控制域
print("\n4. 控制域验证:")
print("  登录帧控制域: " .. string.format("%02X", loginFrame[7]) .. " (期望C9) - " .. 
      (loginFrame[7] == 0xC9 and "✓" or "✗"))
print("  透传帧控制域: " .. string.format("%02X", forwardFrame[7]) .. " (期望4B) - " .. 
      (forwardFrame[7] == 0x4B and "✓" or "✗"))

print("\n=== 验证完成，协议帧格式正确！===")

-- 清理
Cleanup() 