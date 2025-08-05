-- bit.lua - 位操作函数库
-- 实现基本的位操作函数，用于Lua 5.1及以上版本

local bit = {}

-- 按位与操作
function bit.band(a, b)
    local result = 0
    local bitval = 1
    while a > 0 and b > 0 do
        if a % 2 == 1 and b % 2 == 1 then
            result = result + bitval
        end
        bitval = bitval * 2
        a = math.floor(a / 2)
        b = math.floor(b / 2)
    end
    return result
end

-- 按位或操作
function bit.bor(a, b)
    local result = 0
    local bitval = 1
    while a > 0 or b > 0 do
        if a % 2 == 1 or b % 2 == 1 then
            result = result + bitval
        end
        bitval = bitval * 2
        a = math.floor(a / 2)
        b = math.floor(b / 2)
    end
    return result
end

-- 按位异或操作
function bit.bxor(a, b)
    local result = 0
    local bitval = 1
    while a > 0 or b > 0 do
        if (a % 2) ~= (b % 2) then
            result = result + bitval
        end
        bitval = bitval * 2
        a = math.floor(a / 2)
        b = math.floor(b / 2)
    end
    return result
end

-- 按位非操作 (限制为32位)
function bit.bnot(n)
    return 4294967295 - n  -- 2^32 - 1 - n
end

-- 左移位操作
function bit.lshift(a, b)
    return a * (2 ^ b)
end

-- 右移位操作
function bit.rshift(a, b)
    return math.floor(a / (2 ^ b))
end

return bit 