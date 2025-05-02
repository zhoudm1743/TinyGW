package task

import (
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"sync"
	"tinyGW/app/api/repository"
	"tinyGW/app/models"
	"tinyGW/pkg/service/collect"
	"tinyGW/pkg/service/event"
)

type (
	// CollectTaskServer 采集任务服务器
	CollectTaskServer struct {
		tasks            *sync.Map
		collectorServer  *collect.CollectorServer
		deviceRepository repository.DeviceRepository
		event            *event.EventService
	}
)

// NewCollectTaskServer 实例化
func NewCollectTaskServer(collectorServer *collect.CollectorServer, deviceRepository repository.DeviceRepository, e *event.EventService) *CollectTaskServer {
	return &CollectTaskServer{
		tasks:            &sync.Map{},
		collectorServer:  collectorServer,
		deviceRepository: deviceRepository,
		event:            e,
	}
}

// InitCollectTaskServer 初始化
func InitCollectTaskServer(collectTaskServer *CollectTaskServer, repository repository.CollectTaskRepository, e *event.EventService) {
	zap.S().Info("初始化采集定时任务")
	if tasks, err := repository.FindAll(); err == nil {
		for _, task := range tasks {
			collectTaskServer.Add(&task)
		}
	}
	e.Subscribe("CollectTask_Add", func(e event.Event) {
		zap.S().Info("新增采集定时任务", e.Data)
		task := e.Data.(*models.CollectTask)
		collectTaskServer.Add(task)
	})
	e.Subscribe("CollectTask_Delete", func(e event.Event) {
		zap.S().Info("删除采集定时任务", e.Data)
		name := e.Data.(string)
		collectTaskServer.Delete(name)
	})
	e.Subscribe("CollectTask_Update", func(e event.Event) {
		zap.S().Info("修改采集定时任务", e.Data)
		task := e.Data.(*models.CollectTask)
		collectTaskServer.Update(task)
	})
}

// Add 新增定时任务，如果定时任务是开启状态，则开启定时任务
func (cts *CollectTaskServer) Add(task *models.CollectTask) {
	zap.S().Info("新增定时任务", task)
	c := cron.New()
	currentDevices := task.DeviceList
	c.AddFunc(task.Cron, func() {
		zap.S().Info("定时任务执行", task)
		cts.Collect(currentDevices)
	})
	cts.tasks.Store(task.Name, c)

	// 启动定时任务
	if task.Status == 1 {
		cts.Start(task)
	}
}

// Delete 删除定时任务，如果定时任务是开启状态，则停止定时任务
func (cts *CollectTaskServer) Delete(name string) {
	zap.S().Info("删除定时任务", name)
	if value, loaded := cts.tasks.LoadAndDelete(name); loaded {
		c := value.(*cron.Cron)
		c.Stop()
	}
}

// Update 修改定时任务，删除存在的定时任务，同时增加一个新的定时任务
func (cts *CollectTaskServer) Update(task *models.CollectTask) {
	cts.Delete(task.Name)
	cts.Add(task)
}

// Start 启动定时任务
func (cts *CollectTaskServer) Start(task *models.CollectTask) {
	zap.S().Info("启动定时任务", task)
	if value, loaded := cts.tasks.Load(task.Name); loaded {
		c := value.(*cron.Cron)
		c.Start()
	}
}

// Stop 停止定时任务
func (cts *CollectTaskServer) Stop(task *models.CollectTask) {
	zap.S().Info("停止定时任务", task)
	if value, loaded := cts.tasks.Load(task.Name); loaded {
		c := value.(*cron.Cron)
		c.Stop()
	}
}

// Collect 数据采集
func (cts *CollectTaskServer) Collect(ds []string) {
	zap.S().Info("采集开始：CollectTaskServer->Collect")
	if len(ds) == 0 {
		devices, _ := cts.deviceRepository.FindAll()
		for _, device := range devices {
			if worker, ok := cts.collectorServer.FindByCollectorName(device.Collector.Name); ok {
				if worker.CollectTaskIsFull() {
					zap.S().Errorf("采集队列【%s】已满, 忽略掉本次周期对设备【%s】的采集任务。请适当调整采集周期！", device.Collector.Name, device.Name)
					continue
				}

				worker.CollectTask(device)
			}
		}
	} else {
		for _, d := range ds {
			device, _ := cts.deviceRepository.Find(d)
			if worker, ok := cts.collectorServer.FindByCollectorName(device.Collector.Name); ok {
				if worker.CollectTaskIsFull() {
					zap.S().Errorf("采集队列【%s】已满, 忽略掉本次周期对设备【%s】的采集任务。请适当调整采集周期！", device.Collector.Name, device.Name)
					continue
				}
				worker.CollectTask(device)
			}
		}
	}
	zap.S().Info("采集结束：CollectTaskServer->Collect")
	cts.event.Publish(event.Event{
		Name: "CollectTask_CollectFinish",
		Data: ds,
	})
}
