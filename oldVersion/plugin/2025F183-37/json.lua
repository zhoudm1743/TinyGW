-- 简单json模块实现，用于测试目的
local json = {}

-- 简单的json解析实现
function json.jsondecode(jsonStr)
    -- 空字符串或nil返回nil
    if not jsonStr or jsonStr == "" then
        return nil
    end
    
    -- 空对象返回空表
    if jsonStr == "{}" then
        return {}
    end
    
    -- 简单的键值对解析
    local result = {}
    
    -- 先去除{}
    jsonStr = jsonStr:match("{(.*)}")
    if not jsonStr then 
        return result
    end
    
    -- 以逗号分隔键值对
    for pair in jsonStr:gmatch('[^,]+') do
        local key, value = pair:match('"([^"]+)"%s*:%s*(.+)')
        if key and value then
            -- 根据值的类型转换
            if value:match('^".*"$') then
                -- 字符串
                result[key] = value:match('"(.*)"')
            elseif value == "true" then
                result[key] = true
            elseif value == "false" then
                result[key] = false
            elseif value == "null" then
                result[key] = nil
            else
                -- 尝试转换为数字
                local num = tonumber(value)
                if num then
                    result[key] = num
                else
                    -- 默认为字符串
                    result[key] = value
                end
            end
        end
    end
    
    return result
end

-- 简单的json编码实现
function json.jsonencode(value)
    local t = type(value)
    if t == "table" then
        local result = "{"
        local first = true
        for k, v in pairs(value) do
            if not first then result = result .. "," else first = false end
            result = result .. '"' .. tostring(k) .. '":'
            
            local vtype = type(v)
            if vtype == "string" then
                result = result .. '"' .. v .. '"'
            elseif vtype == "number" or vtype == "boolean" then
                result = result .. tostring(v)
            elseif vtype == "table" then
                result = result .. json.jsonencode(v)
            else
                result = result .. '"' .. tostring(v) .. '"'
            end
        end
        result = result .. "}"
        return result
    elseif t == "string" then
        return '"' .. value .. '"'
    else
        return tostring(value)
    end
end

return json