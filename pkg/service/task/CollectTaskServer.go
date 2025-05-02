package task

import (
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"sync"
	"zsxagw/api/domain"
	"zsxagw/api/repository"
	"zsxagw/core/collect"
)

type (
	// CollectTaskServer 采集任务服务器
	CollectTaskServer struct {
		tasks            *sync.Map
		collectorServer  *collect.CollectorServer
		deviceRepository repository.DeviceRepository
	}
)

// NewCollectTaskServer 实例化
func NewCollectTaskServer(collectorServer *collect.CollectorServer, deviceRepository repository.DeviceRepository) *CollectTaskServer {
	return &CollectTaskServer{
		tasks:            &sync.Map{},
		collectorServer:  collectorServer,
		deviceRepository: deviceRepository,
	}
}

// InitCollectTaskServer 初始化
func InitCollectTaskServer(collectTaskServer *CollectTaskServer, repository repository.CollectTaskRepository) {
	zap.S().Info("初始化采集定时任务")
	if tasks, err := repository.FindAll(); err == nil {
		for _, task := range tasks {
			collectTaskServer.Add(&task)
		}
	}
}

// Add 新增定时任务，如果定时任务是开启状态，则开启定时任务
func (cts *CollectTaskServer) Add(task *domain.CollectTask) {
	zap.S().Info("新增定时任务", task)
	c := cron.New()
	c.AddFunc(task.Cron, cts.Collect)
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
func (cts *CollectTaskServer) Update(task *domain.CollectTask) {
	cts.Delete(task.Name)
	cts.Add(task)
}

// Start 启动定时任务
func (cts *CollectTaskServer) Start(task *domain.CollectTask) {
	zap.S().Info("启动定时任务", task)
	if value, loaded := cts.tasks.Load(task.Name); loaded {
		c := value.(*cron.Cron)
		c.Start()
	}
}

// Stop 停止定时任务
func (cts *CollectTaskServer) Stop(task *domain.CollectTask) {
	zap.S().Info("停止定时任务", task)
	if value, loaded := cts.tasks.Load(task.Name); loaded {
		c := value.(*cron.Cron)
		c.Stop()
	}
}

// Collect 数据采集
func (cts *CollectTaskServer) Collect() {
	zap.S().Info("采集开始：CollectTaskServer->Collect")
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
	zap.S().Info("采集结束：CollectTaskServer->Collect")
}
