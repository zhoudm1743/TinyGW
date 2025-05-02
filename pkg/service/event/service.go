package event

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"sync"
	"tinyGW/pkg/service/conf"
)

type Event struct {
	Name string
	Data interface{}
}

type Subscription func(e Event)

type EventService struct {
	subs   map[string][]Subscription
	mu     sync.RWMutex
	logger *zap.Logger
}

func NewEventService(config *conf.Config, log *zap.Logger) *EventService {
	return &EventService{
		subs:   make(map[string][]Subscription),
		logger: log,
	}
}

func (s *EventService) Subscribe(eventName string, fn Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs[eventName] = append(s.subs[eventName], fn)
	s.logger.Debug("注册新事件订阅", zap.String("事件名称", eventName))
}

func (s *EventService) Publish(e Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.logger.Info("发布事件", zap.String("名称", e.Name))

	if subscribers, ok := s.subs[e.Name]; ok {
		for _, sub := range subscribers {
			go func(fn Subscription) {
				defer func() {
					if err := recover(); err != nil {
						s.logger.Error("事件处理异常",
							zap.String("事件", e.Name),
							zap.Any("错误", err))
					}
				}()
				fn(e)
			}(sub)
		}
	}
}

var Module = fx.Provide(
	NewEventService,
)
