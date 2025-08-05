package collect

import (
	"runtime"
	"sync"
	"tinyGW/app/api/repository"
	"tinyGW/app/models"
	"tinyGW/pkg/service/collect/worker"
	"tinyGW/pkg/service/command"
	"tinyGW/pkg/service/conf"
	"tinyGW/pkg/service/event"

	"go.uber.org/zap"
)

type (
	// CollectorServer 采集接口服务器
	CollectorServer struct {
		collectors           *sync.Map
		deviceRepository     repository.DeviceRepository
		deviceTypeRepository repository.DeviceTypeRepository
		config               *conf.Config
		eventSrv             *event.EventService
		workerLimiter        chan struct{} // 工作线程限制器
		maxWorkers           int           // 最大同时工作线程数
		workerCount          int32         // 当前工作线程计数
		commandManager       *command.Manager
	}
)

// NewCollectorServer 实例化
func NewCollectorServer(
	deviceRepository repository.DeviceRepository,
	config *conf.Config,
	e *event.EventService,
	deviceTypeRepository repository.DeviceTypeRepository,
	commandManager *command.Manager,
) *CollectorServer {
	// 根据CPU核心数和配置决定最大工作线程数
	maxWorkers := runtime.NumCPU() * 2
	if config.Server.MaxWorkers > 0 {
		maxWorkers = config.Server.MaxWorkers
	}

	zap.S().Infof("初始化采集服务器，最大工作线程: %d", maxWorkers)

	return &CollectorServer{
		collectors:           &sync.Map{},
		deviceRepository:     deviceRepository,
		deviceTypeRepository: deviceTypeRepository,
		config:               config,
		eventSrv:             e,
		workerLimiter:        make(chan struct{}, maxWorkers), // 限制最大并发工作线程
		maxWorkers:           maxWorkers,
		commandManager:       commandManager,
	}
}

// InitCollectorServer 初始化
func InitCollectorServer(server *CollectorServer, repository repository.CollectorRepository, e *event.EventService) {
	zap.S().Info("初始化采集接口服务器")

	// 批量加载所有采集器，而不是一个个加载
	if collectors, err := repository.FindAll(); err == nil {
		// 预先统计活跃的采集器数量
		activeCount := 0
		for _, _ = range collectors {
			activeCount++
			// if collector.Enable {
			// 	activeCount++
			// }
		}

		zap.S().Infof("总采集器数量: %d, 活跃采集器: %d", len(collectors), activeCount)

		// 只处理启用的采集器
		for _, collector := range collectors {
			server.Add(collector)
			// if collector.Enable {
			// 	server.Add(collector)
			// } else {
			// 	zap.S().Infof("采集器 %s 已禁用，跳过加载", collector.Name)
			// }
		}
	}

	// 注册事件处理
	e.Subscribe("collector_add", func(e event.Event) {
		server.Add(e.Data.(models.Collector))
	})
	e.Subscribe("collector_update", func(e event.Event) {
		server.Update(e.Data.(models.Collector))
	})
	e.Subscribe("collector_delete", func(e event.Event) {
		server.Delete(e.Data.(string))
	})
}

// Add 新增采集接口服务器
func (cs *CollectorServer) Add(collector models.Collector) {
	if len(collector.Name) == 0 {
		zap.S().Error("新增采集接口服务器失败，名称不能为空")
		return
	}

	// 只加载启用的采集器
	// if !collector.Enable {
	// 	zap.S().Infof("采集器 %s 已禁用，不加载", collector.Name)
	// 	return
	// }

	zap.S().Infof("新增采集接口服务器: %s (类型: %s)", collector.Name, collector.Type)

	w := worker.NewWorker(
		collector, cs.deviceRepository,
		cs.config, cs.eventSrv,
		cs.deviceTypeRepository,
		cs.commandManager,
	)

	w.Start()
	cs.collectors.Store(collector.Name, w)
}

// Delete 删除采集接口服务器
func (cs *CollectorServer) Delete(name string) {
	zap.S().Info("删除采集接口服务器", name)
	if value, loaded := cs.collectors.LoadAndDelete(name); loaded {
		if c, ok := value.(worker.Worker); ok {
			c.Stop()
		}
	}
}

// Update 修改采集接口服务器
func (cs *CollectorServer) Update(collector models.Collector) {
	cs.Delete(collector.Name)
	cs.Add(collector)
	// if collector.Enable {
	// 	cs.Add(collector)
	// } else {
	// 	zap.S().Infof("采集器 %s 已更新为禁用状态，不重新加载", collector.Name)
	// }
}

func (cs *CollectorServer) FindByCollectorName(collectorName string) (worker.Worker, bool) {
	if value, ok := cs.collectors.Load(collectorName); ok {
		if work, ok := value.(worker.Worker); ok {
			return work, ok
		}
	}

	return nil, false
}

func (cs *CollectorServer) FindByDeviceName(deviceName string) (worker.Worker, bool) {
	device, err := cs.deviceRepository.Find(deviceName)
	if err != nil {
		return nil, false
	}
	if value, ok := cs.collectors.Load(device.Collector.Name); ok {
		if work, ok := value.(worker.Worker); ok {
			return work, ok
		}
	}

	return nil, false
}
