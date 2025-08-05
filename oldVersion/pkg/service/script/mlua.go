package script

import (
	"sync"

	"go.uber.org/zap"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

type crc struct {
	once  sync.Once
	table []uint16
}

var crcTb crc

// initTable 初始化表
func (c *crc) initTable() {
	crcPoly16 := uint16(0xa001)
	c.table = make([]uint16, 256)

	for i := uint16(0); i < 256; i++ {
		crc := uint16(0)
		b := i

		for j := uint16(0); j < 8; j++ {
			if ((crc ^ b) & 0x0001) > 0 {
				crc = (crc >> 1) ^ crcPoly16
			} else {
				crc = crc >> 1
			}
			b = b >> 1
		}
		c.table[i] = crc
	}
}

func crc16(bs []byte) uint16 {
	crcTb.once.Do(crcTb.initTable)

	val := uint16(0xFFFF)
	for _, v := range bs {
		val = (val >> 8) ^ crcTb.table[(val^uint16(v))&0x00FF]
	}
	return val
}

func LuaCallNewVariables(L *lua.LState) {
	// 调用NewVariables
	err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("NewVariables"),
		NRet:    1,
		Protect: true,
	})
	if err != nil {
		panic(any(err))
	}
	// 获取返回结果
	ret := L.Get(-1)
	L.Pop(1)
	switch ret.(type) {
	case lua.LString:
		zap.S().Info("string")
	case *lua.LTable:
		zap.S().Info("table")
	}

	type VariableTemplate struct {
		Index int
		Name  string
		Label string
		Type  string
	}

	type VariableMapTemplate struct {
		Variable []*VariableTemplate
	}

	VariableMap := VariableMapTemplate{}

	if err := gluamapper.Map(ret.(*lua.LTable), &VariableMap); err != nil {
		panic(any(err))
	}

	for _, v := range VariableMap.Variable {
		zap.S().Infof("%+v", v.Label)
	}
}

//func LuaInit() {
//
//	exeCurDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
//
//	path := exeCurDir + "/plugin/"
//
//	L := lua.NewState()
//	defer L.Close()
//	//加载Lua
//	if err := L.DoFile(path + "td200.lua"); err != nil {
//		logger.Zap.Warnf("open td200.lua fail %v", err)
//	}
//	logger.Zap.Info("open TD200.lua OK")
//
//	LuaCallNewVariables(L)
//}

func GetCRCModbus(L *lua.LState) int {
	type LuaVariableMapTemplate struct {
		Variable []*byte
	}

	lv := L.ToTable(1)

	LuaVariableMap := LuaVariableMapTemplate{}
	if err := gluamapper.Map(lv, &LuaVariableMap); err != nil {
		zap.S().Warnf("GetCRC16 gluamapper.Map err %v", err)
	}

	nBytes := make([]byte, 0)
	if len(LuaVariableMap.Variable) > 0 {
		for _, v := range LuaVariableMap.Variable {
			nBytes = append(nBytes, *v)
		}
	}

	crc := crc16(nBytes)

	L.Push(lua.LNumber(crc)) /* push result */

	return 1 /* number of results */
}

func CheckCRCModbus(L *lua.LState) int {
	type LuaVariableMapTemplate struct {
		Variable []*byte
	}

	lv := L.ToTable(1)

	LuaVariableMap := LuaVariableMapTemplate{}
	if err := gluamapper.Map(lv, &LuaVariableMap); err != nil {
		zap.S().Warnf("GetCRC16 gluamapper.Map err %v", err)
	}

	nBytes := make([]byte, 0)
	if len(LuaVariableMap.Variable) > 0 {
		for _, v := range LuaVariableMap.Variable {
			nBytes = append(nBytes, *v)
		}
	}

	crc := crc16(nBytes)
	L.Push(lua.LNumber(crc)) /* push result */

	return 1 /* number of results */
}

func GetCRCModbusLittleEndian(L *lua.LState) int {
	type LuaVariableMapTemplate struct {
		Variable []*byte
	}

	lv := L.ToTable(1)

	LuaVariableMap := LuaVariableMapTemplate{}
	if err := gluamapper.Map(lv, &LuaVariableMap); err != nil {
		zap.S().Warnf("GetCRC16 gluamapper.Map err %v", err)
	}

	nBytes := make([]byte, 0)
	if len(LuaVariableMap.Variable) > 0 {
		for _, v := range LuaVariableMap.Variable {
			nBytes = append(nBytes, *v)
		}
	}
	crc := crc16(nBytes)
	crc16 := (crc&0x00FF)*256 + (crc / 256)
	L.Push(lua.LNumber(crc16)) /* push result */

	return 1 /* number of results */
}
