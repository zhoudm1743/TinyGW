
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

return utils