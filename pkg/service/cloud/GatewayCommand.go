package mqtt

import (
	"encoding/base64"
	"github.com/gookit/color"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"os"
	"strings"
	"zsxagw/api/domain"
	"zsxagw/core/io"
	"zsxagw/util"
)

// 处理网关操作 网关重启(reboot)、远程升级(upgrade)
// 获取表具设备(getdevice)、下发表具设备(setdevice)、PING(ping)
const (
	GETDEVICETYPE       = "GetDeviceType"
	GETCOLLECTORS       = "GetCollectors"
	SETCOLLECTORS       = "SetCollectors"
	SETINSTRUMENTS      = "SetInstruments"
	GETINSTRUMENTS      = "GetInstruments"
	SETINSTRUMENTTYPES  = "SetInstrumentTypes"
	SetInstrumentDriver = "SetInstrumentDriver"
	//------------------------------------------
	REBOOT     = "Reboot"
	UPGRADE    = "Upgrade"
	GETDEVICE  = "GetDevices"
	SETDEVICE  = "SetDevices"
	GETDEVICE2 = "GetDevices2"
	SETDEVICE2 = "SetDevices2"
	PING       = "Ping"
)

type commandExecutor func(params map[string]interface{}, client Client) (int, interface{})

var (
	GatewayCommand map[string]commandExecutor = initGatewayCommand()
)

func initGatewayCommand() map[string]commandExecutor {
	result := make(map[string]commandExecutor)
	result[GETDEVICETYPE] = getDeviceType
	result[GETCOLLECTORS] = getCollectors
	result[SETCOLLECTORS] = setCollectors
	result[SETINSTRUMENTS] = setInstruments
	result[GETINSTRUMENTS] = getInstruments
	result[SETINSTRUMENTTYPES] = setInstrumentTypes
	result[SetInstrumentDriver] = setInstrumentDriver
	//------------------------------------------------------
	result[REBOOT] = reboot
	result[UPGRADE] = upgrade
	result[GETDEVICE] = getDevice
	result[SETDEVICE] = setDevice
	result[GETDEVICE2] = getDevice2
	result[SETDEVICE2] = setDevice2
	result[PING] = ping

	return result
}

func setInstrumentDriver(params map[string]interface{}, client Client) (int, interface{}) {
	var err error
	instrument, ok := params["instrument"].(string)
	if !ok {
		zap.S().Errorf("RPC设置采集器失败: 参数instrument类型不正确")
		return 1, nil
	}
	driver, ok := params["driver"].(string)
	if !ok {
		zap.S().Errorf("RPC设置采集器失败: 参数driver类型不正确")
		return 1, nil
	}
	color.Greenln("RPC设置采集器:", instrument, driver)
	driverName, ok := params["driverName"].(string)
	find, err := client.deviceTypeRepository.Find(instrument)
	if err != nil {
		color.Redln("RPC设置采集器失败: 参数collectors类型不正确-")
		return 1, nil
	}
	color.Greenln(find.Driver)
	sDec, _ := base64.StdEncoding.DecodeString(driver)
	// 写到文件
	err = io.WriteFile("/plugin/"+driverName, sDec)
	if err != nil {
		color.Redln("RPC设置采集器失败: 写入驱动文件时发生错误", err.Error())
		return 1, nil
	}
	dir := io.GetPluginPath() + driverName
	err = io.Unzip(dir, io.GetPluginPath())
	if err != nil {
		color.Redln("RPC设置采集器失败: 解压驱动文件时发生错误", err.Error())
		return 1, nil
	}
	err = os.Remove(dir)
	if err != nil {
		color.Redln("RPC设置采集器失败: 删除驱动文件时发生错误", err.Error())
		return 0, nil
	}
	dn := strings.Split(driverName, ".")
	color.Greenln("RPC设置采集器:", dn[0], driverName)
	find.Driver = dn[0]
	err = client.deviceTypeRepository.Save(&find)
	if err != nil {
		color.Redln("RPC设置采集器失败: 保存驱动文件时发生错误", err.Error())
		return 1, nil
	}
	return 0, nil
}

func getInstruments(params map[string]interface{}, client Client) (int, interface{}) {
	all, err := client.deviceRepository.FindAll()
	if err != nil {
		zap.S().Errorf("RPC获取设备失败, ERROR: %v", err)
		return 1, nil
	}
	return 0, all
}

