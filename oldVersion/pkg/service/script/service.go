package script

import (
	"go.uber.org/fx"
	"math"
	"path/filepath"
	"tinyGW/app/models"
	"tinyGW/pkg/plugin/io"

	"github.com/gookit/color"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
	luar "layeh.com/gopher-luar"
)

type (
	Runner interface {
		Open(filename string) error
		Close()
		GenerateGetRealVariables(address string, step int) (data []byte, result bool, continued bool)
		DeviceCustomCmd(sAddr string, cmdName string, cmdParam string, step int) ([]byte, bool, bool)
		AnalysisRx(sAddr string, variables []models.DeviceProperty, rxBuf []byte, rxBufCnt int, tempVariables *[]models.DeviceProperty) bool
	}

	luaRunner struct {
		lState *lua.LState
	}
)

func NewLuaRunner() Runner {
	return &luaRunner{
		lState: nil,
	}
}

func (l *luaRunner) Open(filename string) error {
	l.lState = lua.NewState()

	path := filepath.Join(io.GetCurrentPath(), "plugin", filename, filename+".lua")
	color.Blueln("lua script file: ", path) // lua script file:  D:\gitlab\energy\zsxagw/plugin//.lua
	if err := l.lState.DoFile(path); err != nil {
		zap.S().Errorf("Failed to load Lua script: %v, path: %s", err, path)
		return err
	}

	l.lState.SetGlobal("GetCRCModbus", l.lState.NewFunction(GetCRCModbus))
	l.lState.SetGlobal("CheckCRCModbus", l.lState.NewFunction(CheckCRCModbus))
	l.lState.SetGlobal("GetCRCModbusLittleEndian", l.lState.NewFunction(GetCRCModbusLittleEndian))
	return nil
}

func (l *luaRunner) Close() {
	l.lState.Close()
}

func (l *luaRunner) GenerateGetRealVariables(address string, step int) ([]byte, bool, bool) {
	if l.lState == nil {
		return nil, false, false
	}

	// 调用GenerateGetRealVariables
	if err := l.lState.CallByParam(lua.P{
		Fn:      l.lState.GetGlobal("GenerateGetRealVariables"),
		NRet:    1,
		Protect: true,
	},
		lua.LString(address),
		lua.LNumber(step)); err != nil {
		zap.S().Errorf("GenerateGetRealVariables err %v", err)
		return nil, false, false
	}

	// 获取返回结果
	ret := l.lState.Get(-1)
	l.lState.Pop(1)

	type Result struct {
		Status   string `json:"Status"`
		Variable []*byte
	}
	variables := Result{}
	if err := gluamapper.Map(ret.(*lua.LTable), &variables); err != nil {
		zap.S().Errorf("GenerateGetRealVariables gluamapper.Map err %v", err)
		return nil, false, false
	}

	data := make([]byte, 0)
	for _, v := range variables.Variable {
		data = append(data, *v)
	}

	return data, true, variables.Status != "0"
}

func (l *luaRunner) DeviceCustomCmd(address string, commandName string, commandParam string, step int) ([]byte, bool, bool) {
	if l.lState == nil {
		return nil, false, false
	}

	// 调用DeviceCustomCmd
	if err := l.lState.CallByParam(lua.P{
		Fn:      l.lState.GetGlobal("DeviceCustomCmd"),
		NRet:    1,
		Protect: true,
	}, lua.LString(address),
		lua.LString(commandName),
		lua.LString(commandParam),
		lua.LNumber(step)); err != nil {
		zap.S().Errorf("DeviceCustomCmd err %v", err)
		return nil, false, false
	}

	// 获取返回结果
	ret := l.lState.Get(-1)
	l.lState.Pop(1)

	type Result struct {
		Status   string  `json:"Status"`
		Variable []*byte `json:"Variable"`
	}
	result := Result{}
	if err := gluamapper.Map(ret.(*lua.LTable), &result); err != nil {
		zap.S().Errorf("DeviceCustomCmd gluamapper.Map err %v", err)
		return nil, false, false
	}

	data := make([]byte, 0)
	for _, v := range result.Variable {
		data = append(data, *v)
	}
	return data, true, result.Status != "0"
}

func (l *luaRunner) AnalysisRx(sAddr string, variables []models.DeviceProperty, rxBuf []byte, rxBufCnt int, tempVariables *[]models.DeviceProperty) bool {
	if l.lState == nil {
		return false
	}

	table := lua.LTable{}
	for _, v := range rxBuf {
		table.Append(lua.LNumber(v))
	}
	l.lState.SetGlobal("rxBuf", luar.New(l.lState, &table))

	// AnalysisRx
	err := l.lState.CallByParam(lua.P{
		Fn:      l.lState.GetGlobal("AnalysisRx"),
		NRet:    1,
		Protect: true,
	}, lua.LString(sAddr), lua.LNumber(rxBufCnt))
	if err != nil {
		zap.S().Warnf("AnalysisRx err %v", err)
		return false
	}

	// 获取返回结果
	ret := l.lState.Get(-1)
	if ret == nil {
		return false
	}
	l.lState.Pop(1)

	type LuaVariableTemplate struct {
		Index   int
		Name    string
		Label   string
		Type    string
		Value   interface{}
		Explain string
	}

	type LuaVariableMapTemplate struct {
		Status   string `json:"Status"`
		Variable []*LuaVariableTemplate
	}
	LuaVariableMap := LuaVariableMapTemplate{}

	if err := gluamapper.Map(ret.(*lua.LTable), &LuaVariableMap); err != nil {
		zap.S().Warnf("AnalysisRx gluamapper.Map err %v", err)
		return false
	}

	if LuaVariableMap.Status != "0" {
		return false
	}

	for _, lv := range LuaVariableMap.Variable {
		for k, p := range variables {
			if lv.Name == p.Name {
				if lv.Value == nil {
					if p.Type == "string" {
						lv.Value = ""
					} else {
						lv.Value = float64(0)
					}
				}
				if p.Scale > 1 {
					lv.Value = lv.Value.(float64) * p.Scale
				}
				switch p.Type {
				case "int":
					variables[k].Value = (int32)(lv.Value.(float64))
				case "uint":
					variables[k].Value = (uint32)(lv.Value.(float64))
				case "double":
					variables[k].Value = lv.Value.(float64) / math.Pow10(p.Decimal)
				case "string":
					variables[k].Value = lv.Value.(string)
				}
				// 添加到临时变量表中
				*tempVariables = append(*tempVariables, variables[k])
			}
		}
	}

	return true
}

var Module = fx.Provide(
	NewLuaRunner,
)
