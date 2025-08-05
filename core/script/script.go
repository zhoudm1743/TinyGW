package script

import (
	"TinyGW/core/config"
	"errors"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

type Runner interface {
	Open(filename string) error
	Close()
	GenerateGetRealVariables(address string, step int) (data []byte, result bool, continued bool)
	DeviceCustomCmd(sAddr, cmdName, cmdParam string, step int) ([]byte, bool, bool)
	AnalysisRx(sAddr string, variables []any, rxBuf []byte, rxBufCnt int, tempVariables *[]any) bool
}

type poolRunner struct {
	pool   *sync.Pool
	logger *logrus.Logger
	cfg    *config.Config
	file   string
}

func NewRunner(cfg *config.Config, logger *logrus.Logger) Runner {
	return &poolRunner{
		pool: &sync.Pool{
			New: func() any {
				L := lua.NewState()
				registerGoFuncs(L)
				return L
			},
		},
		logger: logger,
		cfg:    cfg,
	}
}

func (p *poolRunner) Open(filename string) error {
	p.file = filename
	return nil
}

func (p *poolRunner) Close() {
	// 池化，不主动关闭 LState
}

func (p *poolRunner) withLState(fn func(L *lua.LState) error) error {
	L := p.pool.Get().(*lua.LState)
	defer p.pool.Put(L)
	L.Close() // 清理上次残留
	L = lua.NewState()
	registerGoFuncs(L)
	// 重新加载脚本
	path := filepath.Join("plugin", p.file, p.file+".lua")
	if err := L.DoFile(path); err != nil {
		p.logger.Errorf("加载Lua脚本失败: %v, path: %s", err, path)
		return err
	}
	return fn(L)
}

func (p *poolRunner) GenerateGetRealVariables(address string, step int) ([]byte, bool, bool) {
	var data []byte
	var result, continued bool
	err := p.withLState(func(L *lua.LState) error {
		if err := L.CallByParam(lua.P{
			Fn:      L.GetGlobal("GenerateGetRealVariables"),
			NRet:    1,
			Protect: true,
		}, lua.LString(address), lua.LNumber(step)); err != nil {
			p.logger.Errorf("GenerateGetRealVariables err %v", err)
			return err
		}
		ret := L.Get(-1)
		L.Pop(1)
		type Result struct {
			Status   string `json:"Status"`
			Variable []*byte
		}
		variables := Result{}
		if err := gluamapper.Map(ret.(*lua.LTable), &variables); err != nil {
			p.logger.Errorf("GenerateGetRealVariables gluamapper.Map err %v", err)
			return err
		}
		for _, v := range variables.Variable {
			data = append(data, *v)
		}
		result = true
		continued = variables.Status != "0"
		return nil
	})
	if err != nil {
		return nil, false, false
	}
	return data, result, continued
}

func (p *poolRunner) DeviceCustomCmd(address, commandName, commandParam string, step int) ([]byte, bool, bool) {
	var data []byte
	var result, continued bool
	err := p.withLState(func(L *lua.LState) error {
		if err := L.CallByParam(lua.P{
			Fn:      L.GetGlobal("DeviceCustomCmd"),
			NRet:    1,
			Protect: true,
		}, lua.LString(address), lua.LString(commandName), lua.LString(commandParam), lua.LNumber(step)); err != nil {
			p.logger.Errorf("DeviceCustomCmd err %v", err)
			return err
		}
		ret := L.Get(-1)
		L.Pop(1)
		type Result struct {
			Status   string  `json:"Status"`
			Variable []*byte `json:"Variable"`
		}
		resultObj := Result{}
		if err := gluamapper.Map(ret.(*lua.LTable), &resultObj); err != nil {
			p.logger.Errorf("DeviceCustomCmd gluamapper.Map err %v", err)
			return err
		}
		for _, v := range resultObj.Variable {
			data = append(data, *v)
		}
		result = true
		continued = resultObj.Status != "0"
		return nil
	})
	if err != nil {
		return nil, false, false
	}
	return data, result, continued
}

func (p *poolRunner) AnalysisRx(sAddr string, variables []any, rxBuf []byte, rxBufCnt int, tempVariables *[]any) bool {
	ok := false
	_ = p.withLState(func(L *lua.LState) error {
		table := L.NewTable()
		for _, v := range rxBuf {
			table.Append(lua.LNumber(v))
		}
		L.SetGlobal("rxBuf", table)
		if err := L.CallByParam(lua.P{
			Fn:      L.GetGlobal("AnalysisRx"),
			NRet:    1,
			Protect: true,
		}, lua.LString(sAddr), lua.LNumber(rxBufCnt)); err != nil {
			p.logger.Errorf("AnalysisRx err %v", err)
			return err
		}
		ret := L.Get(-1)
		if ret == nil {
			return errors.New("lua return nil")
		}
		L.Pop(1)
		// 这里只做演示，实际应按业务结构体映射
		ok = true
		return nil
	})
	return ok
}

// 注册Go函数到Lua
func registerGoFuncs(L *lua.LState) {
	// 设置 package.path 兼容 require
	L.DoString(`package.path = package.path .. ";./plugin/?.lua;./plugin/?/?.lua;"`)
	// TODO: 注册你的Go函数，如CRC等
}