func setInstruments(params map[string]interface{}, client Client) (int, interface{}) {
	instruments, ok := params["instruments"].([]interface{})
	if !ok {
		zap.S().Errorf("RPC设置仪表失败: 参数instruments类型不正确")
		return 1, nil
	}
	finds, err := client.deviceRepository.FindAll()
	if err != nil {
		zap.S().Errorf("RPC设置仪表失败: 获取仪表列表时发生错误, ERROR: %v", err)
		return 1, nil
	}
	for _, find := range finds {
		err := client.deviceRepository.Delete(find.Name)
		if err != nil {
			color.Redln("RPC设置仪表失败: 删除仪表时发生错误, ERROR: %v")
			return 1, nil
		}
	}
	for _, v := range instruments {
		instrumentMap, ok := v.(map[string]interface{})
		if !ok {
			color.Redln("RPC设置仪表失败: 参数instruments解析失败")
			return 1, nil
		}
		c := instrumentMap["collector"].(map[string]interface{})
		cn := c["name"].(string)
		//instrumentMap["collectTime"] = time.Now()
		//instrumentMap["ReportTime"] = time.Now()
		var instrument domain.Device
		err := mapstructure.Decode(instrumentMap, &instrument)
		if err != nil {
			color.Redln("RPC设置仪表失败: 参数instruments类型转换失败", err.Error())
			return 1, nil
		}
		deviceType, err := client.deviceTypeRepository.Find(instrument.Type.Name)
		if err != nil {
			util.ToolUtil.Copy(&deviceType, instrument.Type)
			if err := client.deviceTypeRepository.Save(&deviceType); err != nil {
				color.Redln("RPC设置仪表失败: 保存设备类型时发生错误, ERROR: %v", err)
				return 1, nil
			}
		}
		instrument.Type = deviceType
		collector, err := client.collectorRepository.Find(cn)
		if err != nil {
			color.Redln("RPC设置仪表失败: 参数collectorName类型不正确")
			return 1, nil
		}
		instrument.Collector = collector
		if err := client.deviceRepository.Save(&instrument); err != nil {
			color.Redln("RPC设置仪表失败: 保存设备时发生错误, ERROR: %v", err)
			return 1, nil
		}
	}
	return 0, nil
}

func setInstrumentTypes(params map[string]interface{}, client Client) (int, interface{}) {
	instrumentTypes, ok := params["instrumentTypes"].([]interface{})
	if !ok {
		zap.S().Errorf("RPC设置仪表失败: 参数instruments类型不正确")
		return 1, nil
	}
	finds, err := client.deviceTypeRepository.FindAll()
	if err != nil {
		zap.S().Errorf("RPC设置仪表失败: 获取仪表类型列表时发生错误, ERROR: %v", err)
		return 1, nil
	}
	for _, find := range finds {
		err := client.deviceTypeRepository.Delete(find.Name)
		if err != nil {
			color.Redln("RPC设置仪表失败: 删除仪表类型时发生错误, ERROR: %v")
			return 1, nil
		}
	}
	for _, v := range instrumentTypes {
		instrumentMap, ok := v.(map[string]interface{})
		if !ok {
			color.Redln("RPC设置仪表失败: 参数instruments解析失败")
			return 1, nil
		}

		var instrument domain.DeviceType
		err = mapstructure.Decode(instrumentMap, &instrument)
		if err != nil {
			color.Redln("RPC设置仪表失败: 参数instruments类型转换失败", err.Error())
			return 1, nil
		}
		deviceType, err := client.deviceTypeRepository.Find(instrument.Name)
		if err != nil {
			var dt domain.DeviceType
			util.ToolUtil.Copy(&dt, instrument)
			if err := client.deviceTypeRepository.Save(&dt); err != nil {
				color.Redln("RPC设置仪表失败: 保存设备类型时发生错误, ERROR: %v")
				return 1, nil
			}
		}
		util.ToolUtil.Copy(&deviceType, instrument)
		//deviceType.Driver =
		if err := client.deviceTypeRepository.Save(&deviceType); err != nil {
			color.Redln("RPC设置仪表失败: 保存设备类型时发生错误, ERROR: %v")
			return 1, nil
		}

	}
	return 0, nil
}

func setCollectors(params map[string]interface{}, client Client) (int, interface{}) {
	collectors, ok := params["collectors"].([]interface{})
	if !ok {
		zap.S().Errorf("RPC设置采集器失败: 参数collectors类型不正确")
		return 1, nil
	}
	finds, err := client.collectorRepository.FindAll()
	if err != nil {
		zap.S().Errorf("RPC设置采集器失败: 获取采集器列表时发生错误, ERROR: %v", err)
		return 1, nil
	}
	for _, find := range finds {
		err := client.collectorRepository.Delete(find.Name)
		if err != nil {
			color.Redln("RPC设置采集器失败: 删除采集器时发生错误, ERROR: %v")
			return 1, nil
		}
	}
	for _, collectorInterface := range collectors {
		collector, ok := collectorInterface.(map[string]interface{})
		if !ok {
			zap.S().Errorf("RPC设置采集器失败: collector类型不正确")
			return 1, nil
		}
		var c domain.Collector
		if err := mapstructure.Decode(collector, &c); err != nil {
			zap.S().Errorf("RPC设置采集器失败: 解析collector失败, ERROR: %v", err)
			return 1, nil
		}
		find, err := client.collectorRepository.Find(c.Name)
		if err != nil {
			util.ToolUtil.Copy(&find, c)
			if err := client.collectorRepository.Save(&find); err != nil {
				zap.S().Errorf("RPC设置采集器失败: 保存采集器时发生错误, ERROR: %v", err)
				return 1, nil
			}
			return 0, nil
		}
		if find.Name != c.Name {
			zap.S().Errorf("RPC设置采集器失败: 采集器名称不匹配")
			return 1, nil
		}
		util.ToolUtil.Copy(&find, c)
		if err := client.collectorRepository.Save(&find); err != nil {
			zap.S().Errorf("RPC设置采集器失败: 保存采集器时发生错误, ERROR: %v", err)
			return 1, nil
		}
	}
	return 0, nil
}

