utils = {}

-- 添加变量到结果集
function utils.AddVariable(vindex, vname, vlabel, vtype)
    local Variable = {
        Index = vindex,
        Name  = vname,
        Label = vlabel,
        Type  = vtype,
    }
    return Variable
end

-- 添加带值的变量到结果集
function utils.AppendVariable(vindex, vname, vlabel, vtype, vValue, vExplain)
    local Variable = {
        Index = vindex,
        Name  = vname,
        Label = vlabel,
        Value = vValue,
        Explain = vExplain,
        Type  = vtype,
    }
    return Variable
end

-- 位与运算
function utils.And(num1, num2)
    local tmp1 = num1
    local tmp2 = num2
    local ret = 0
    local count = 0
    repeat
        local s1 = tmp1 % 2
        local s2 = tmp2 % 2
        if s1 == s2 and s1 == 1 then
            ret = ret + 2^count
        end
        tmp1 = math.modf(tmp1/2)
        tmp2 = math.modf(tmp2/2)
        count = count + 1
    until(tmp1 == 0 and tmp2 == 0)
    return ret
end

-- 位或运算
function utils.Or(num1, num2)
    local tmp1 = num1
    local tmp2 = num2
    local ret = 0
    local count = 0
    repeat
        local s1 = tmp1 % 2
        local s2 = tmp2 % 2
        if s1 == 1 or s2 == 1 then
            ret = ret + 2^count
        end
        tmp1 = math.modf(tmp1/2)
        tmp2 = math.modf(tmp2/2)
        count = count + 1
    until(tmp1 == 0 and tmp2 == 0)
    return ret
end

-- 位异或运算
function utils.Xor(num1, num2)
    local tmp1 = num1
    local tmp2 = num2
    local ret = 0
    local count = 0
    repeat
        local s1 = tmp1 % 2
        local s2 = tmp2 % 2
        if s1 ~= s2 then
            ret = ret + 2^count
        end
        tmp1 = math.modf(tmp1/2)
        tmp2 = math.modf(tmp2/2)
        count = count + 1
    until(tmp1 == 0 and tmp2 == 0)
    return ret
end

-- 字节数组转十六进制字符串
function utils.BytesToHex(bytes)
    local hexStr = ""
    for i = 1, #bytes do
        hexStr = hexStr .. string.format("%02X", bytes[i])
    end
    return hexStr
end

-- 十六进制字符串转字节数组
function utils.HexToBytes(hexStr)
    local bytes = {}
    for i = 1, #hexStr, 2 do
        local byteStr = hexStr:sub(i, i + 1)
        table.insert(bytes, tonumber(byteStr, 16))
    end
    return bytes
end

-- BCD码转十进制
function utils.BcdToDec(bcd)
    return math.floor(bcd / 16) * 10 + (bcd % 16)
end

-- 十进制转BCD码
function utils.DecToBcd(dec)
    return math.floor(dec / 10) * 16 + (dec % 10)
end

-- 小端字节序转数值
function utils.LittleEndianToInt(bytes, startIndex, length)
    local value = 0
    for i = 0, length - 1 do
        value = value + bytes[startIndex + i] * (256 ^ i)
    end
    return value
end

-- 数值转小端字节序
function utils.IntToLittleEndian(value, length)
    local bytes = {}
    for i = 1, length do
        bytes[i] = value % 256
        value = math.floor(value / 256)
    end
    return bytes
end

-- 大端字节序转数值
function utils.BigEndianToInt(bytes, startIndex, length)
    local value = 0
    for i = 0, length - 1 do
        value = value * 256 + bytes[startIndex + i]
    end
    return value
end

-- 数值转大端字节序
function utils.IntToBigEndian(value, length)
    local bytes = {}
    for i = length, 1, -1 do
        bytes[i] = value % 256
        value = math.floor(value / 256)
    end
    return bytes
end

-- 格式化时间字符串
function utils.FormatTime(timeBytes)
    if #timeBytes < 6 then
        return "无效时间"
    end
    
    local sec = utils.BcdToDec(timeBytes[1])
    local min = utils.BcdToDec(timeBytes[2])
    local hour = utils.BcdToDec(timeBytes[3])
    local day = utils.BcdToDec(timeBytes[4])
    local monthWeek = timeBytes[5]
    local month = utils.BcdToDec(monthWeek % 32)
    local week = math.floor(monthWeek / 32)
    local year = 2000 + utils.BcdToDec(timeBytes[6])
    
    local weekNames = {"", "一", "二", "三", "四", "五", "六", "日"}
    local weekStr = weekNames[week + 1] or ""
    
    return string.format("%04d-%02d-%02d %02d:%02d:%02d 周%s", 
                        year, month, day, hour, min, sec, weekStr)
end

-- 计算模256校验和
function utils.CheckSum(buffer, startIndex, endIndex)
    local sum = 0
    for i = startIndex, endIndex do
        sum = sum + (buffer[i] or 0)
    end
    return sum % 256
end

-- 调试打印函数
function utils.DebugPrint(title, data)
    if type(data) == "table" then
        print(title .. ": " .. utils.BytesToHex(data))
    else
        print(title .. ": " .. tostring(data))
    end
end

-- 深拷贝表
function utils.DeepCopy(orig)
    local orig_type = type(orig)
    local copy
    if orig_type == 'table' then
        copy = {}
        for orig_key, orig_value in next, orig, nil do
            copy[utils.DeepCopy(orig_key)] = utils.DeepCopy(orig_value)
        end
        setmetatable(copy, utils.DeepCopy(getmetatable(orig)))
    else
        copy = orig
    end
    return copy
end

-- 数组合并
function utils.MergeArrays(...)
    local result = {}
    local arrays = {...}
    for _, array in ipairs(arrays) do
        for _, value in ipairs(array) do
            table.insert(result, value)
        end
    end
    return result
end

return utils 