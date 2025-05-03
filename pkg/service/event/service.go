package event

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"runtime"
	"sync"
)

type Event struct {
	Name string
	Data interface{}
}

type Subscription func(e Event)

type EventService struct {
	subs        map[string][]Subscription
	mu          sync.RWMutex
	logger      *zap.Logger
	eventChan   chan *Event
	pool        sync.Pool
	workerCount int
}

func NewEventService(log *zap.Logger) *EventService {
	es := &EventService{
		subs:        make(map[string][]Subscription),
		logger:      log,
		eventChan:   make(chan *Event, 1000),
		workerCount: runtime.NumCPU() * 2,
		pool: sync.Pool{
			New: func() interface{} { return &Event{} },
		},
	}

	for i := 0; i < es.workerCount; i++ {
		go es.worker()
	}
	return es
}

func (s *EventService) Subscribe(eventName string, fn Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs[eventName] = append(s.subs[eventName], fn)
	s.logger.Debug("New event subscription",
		zap.String("event", eventName),
		zap.Int("subscribers", len(s.subs[eventName])))
}

func (s *EventService) Publish(e Event) {
	s.logger.Info("Publishing event", zap.String("name", e.Name))

	// 从对象池获取并初始化事件
	event := s.pool.Get().(*Event)
	event.Name = e.Name
	event.Data = e.Data

	select {
	case s.eventChan <- event:
	default:
		s.pool.Put(event)
		s.logger.Warn("Event channel full, discarding event",
			zap.String("name", e.Name),
			zap.Int("channel_size", len(s.eventChan)))
	}
}

func (s *EventService) worker() {
	for e := range s.eventChan {
		s.mu.RLock()
		subscribers, ok := s.subs[e.Name]
		s.mu.RUnlock()

		if ok {
			for _, sub := range subscribers {
				func() {
					defer func() {
						if err := recover(); err != nil {
							s.logger.Error("Event processing panic",
								zap.String("event", e.Name),
								zap.Any("error", err),
								zap.Stack("stack"))
						}
					}()
					sub(*e)
				}()
			}
		}

		// 重置并放回对象池
		e.Name = ""
		e.Data = nil
		s.pool.Put(e)
	}
}

var Module = fx.Provide(
	NewEventService,
)