func getCollectors(params map[string]interface{}, client Client) (int, interface{}) {
	all, err := client.collectorRepository.FindAll()
	if err != nil {
		zap.S().Errorf("RPC获取采集器失败, ERROR: %v", err)
		return 1, nil
	}
	return 0, all
}

func getDeviceType(params map[string]interface{}, client Client) (int, interface{}) {
	all, err := client.deviceTypeRepository.FindAll()
	if err != nil {
		zap.S().Errorf("RPC获取设备类型失败, ERROR: %v", err)
		return 1, nil
	}
	result := make([]DeviceType, 0, len(all))
	for _, d := range all {
		add := DeviceType{
			Name:   d.Name,
			Driver: d.Driver,
		}
		util.ToolUtil.Copy(&add.Properties, d.Properties)
		result = append(result, add)
	}
	return 0, result
}

func reboot(params map[string]interface{}, client Client) (int, interface{}) {
	zap.S().Info("执行RPC命令reboot")

	io.SystemReboot()

	return 0, nil
}

func upgrade(params map[string]interface{}, client Client) (int, interface{}) {
	zap.S().Info("执行RPC命令upgrade")

	sourceFile, ok := params["url"].(string)
	if !ok {
		zap.S().Error("升级失败，缺少url参数")
		return 1, nil
	}

	fileName := "zsxagw.zip"
	if len(sourceFile) > 0 {
		io.Upgrade(sourceFile, io.GetCurrentPath()+fileName)
	}

	return 0, nil
}

func getDevice(params map[string]interface{}, client Client) (int, interface{}) {
	zap.S().Info("执行RPC命令GetDevices")

	var result []Device

	devices, err := client.deviceRepository.FindAll()
	if err != nil {
		zap.S().Error("RPC获取设备失败")
		return 1, result
	}

	for _, d := range devices {
		device := Device{
			CollectName: d.Collector.Name,
			Name:        d.Name,
			Addr:        d.Address,
			Type:        d.Type.Name,
		}
		result = append(result, device)
	}

	return 0, result
}

func getDevice2(params map[string]interface{}, client Client) (int, interface{}) {
	zap.S().Info("执行RPC命令GetDevices2")

	var result []Device2

	devices, err := client.deviceRepository.FindAll()
	if err != nil {
		zap.S().Error("RPC获取设备失败")
		return 1, result
	}

	for _, d := range devices {
		device := Device2{
			CollectName: d.Collector.Name,
			Name:        d.Name,
			Addr:        d.Address,
			Type:        d.Type.Name,
			Alone:       d.Alone,
			Serial:      d.Serial,
		}
		result = append(result, device)
	}

	return 0, result
}

func setDevice(params map[string]interface{}, client Client) (int, interface{}) {
	zap.S().Info("执行RPC命令SetDevices")

	// 1. 获取设备列表
	devices := convertDevices(params)

	if devices == nil || len(devices) == 0 {
		return 0, nil
	}

	// 2. 删除设备信息
	if clears, err := client.deviceRepository.FindAll(); err == nil {
		for _, d := range clears {
			// 删除
			client.deviceRepository.Delete(d.Name)
		}
	}

	// 3. 添加设备信息
	for _, d := range devices {
		deviceType, err := client.deviceTypeRepository.Find(d.Type)
		if err != nil {
			zap.S().Debug("设备类型[" + d.Type + "]不存在！")
			continue
		}
		collector, err := client.collectorRepository.Find(d.CollectName)

		if err != nil {
			zap.S().Debug("采集接口[" + d.CollectName + "]不存在！")
			continue
		}

		device := &domain.Device{
			Name:           d.Name,
			Type:           deviceType,
			Address:        d.Addr,
			Collector:      collector,
			Alone:          false,
			Serial:         collector.Serial,
			Online:         false,
			CollectTime:    0,
			CollectTotal:   0,
			CollectSuccess: 0,
			ReportTime:     0,
			ReportTotal:    0,
			ReportSuccess:  0,
		}
		client.deviceRepository.Save(device)
	}

	return 0, nil
}

