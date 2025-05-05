package collect

import (
	"go.uber.org/zap"
	"sync"
	"tinyGW/app/api/repository"
	"tinyGW/app/models"
	"tinyGW/pkg/service/collect/worker"
	"tinyGW/pkg/service/conf"
	"tinyGW/pkg/service/event"
)

type (
	// CollectorServer 采集接口服务器
	CollectorServer struct {
		collectors       *sync.Map
		deviceRepository repository.DeviceRepository
		config           *conf.Config
		eventSrv         *event.EventService
	}
)

// NewCollectorServer 实例化
func NewCollectorServer(
	deviceRepository repository.DeviceRepository,
	config *conf.Config,
	e *event.EventService,
) *CollectorServer {
	return &CollectorServer{
		collectors:       &sync.Map{},
		deviceRepository: deviceRepository,
		config:           config,
		eventSrv:         e,
	}
}

// InitCollectorServer 初始化
func InitCollectorServer(server *CollectorServer, repository repository.CollectorRepository, e *event.EventService) {
	zap.S().Info("初始化采集接口服务器")
	if collectors, err := repository.FindAll(); err == nil {
		for _, collector := range collectors {
			server.Add(collector)
		}
	}
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
	zap.S().Info("新增采集接口服务器", collector)
	w := worker.NewWorker(
		collector, cs.deviceRepository,
		cs.config, cs.eventSrv,
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
