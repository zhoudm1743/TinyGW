package worker

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"strings"
	"time"
	"tinyGW/app/api/repository"
	"tinyGW/app/models"
	"tinyGW/pkg/service/collect/channel"
	"tinyGW/pkg/service/collect/collector"
	"tinyGW/pkg/service/conf"
	"tinyGW/pkg/service/script"
)

type Worker interface {
	Start()
	Stop()
	CommandTask(task any)    // 指派高优先级任务
	CollectTask(task any)    // 指派正常任务
	CollectTaskIsFull() bool // 采集通道是否为空
}

var _ Worker = (*worker)(nil)

type worker struct {
	channel.PriorityChannel
	collector.Collector
	script.Runner
	deviceTypeRepository repository.DeviceTypeRepository
	deviceRepository     repository.DeviceRepository
	dataChan             chan []byte
	stopChan             chan int
	config               *conf.Config
}

func NewWorker(
	domainCollector models.Collector,
	deviceRepository repository.DeviceRepository,
	config *conf.Config,
) Worker {
	result := &worker{
		PriorityChannel:  channel.NewPriorityChannel(),
		Collector:        collector.ConnectorFactory(domainCollector),
		Runner:           script.NewLuaRunner(),
		deviceRepository: deviceRepository,
		dataChan:         make(chan []byte, 1024),
		stopChan:         make(chan int, 1),
		config:           config,
	}
	zap.S().Info("创建数据采集线程！", result.Collector)
	// RPC 命令优先级高
	result.PriorityChannel.SetPriorWorker(result.CommandExecutor)
	// 采集任务优先级低
	result.PriorityChannel.SetNormalWorker(result.CollectExecutor)

	return result
}

func (w *worker) CommandTask(task any) {
	w.PriorityChannel.DispatchPriorTask(task)
}

func (w *worker) CollectTask(task any) {
	w.PriorityChannel.DispatchNormalTask(task)
}

func (w *worker) CollectTaskIsFull() bool {
	return w.PriorityChannel.NormalTaskIsFull()
}

func (w *worker) Start() {
	// 如果不是每次都打开，那么只在这里打开一次
	if !w.config.Serial.OpenEveryTime {
		w.Collector.Open(nil)
	}

	go w.BlockRead()
	w.PriorityChannel.Start()
}

func (w *worker) Stop() {
	w.PriorityChannel.Stop()
	w.stopChan <- 1

	// 如果没事每次都打开，那么只在这里关闭一次
	if !w.config.Serial.OpenEveryTime {
		w.Collector.Close()
	}
}