// ------------------------------------------------------------
func convertDevices(params map[string]interface{}) (result []Device) {
	result = []Device{}

	if devices, ok := params["devices"]; ok {
		if devs, ok := devices.([]interface{}); ok {
			for _, device := range devs {
				if dev, ok := device.(map[string]interface{}); ok {
					d := Device{}

					if collectName, ok := dev["collectName"]; ok {
						d.CollectName = convertString(collectName)
					}

					if name, ok := dev["name"]; ok {
						d.Name = convertString(name)
					}

					if addr, ok := dev["addr"]; ok {
						d.Addr = convertString(addr)
					}

					if type2, ok := dev["type"]; ok {
						d.Type = convertString(type2)
					}

					result = append(result, d)
				}
			}
		}
	}
	return result
}

func setDevice2(params map[string]interface{}, client Client) (int, interface{}) {
	zap.S().Info("执行RPC命令SetDevices2")

	// 1. 获取设备列表
	devices := convertDevices2(params)

	if devices == nil || len(devices) == 0 {
		return 1, nil
	}

	// 2. 删除设备信息
	if clears, err := client.deviceRepository.FindAll(); err == nil {
		for _, d := range clears {
			// 删除
			client.deviceRepository.Delete(d.Name)
		}
	}

	// 3. 添加设备信息
	for _, d := range devices {
		deviceType, err := client.deviceTypeRepository.Find(d.Type)
		if err != nil {
			zap.S().Debug("设备类型[" + d.Type + "]不存在！")
			continue
		}
		collector, err := client.collectorRepository.Find(d.CollectName)

		if err != nil {
			zap.S().Debug("采集接口[" + d.CollectName + "]不存在！")
			continue
		}

		device := &domain.Device{
			Name:           d.Name,
			Type:           deviceType,
			Address:        d.Addr,
			Collector:      collector,
			Alone:          d.Alone,
			Serial:         d.Serial,
			Online:         false,
			CollectTime:    0,
			CollectTotal:   0,
			CollectSuccess: 0,
			ReportTime:     0,
			ReportTotal:    0,
			ReportSuccess:  0,
		}

		client.deviceRepository.Save(device)
	}

	return 0, nil
}

// ------------------------------------------------------------
func convertDevices2(params map[string]interface{}) (result []Device2) {
	result = []Device2{}

	if devices, ok := params["devices"]; ok {
		if devs, ok := devices.([]interface{}); ok {
			for _, device := range devs {
				if dev, ok := device.(map[string]interface{}); ok {
					d := Device2{}

					if collectName, ok := dev["collectName"]; ok {
						d.CollectName = convertString(collectName)
					}

					if name, ok := dev["name"]; ok {
						d.Name = convertString(name)
					}

					if addr, ok := dev["addr"]; ok {
						d.Addr = convertString(addr)
					}

					if type2, ok := dev["type"]; ok {
						d.Type = convertString(type2)
					}

					if alone, ok := dev["alone"]; ok {
						d.Alone = convertBoolean(alone)
					}

					if d.Alone {
						if serial, ok := dev["serial"]; ok {
							d.Serial = *convertSerial(serial)
						}
					}

					result = append(result, d)
				}
			}
		}
	}
	return
}

func convertSerial(value interface{}) (result *domain.Serial) {
	result = &domain.Serial{}

	if serial, ok := value.(map[string]interface{}); ok {
		if name, ok := serial["name"]; ok {
			result.Name = convertString(name)
		}

		if deviceName, ok := serial["deviceName"]; ok {
			result.DeviceName = convertString(deviceName)
		}

		if baudRate, ok := serial["baudRate"]; ok {
			result.BaudRate = (int)(convertFloat64(baudRate))
		}
		if dataBit, ok := serial["dataBit"]; ok {
			result.DataBit = (int)(convertFloat64(dataBit))
		}

		if stopBit, ok := serial["stopBit"]; ok {
			result.StopBit = convertString(stopBit)
		}

		if check, ok := serial["check"]; ok {
			result.Check = convertString(check)
		}
	}

	return
}

func convertString(value interface{}) (result string) {
	result = ""

	if temp, ok := value.(string); ok {
		result = temp
	}

	return
}

func convertFloat64(value interface{}) (result float64) {
	result = 0

	if temp, ok := value.(float64); ok {
		result = temp
	}

	return
}

func convertBoolean(value interface{}) (result bool) {
	result = false

	if temp, ok := value.(bool); ok {
		result = temp
	}

	return
}

func ping(params map[string]interface{}, client Client) (int, interface{}) {
	return 0, nil
}
