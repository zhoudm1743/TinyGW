package cloud

import (
	"encoding/base64"
	"github.com/gookit/color"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"os"
	"strings"
	"tinyGW/app/models"
	"tinyGW/pkg/plugin/io"
	"tinyGW/pkg/util"
)

// 处理网关操作 网关重启(reboot)、远程升级(upgrade)
// 获取表具设备(getdevice)、下发表具设备(setdevice)、PING(ping)
const (
	GETLOG              = "GetLog"
	GETDEVICETYPE       = "GetDeviceType"
	GETCOLLECTORS       = "GetCollectors"
	SETCOLLECTORS       = "SetCollectors"
	SETINSTRUMENTS      = "SetInstruments"
	GETINSTRUMENTS      = "GetInstruments"
	SETINSTRUMENTTYPES  = "SetInstrumentTypes"
	SetInstrumentDriver = "SetInstrumentDriver"
	//------------------------------------------
	REBOOT  = "Reboot"
	UPGRADE = "Upgrade"
	PING    = "Ping"
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
	result[GETLOG] = getLog
	result[REBOOT] = reboot
	result[UPGRADE] = upgrade
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
		var instrument models.Device
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
		instrument.Type = &deviceType
		collector, err := client.collectorRepository.Find(cn)
		if err != nil {
			var coll models.Collector
			util.ToolUtil.Copy(&coll, c)
			if err := client.collectorRepository.Save(&coll); err != nil {
				color.Redln("RPC<UNK>: <UNK>, ERROR: %v", err)
				return 1, nil
			}
			collector = coll
		}
		instrument.Collector = &collector
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

		var instrument models.DeviceType
		err = mapstructure.Decode(instrumentMap, &instrument)
		if err != nil {
			color.Redln("RPC设置仪表失败: 参数instruments类型转换失败", err.Error())
			return 1, nil
		}
		deviceType, err := client.deviceTypeRepository.Find(instrument.Name)
		if err != nil {
			util.ToolUtil.Copy(&deviceType, instrument)
			if err := client.deviceTypeRepository.Save(&deviceType); err != nil {
				color.Redln("RPC设置仪表失败: 保存设备类型时发生错误, ERROR: %v", err)
				return 1, nil
			}
		}
		util.ToolUtil.Copy(&deviceType, instrument)
		//deviceType.Driver =
		if err := client.deviceTypeRepository.Save(&deviceType); err != nil {
			color.Redln("RPC设置仪表失败: 保存设备类型时发生错误, ERROR: %v", err)
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
		err = client.collectorRepository.Delete(find.Name)
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
		var c models.Collector
		if err = mapstructure.Decode(collector, &c); err != nil {
			zap.S().Errorf("RPC设置采集器失败: 解析collector失败, ERROR: %v", err)
			return 1, nil
		}
		find, err := client.collectorRepository.Find(c.Name)
		color.Greenln("RPC<UNK>:", find.Name)
		if err != nil {
			var coll models.Collector
			util.ToolUtil.Copy(&coll, c)
			if err := client.collectorRepository.Save(&coll); err != nil {
				zap.S().Errorf("RPC设置采集器失败: 保存采集器时发生错误, ERROR: %v", err)
				return 1, nil
			}
			color.Greenln("RPC设置采集器:", c.Name)
			find = coll
		}
		if find.Name != c.Name {
			zap.S().Errorf("RPC设置采集器失败: 采集器名称不匹配")
			return 1, nil
		}
		find = models.Collector{}
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
	result := make([]models.DeviceType, 0, len(all))
	for _, d := range all {
		add := models.DeviceType{
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

func ping(params map[string]interface{}, client Client) (int, interface{}) {
	return 0, nil
}

func getLog(params map[string]interface{}, client Client) (int, interface{}) {
	// 获取主日志路径
	mainLogPath := io.GetLogPath()
	// 备用日志路径
	fallbackLogPath := "/www/wwwlogs/go/zsxagw.log"

	// 优先检查主路径
	logPath := mainLogPath
	if !io.PathExists(logPath) {
		// 主路径不存在时使用备用路径
		logPath = fallbackLogPath
		if !io.PathExists(logPath) {
			zap.S().Errorf("日志文件不存在，主路径:%s 备用路径:%s", mainLogPath, fallbackLogPath)
			return 1, nil
		}
	}

	content, err := io.ReadLastNLines(logPath, 100)
	if err != nil {
		zap.S().Errorf("读取日志文件失败: %v", err)
		return 1, nil
	}

	return 0, content
}
