package event

import (
	"sync"
)

// EventHandler 是事件回调函数的泛型类型。
type EventHandler[T any] func(data T)

// eventEntry 保存每个事件的监听器。
type eventEntry[T any] struct {
	handlers map[uintptr]EventHandler[T]
	mu       sync.RWMutex
	nextID   uintptr
}

// EventBus 是全局事件总线。
type EventBus struct {
	entries map[string]any // key: event name, value: *eventEntry[T]
	mu      sync.RWMutex
}

// NewEventBus 创建一个新的事件总线实例。
func NewEventBus() *EventBus {
	return &EventBus{
		entries: make(map[string]any),
	}
}

// Register 注册一个事件监听器，返回监听器ID（用于注销）。
func Register[T any](bus *EventBus, event string, handler EventHandler[T]) uintptr {
	bus.mu.Lock()
	entryAny, ok := bus.entries[event]
	if !ok {
		entry := &eventEntry[T]{handlers: make(map[uintptr]EventHandler[T])}
		bus.entries[event] = entry
		bus.mu.Unlock()
		return entry.add(handler)
	}
	bus.mu.Unlock()
	entry, ok := entryAny.(*eventEntry[T])
	if !ok {
		panic("event type mismatch for event: " + event)
	}
	return entry.add(handler)
}

// Unregister 注销一个事件监听器。
func Unregister[T any](bus *EventBus, event string, id uintptr) {
	bus.mu.RLock()
	entryAny, ok := bus.entries[event]
	bus.mu.RUnlock()
	if !ok {
		return
	}
	entry, ok := entryAny.(*eventEntry[T])
	if !ok {
		panic("event type mismatch for event: " + event)
	}
	entry.remove(id)
}

// Emit 触发事件，通知所有监听器。
func Emit[T any](bus *EventBus, event string, data T) {
	bus.mu.RLock()
	entryAny, ok := bus.entries[event]
	bus.mu.RUnlock()
	if !ok {
		return
	}
	entry, ok := entryAny.(*eventEntry[T])
	if !ok {
		panic("event type mismatch for event: " + event)
	}
	entry.emit(data)
}

// --- eventEntry 方法 ---

func (e *eventEntry[T]) add(handler EventHandler[T]) uintptr {
	e.mu.Lock()
	id := e.nextID
	e.nextID++
	e.handlers[id] = handler
	e.mu.Unlock()
	return id
}

func (e *eventEntry[T]) remove(id uintptr) {
	e.mu.Lock()
	delete(e.handlers, id)
	e.mu.Unlock()
}

func (e *eventEntry[T]) emit(data T) {
	e.mu.RLock()
	handlers := make([]EventHandler[T], 0, len(e.handlers))
	for _, h := range e.handlers {
		handlers = append(handlers, h)
	}
	e.mu.RUnlock()
	for _, h := range handlers {
		h(data)
	}
}
