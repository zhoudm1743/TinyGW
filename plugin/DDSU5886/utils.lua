utils = {}

function utils.AddVariable(vindex,vname,vlable,vtype)
    local Variable = {
        Index = vindex,
        Name  = vname,
        Label = vlable,
        --Value = "",
        Type  = vtype,
    }
    --print(variable)
    return Variable
end

function utils.AppendVariable(vindex,vname,vlable,vtype,vValue,vExplain)
    local Variable = {
        Index = vindex,
        Name  = vname,
        Label = vlable,
        Value = vValue,
        Explain = vExplain,
        Type  = vtype,
    }
    --print(variable)
    return Variable
end

function utils.And(num1,num2)
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

--备注：math.modf() 这个函数就是为了获取除法的整数部分
function utils.hexToFloat( hexString )
    if hexString == nil then
        return 0
    end
    local t = type( hexString )
    if t == "string" then
        hexString = tonumber(hexString , 16)
    end

    local hexNums = hexString

    local sign = math.modf(hexNums/(2^31))

    local exponent = hexNums % (2^31)
    exponent = math.modf(exponent/(2^23)) -127

    local mantissa = hexNums % (2^23)

    for i=1,23 do
        mantissa = mantissa / 2
    end
    mantissa = 1+mantissa
    --	print(mantissa)
    local result = (-1)^sign * mantissa * 2^exponent
    return result
end

function utils.FloatToHex( floatNum )
    local S = 0
    local E = 0
    local M = 0
    if floatNum == 0 then
        return "00000000"
    end
    local num1,num2 = math.modf(floatNum/1)
    local InterPart = num1

    if floatNum > 0 then
        S = 0 * 2^31
    else
        S = 1 * 2^31
    end
    local intercount = 0
    repeat
        num1 = math.modf(num1/2)
        intercount = intercount + 1
    until (num1 == 0)

    E = intercount - 1
    local Elen = 23 - E
    InterPart = InterPart % (2^E)
    InterPart = InterPart * (2^Elen)

    E = E + 127
    E = E * 2^23

    for i=1,Elen do
        num2 = num2 * 2
        num1,num2 = math.modf(num2/1)
        M = M + num1 * 2^(Elen - i)
    end

    M = InterPart + M

    --E值为整数部分转成二进制数后左移位数：22.8125 转成二进制10110.1101，左移4位 1.01101101
    --E=4 ，再加上127 就为所需E值
    --010000011 01101101 000000000000000

    local Result = S + E + M

    Result = string.format("%08X",Result)
    return Result
end
return utils