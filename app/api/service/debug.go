package service

import (
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
	"tinyGW/app/api/repository"
	"tinyGW/app/api/schemas/req"
	"tinyGW/pkg/plugin/io"
	"tinyGW/pkg/service/collect"
	"tinyGW/pkg/service/collect/worker"
	"tinyGW/pkg/service/conf"
)

type DebugService interface {
	Test(testReq *req.DebugTestReq) (interface{}, error)
	Upgrade(c *gin.Context) error
	SystemStatus() (io.SystemStateTemplate, error)
	SystemReboot() error
	SystemNtp() (string, error)
}

type debugService struct {
	collectorServer      *collect.CollectorServer
	deviceRepository     repository.DeviceRepository
	reportTaskRepository repository.ReportTaskRepository
	config               *conf.Config
}

func (d debugService) Test(testReq *req.DebugTestReq) (interface{}, error) {
	testReq.Data = strings.ReplaceAll(testReq.Data, " ", "")
	if len(testReq.Data)%2 != 0 {
		return nil, fmt.Errorf("16进制数据格式不正确，必须是偶数个数字！")
	}
	commandRequest := worker.CommandRequest{
		Method: "test",
		Params: []worker.RequestParam{
			{
				ClientID:  testReq.CollectorName,
				CmdName:   "test",
				CmdParams: make(map[string]interface{}),
			},
		},
	}
	reqData, _ := hex.DecodeString(testReq.Data)
	commandRequest.Params[0].CmdParams["param"] = reqData
	commandRequest.ResponseParamChan = make(chan worker.ResponseParam, 1)
	if work, ok := d.collectorServer.FindByCollectorName(testReq.CollectorName); ok {
		work.CommandTask(commandRequest)

		resp := <-commandRequest.ResponseParamChan
		if resp.CmdStatus == 0 {
			data := resp.CmdResult.([]byte)
			return data, nil
		} else {
			return nil, fmt.Errorf("命令执行失败：%s", resp.Err)
		}
	} else {
		return nil, fmt.Errorf("找不到采集器：%s", testReq.CollectorName)
	}
}

func (d debugService) Upgrade(c *gin.Context) error {
	file, err := c.FormFile("FileName")
	if err != nil {
		return err
	}
	filepath := io.GetCurrentPath() + file.Filename
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		return fmt.Errorf("文件上传出错！")
	}

	err = io.Unzip(filepath, io.GetCurrentPath())
	if err != nil {
		return fmt.Errorf("解压缩文件出错！")
	}
	os.Remove(filepath)
	io.OsCommand("chmod +x " + io.GetCurrentPath() + "tinyGW")
	io.SystemReboot()
	return nil
}

func (d debugService) SystemStatus() (io.SystemStateTemplate, error) {
	io.SystemState.Name = d.config.Cloud.ClientId
	if d.config.Serial.OpenEveryTime {
		io.SystemState.OpenEveryTime = "支持不同波特率"
	} else {
		io.SystemState.OpenEveryTime = "不支持不同波特率"
	}
	io.GetMemState()
	io.GetDiskState()
	io.GetRuntime()
	// 设备在线率计算
	deviceTotalCnt := 0
	deviceOnlineCnt := 0
	devices, _ := d.deviceRepository.FindAll()
	zap.S().Debug(len(devices))
	for _, device := range devices {
		zap.S().Debug(device)
		deviceTotalCnt = deviceTotalCnt + 1
		if device.Online == true {
			deviceOnlineCnt = deviceOnlineCnt + 1
		}
	}

	if deviceOnlineCnt == 0 {
		io.SystemState.DeviceTotal = "0"
		io.SystemState.DeviceOnline = "0"
	} else {
		io.SystemState.DeviceTotal = fmt.Sprintf("%d", deviceTotalCnt)
		io.SystemState.DeviceOnline = fmt.Sprintf("%5.2f", float32(deviceOnlineCnt*100.0/deviceTotalCnt))
	}
	io.SystemState.SoftVer = conf.Version
	return io.SystemState, nil
}

func (d debugService) SystemReboot() error {
	io.SystemReboot()
	return nil
}

func (d debugService) SystemNtp() (string, error) {
	io.ExecuteNtp(d.config)
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	return timeStr, nil
}

func NewDebugService(collectorServer *collect.CollectorServer, deviceRepository repository.DeviceRepository, reportTaskRepository repository.ReportTaskRepository, config *conf.Config) DebugService {
	return &debugService{
		collectorServer:      collectorServer,
		deviceRepository:     deviceRepository,
		reportTaskRepository: reportTaskRepository,
		config:               config,
	}
}