// BlockRead 阻塞读取数据
func (w *worker) BlockRead() {
	for {
		select {
		case <-w.stopChan:
			zap.S().Info("串口数据读取线程正确退出！")
			return
		default:
			buf := make([]byte, 1024)
			count := w.Collector.Read(buf)
			if count > 0 {
				w.dataChan <- append([]byte(nil), buf[:count]...) // 复制数据
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (w *worker) CollectExecutor(task any) {
	device, ok := task.(models.Device)
	if !ok {
		zap.S().Error("数据采集执行器无法转换Device参数")
		return
	}
	zap.S().Info("开始采集设备: ", device.Name)

	// 备份 device（通过序列化，反序列化深度复制）
	tempBuff, _ := json.Marshal(device)
	deviceCopy := models.Device{}
	_ = json.Unmarshal(tempBuff, &deviceCopy)
	driverName := device.Type.Driver
	// 打开驱动
	//color.Redln("开始打开设备驱动！", device.Type.Driver)
	if len(driverName) <= 0 {
		find, err := w.deviceTypeRepository.Find(device.Type.Name)
		if err != nil {
			zap.S().Error("找不到设备类型！", err)
			return
		}
		driverName = find.Driver
	}
	w.Runner.Open(driverName)
	defer w.Runner.Close()

	// 如果每次都打开，那么在这里打开
	if w.config.Serial.OpenEveryTime {
		w.Collector.Open(&device)
		defer w.Collector.Close()
	}

	// 数据读取，超过30次，自动退出
	for step := 0; step < 30; step++ {
		data, result, continued := w.Runner.GenerateGetRealVariables(device.Address, step)
		if result {
			// 数据发送
			zap.S().Infof("向设备[%s]发送数据[%d:%X]", device.Name, len(data), data)
			w.Collector.Write(data)

			rxBuf := make([]byte, 1024)
			rxBufCnt := 0
			rxTotalBuf := make([]byte, 0)
			rxTotalBufCnt := 0

			reader := func() bool {
				for {
					select {
					case rxBuf = <-w.dataChan:
						rxBufCnt = len(rxBuf)
						if rxBufCnt > 0 {
							rxTotalBufCnt += rxBufCnt
							rxTotalBuf = append(rxTotalBuf, rxBuf[:rxBufCnt]...)
							rxBufCnt = 0
							rxBuf = rxBuf[0:0]
						}

						if rxTotalBufCnt > 0 {
							zap.S().Infof("从设备[%s]接收数据[%d:%X]", device.Name, rxTotalBufCnt, rxTotalBuf)
						}
						// 仪表倍率优先于类型倍率
						if device.Scale > 1 {
							for k, p := range device.Type.Properties {
								if p.Name == "dev_consumption" || p.Name == "dev_flow" {
									device.Type.Properties[k].Scale = device.Scale
								}
							}
						}
						var tempVariables []models.DeviceProperty // = make([]models.DeviceProperty, 0)
						if rxTotalBufCnt > 0 && w.Runner.AnalysisRx(device.Address, device.Type.Properties, rxTotalBuf, rxTotalBufCnt, &tempVariables) {
							zap.S().Infof("从设备[%s]接收数据后，正确解析数据", device.Name)
							device.CollectTime = time.Now().Unix()
							device.Online = true
							device.CollectTotal += 1
							device.CollectSuccess += 1
							if deviceCopy.CollectTime >= 0 {
								alarmStr := []string{}
								copyDeviceProperty := deviceCopy.Type.Properties
								for _, temp := range tempVariables {
									val, ok := temp.Value.(float64)
									if !ok {
										continue
									}
									if temp.IsAlarm {
										for _, tempDeviceProperty := range copyDeviceProperty {
											if tempDeviceProperty.Name == temp.Name {
												copyVal, ok := tempDeviceProperty.Value.(float64)
												if !ok {
													continue
												}
												if val < copyVal {
													alarmStr = append(alarmStr, fmt.Sprintf("%s:%02f 读数异常！比上次读数[%02f]要小!", temp.Name, val, copyVal))
												}
												plusVal := copyVal + temp.Threshold
												if val > plusVal {
													alarmStr = append(alarmStr, fmt.Sprintf("%s:%02f 读数异常！读数超出预警值!", temp.Name, val, plusVal))
												}
											}
										}
										if len(alarmStr) > 0 {
											device.AlarmStatus = true
											device.AlarmReason = strings.Join(alarmStr, ",")
											device.AlarmTime = time.Now().Unix()
											device.AlarmTotal += 1
										} else {
											device.AlarmStatus = false
											device.AlarmReason = ""
											device.AlarmTime = 0
										}
									}
								}
								for k, p := range device.Type.Properties {
									if p.AutoCalc {
										for _, cp := range deviceCopy.Type.Properties {
											if cp.Name != p.Name {
												continue
											}
											pv, ok := p.Value.(float64)
											if !ok {
												continue
											}
											cpv, ok := cp.Value.(float64)
											if cpv > 0 {
												device.Type.Properties[k].Used = pv - cpv
											} else {
												if device.InitialVal > 0 {
													device.Type.Properties[k].Used = pv - device.InitialVal
												} else {
													device.Type.Properties[k].Used = 0
												}
											}
										}
									}
								}
							}
							w.deviceRepository.Save(&device)
							return false
						}
					case <-time.After(time.Duration(w.Collector.GetTimeout()) * time.Millisecond):
						if rxTotalBufCnt > 0 {
							zap.S().Infof("[超时]从设备[%s]接收数据[%d:%X]", device.Name, rxTotalBufCnt, rxTotalBuf)
						}
						// 仪表倍率优先于类型倍率
						if device.Scale > 1 {
							for k, p := range device.Type.Properties {
								if p.Name == "dev_consumption" || p.Name == "dev_flow" {
									device.Type.Properties[k].Scale = device.Scale
								}
							}
						}
						var tempVariables []models.DeviceProperty // = make([]models.DeviceProperty, 0)
						if rxTotalBufCnt > 0 && w.Runner.AnalysisRx(device.Address, device.Type.Properties, rxTotalBuf, rxTotalBufCnt, &tempVariables) {
							zap.S().Infof("从设备[%s]接收数据后，超时，但正确解析数据", device.Name)
							device.CollectTime = time.Now().Unix()
							device.Online = true
							device.CollectTotal += 1
							device.CollectSuccess += 1
							if deviceCopy.CollectTime >= 0 {
								alarmStr := []string{}
								copyDeviceProperty := deviceCopy.Type.Properties
								for _, temp := range tempVariables {
									val, ok := temp.Value.(float64)
									if !ok {
										continue
									}
									if temp.IsAlarm {
										for _, tempDeviceProperty := range copyDeviceProperty {
											if tempDeviceProperty.Name == temp.Name {
												copyVal, ok := tempDeviceProperty.Value.(float64)
												if !ok {
													continue
												}
												if val < copyVal {
													alarmStr = append(alarmStr, fmt.Sprintf("%s:%02f 读数异常！比上次读数[%02f]要小!", temp.Name, val, copyVal))
												}
												plusVal := copyVal + temp.Threshold
												if val > plusVal {
													alarmStr = append(alarmStr, fmt.Sprintf("%s:%02f 读数异常！读数超出预警值!", temp.Name, val, plusVal))
												}
											}
										}
										if len(alarmStr) > 0 {
											device.AlarmStatus = true
											device.AlarmReason = strings.Join(alarmStr, ",")
											device.AlarmTime = time.Now().Unix()
											device.AlarmTotal += 1
										} else {
											device.AlarmStatus = false
											device.AlarmReason = ""
											device.AlarmTime = 0
										}
									}
								}
							}
							for k, p := range device.Type.Properties {
								if p.AutoCalc {
									for _, cp := range deviceCopy.Type.Properties {
										if cp.Name != p.Name {
											continue
										}
										pv, ok := p.Value.(float64)
										if !ok {
											continue
										}
										cpv, ok := cp.Value.(float64)
										if cpv > 0 {
											device.Type.Properties[k].Used = pv - cpv
										} else {
											if device.InitialVal > 0 {
												device.Type.Properties[k].Used = pv - device.InitialVal
											} else {
												device.Type.Properties[k].Used = 0
											}
										}
									}
								}
							}
							w.deviceRepository.Save(&device)
							return false
						}

						// 此处超时，可以设置设备离线
						deviceCopy.CollectTime = time.Now().Unix()
						deviceCopy.Online = false
						deviceCopy.CollectTotal += 1
						w.deviceRepository.Save(&deviceCopy)
						return true
					}
				}
			}
			if timeout := reader(); timeout {
				// 如果一次读取超时，那么我们认为读取该设备的其他属性也会超时，所以就退出了
				zap.S().Error("超时退出")
				device.AlarmStatus = true
				device.AlarmTime = time.Now().Unix()
				device.AlarmReason = "未上报读数"
				device.AlarmTotal += 1
				w.deviceRepository.Save(&device)
				break
			}
		}
		if !continued {
			break
		}
		time.Sleep(time.Duration(w.Collector.GetInterval()) * time.Millisecond)
	}

	zap.S().Info("结束采集设备:", device.Name)
}

func (w *worker) CommandExecutor(task any) {
	commandRequest, ok := task.(CommandRequest)
	if !ok {
		zap.S().Error("RPC执行器无法转换CommandRequest参数")
		return
	}
	requestParam := commandRequest.Params[0]

	responseParam := ResponseParam{}
	responseParam.ClientID = requestParam.ClientID
	responseParam.CmdName = requestParam.CmdName
	responseParam.CmdStatus = 1
	responseParam.CmdResult = nil
	responseParam.Err = ""

	// 串口测试
	if commandRequest.Method == "test" {
		w.CommandTest(commandRequest)
		return
	}

	device, err := w.deviceRepository.Find(requestParam.DeviceName)
	if err != nil {
		responseParam.Err = "目标设备不存在！请同步网关设备列表！"
		responseParam.CmdStatus = 1
		commandRequest.ResponseParamChan <- responseParam
		return
	}
	tempBuff, _ := json.Marshal(device)
	deviceCopy := models.Device{}
	_ = json.Unmarshal(tempBuff, &deviceCopy)

	driverName := device.Type.Driver
	// 打开驱动
	//color.Redln("开始打开设备驱动！", device.Type.Driver)
	if len(driverName) <= 0 {
		find, err := w.deviceTypeRepository.Find(device.Type.Name)
		if err != nil {
			zap.S().Error("找不到设备类型！", err, device.Type.Name)
			responseParam.Err = "网关找不到设备类型！"
			responseParam.CmdStatus = 1
			commandRequest.ResponseParamChan <- responseParam
			return
		}
		driverName = find.Driver
	}
	// 打开设备驱动
	w.Runner.Open(driverName)
	defer w.Runner.Close()

	// 如果每次都打开，那么在这里打开
	if w.config.Serial.OpenEveryTime {
		w.Collector.Open(&device)
		defer w.Collector.Close()
	}
	// 串口命令
	// 只有最后一条有返回结果

	// 至少有一条参数（命令）
	for _, requestParam = range commandRequest.Params {
		responseParam.CmdName = requestParam.CmdName
		device, err = w.deviceRepository.Find(requestParam.DeviceName)
		if err != nil {
			break
		}
		// 数据读取，超过30次，自动退出
		for step := 0; step < 30; step++ {
			cmdParams, _ := json.Marshal(&(requestParam.CmdParams))
			data, result, continued := w.Runner.DeviceCustomCmd(device.Address, requestParam.CmdName, string(cmdParams), 0)
			if result {
				// 数据发送
				zap.S().Infof("RPC向设备[%s]发送数据[%d:%X]", device.Name, len(data), data)
				w.Collector.Write(data)

				rxBuf := make([]byte, 1024)
				rxBufCnt := 0
				rxTotalBuf := make([]byte, 0)
				rxTotalBufCnt := 0

				reader := func() bool {
					for {
						select {
						case rxBuf = <-w.dataChan:
							rxBufCnt = len(rxBuf)
							if rxBufCnt > 0 {
								rxTotalBufCnt += rxBufCnt
								rxTotalBuf = append(rxTotalBuf, rxBuf[:rxBufCnt]...)
								rxBufCnt = 0
								rxBuf = rxBuf[0:0]
							}
							if rxTotalBufCnt > 0 {
								zap.S().Infof("从设备[%s]接收数据[%d:%X]", device.Name, rxTotalBufCnt, rxTotalBuf)
							}
							// 仪表倍率优先于类型倍率
							if device.Scale > 1 {
								for k, p := range device.Type.Properties {
									if p.Name == "dev_consumption" || p.Name == "dev_flow" {
										device.Type.Properties[k].Scale = device.Scale
									}
								}
							}
							var tempVariables []models.DeviceProperty //= make([]models.DeviceProperty, 0)
							if rxTotalBufCnt > 0 && w.Runner.AnalysisRx(device.Address, device.Type.Properties, rxTotalBuf, rxTotalBufCnt, &tempVariables) {
								device.CollectTime = time.Now().Unix()
								device.Online = true
								device.CollectTotal += 1
								device.CollectSuccess += 1
								if deviceCopy.CollectTime >= 0 {
									alarmStr := []string{}
									copyDeviceProperty := deviceCopy.Type.Properties
									for _, temp := range tempVariables {
										val, ok := temp.Value.(float64)
										if !ok {
											continue
										}
										if temp.IsAlarm {
											for _, tempDeviceProperty := range copyDeviceProperty {
												if tempDeviceProperty.Name == temp.Name {
													copyVal, ok := tempDeviceProperty.Value.(float64)
													if !ok {
														continue
													}
													if val < copyVal {
														alarmStr = append(alarmStr, fmt.Sprintf("%s:%02f 读数异常！比上次读数[%02f]要小!", temp.Name, val, copyVal))
													}
													plusVal := copyVal + temp.Threshold
													if val > plusVal {
														alarmStr = append(alarmStr, fmt.Sprintf("%s:%02f 读数异常！读数超出预警值!", temp.Name, val, plusVal))
													}
												}
											}
											if len(alarmStr) > 0 {
												device.AlarmStatus = true
												device.AlarmReason = strings.Join(alarmStr, ",")
												device.AlarmTime = time.Now().Unix()
												device.AlarmTotal += 1
											} else {
												device.AlarmStatus = false
												device.AlarmReason = ""
												device.AlarmTime = 0
											}
										}
									}
									for k, p := range device.Type.Properties {
										if p.AutoCalc {
											for _, cp := range deviceCopy.Type.Properties {
												if cp.Name != p.Name {
													continue
												}
												pv, ok := p.Value.(float64)
												if !ok {
													continue
												}
												cpv, ok := cp.Value.(float64)
												if cpv > 0 {
													device.Type.Properties[k].Used = pv - cpv
												} else {
													if device.InitialVal > 0 {
														device.Type.Properties[k].Used = pv - device.InitialVal
													} else {
														device.Type.Properties[k].Used = 0
													}
												}
											}
										}
									}
								}
								w.deviceRepository.Save(&device)

								responseParam.CmdStatus = 0
								responseParam.DeviceName = device.Name
								tempMap := make(map[string]interface{})
								for _, property := range tempVariables {
									tempMap[property.Name] = property.Value
								}
								responseParam.CmdResult = tempMap
								return true
							}
						case <-time.After(time.Duration(w.Collector.GetTimeout()) * time.Millisecond):
							if rxTotalBufCnt > 0 {
								zap.S().Infof("从设备[%s]接收数据[%d:%X]", device.Name, rxTotalBufCnt, rxTotalBuf)
							}
							// 仪表倍率优先于类型倍率
							if device.Scale > 1 {
								for k, p := range device.Type.Properties {
									if p.Name == "dev_consumption" || p.Name == "dev_flow" {
										device.Type.Properties[k].Scale = device.Scale
									}
								}
							}
							var tempVariables []models.DeviceProperty //= make([]models.DeviceProperty, 0)
							if rxTotalBufCnt > 0 && w.Runner.AnalysisRx(device.Address, device.Type.Properties, rxTotalBuf, rxTotalBufCnt, &tempVariables) {
								device.CollectTime = time.Now().Unix()
								device.Online = true
								device.CollectTotal += 1
								device.CollectSuccess += 1
								if deviceCopy.CollectTime >= 0 {
									alarmStr := []string{}
									copyDeviceProperty := deviceCopy.Type.Properties
									for _, temp := range tempVariables {
										val, ok := temp.Value.(float64)
										if !ok {
											continue
										}
										if temp.IsAlarm {
											for _, tempDeviceProperty := range copyDeviceProperty {
												if tempDeviceProperty.Name == temp.Name {
													copyVal, ok := tempDeviceProperty.Value.(float64)
													if !ok {
														continue
													}
													if val < copyVal {
														alarmStr = append(alarmStr, fmt.Sprintf("%s:%02f 读数异常！比上次读数[%02f]要小!", temp.Name, val, copyVal))
													}
													plusVal := copyVal + temp.Threshold
													if val > plusVal {
														alarmStr = append(alarmStr, fmt.Sprintf("%s:%02f 读数异常！读数超出预警值!", temp.Name, val, plusVal))
													}
												}
											}
											if len(alarmStr) > 0 {
												device.AlarmStatus = true
												device.AlarmReason = strings.Join(alarmStr, ",")
												device.AlarmTime = time.Now().Unix()
												device.AlarmTotal += 1
											} else {
												device.AlarmStatus = false
												device.AlarmReason = ""
												device.AlarmTime = 0
											}
										}
									}
									for k, p := range device.Type.Properties {
										if p.AutoCalc {
											for _, cp := range deviceCopy.Type.Properties {
												if cp.Name != p.Name {
													continue
												}
												pv, ok := p.Value.(float64)
												if !ok {
													continue
												}
												cpv, ok := cp.Value.(float64)
												if cpv > 0 {
													device.Type.Properties[k].Used = pv - cpv
												} else {
													if device.InitialVal > 0 {
														device.Type.Properties[k].Used = pv - device.InitialVal
													} else {
														device.Type.Properties[k].Used = 0
													}
												}
											}
										}
									}
								}
								w.deviceRepository.Save(&device)
								responseParam.CmdStatus = 0
								responseParam.DeviceName = device.Name
								tempMap := make(map[string]interface{})
								for _, property := range tempVariables {
									tempMap[property.Name] = property.Value
								}
								responseParam.CmdResult = tempMap
								return true
							}
							responseParam.CmdStatus = 1
							responseParam.CmdResult = nil
							return true
						}
					}
				}
				if timeout := reader(); timeout {
					// 如果一次读取超时，那么我们认为读取该设备的其他属性也会超时，所以就退出了
					zap.S().Error("超时退出")
					break
				}
			}
			if !continued {
				break
			}
			time.Sleep(time.Duration(w.Collector.GetInterval()) * time.Millisecond)
		}
	}

	commandRequest.ResponseParamChan <- responseParam

	zap.S().Info("结束执行RPC命令:", commandRequest.Method)
}

func (w *worker) CommandTest(commandRequest CommandRequest) {
	// 如果每次都打开，那么在这里打开
	if w.config.Serial.OpenEveryTime {
		zap.S().Error("每次都打开设备，不支持调试!")
		return
	}

	if commandRequest.Method != "test" {
		return
	}

	requestParam := commandRequest.Params[0]

	zap.S().Info("开始执行测试命令:", requestParam.CmdName)

	responseParam := ResponseParam{}

	responseParam.CmdStatus = 0
	responseParam.CmdResult = nil

	// 发送指令
	data, ok := requestParam.CmdParams["param"].([]byte)
	if !ok {
		zap.S().Error("串口测试无法获取参数")
		return
	}
	w.Collector.Write(data)
	zap.S().Infof("发送【测试】数据[%d:%X]", len(data), data)
	rxBuf := make([]byte, 1024)
	rxBufCnt := 0
	rxTotalBuf := make([]byte, 0)
	rxTotalBufCnt := 0

	// 接收指令
	func() {
		for {
			select {
			case rxBuf = <-w.dataChan:
				rxBufCnt = len(rxBuf)
				if rxBufCnt > 0 {
					rxTotalBufCnt += rxBufCnt
					rxTotalBuf = append(rxTotalBuf, rxBuf[:rxBufCnt]...)
					rxBufCnt = 0
					rxBuf = rxBuf[0:0]
				}
			case <-time.After(time.Duration(w.Collector.GetTimeout()) * time.Millisecond):
				responseParam.CmdResult = rxTotalBuf[:rxTotalBufCnt]
				if rxTotalBufCnt == 0 {
					responseParam.CmdStatus = 1
				} else {
					responseParam.CmdStatus = 0
					if rxTotalBufCnt > 0 {
						zap.S().Infof("接收【测试】数据[%d:%X]", rxTotalBufCnt, rxTotalBuf)
					}
				}
				return
			}
		}
	}()

	commandRequest.ResponseParamChan <- responseParam
	zap.S().Infof("结束执行测试命令:", requestParam.CmdName)
}
