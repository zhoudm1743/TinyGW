# Event 事件系统

本模块基于 Go 泛型实现线程安全的全局事件系统，支持通过 fx 依赖注入在任意模块/对象中使用。

## 特性
- 支持事件注册、注销、触发
- 支持事件回调/监听器，参数类型为泛型
- 线程安全
- 可通过 fx.Module 注入为单例

## 快速开始

### 1. 在 fx.App 中注入事件总线
```go
import (
    "go.uber.org/fx"
    "yourmodule/core/event" // 替换为实际模块路径
)

var Module = fx.Options(
    event.Module, // 注入事件总线
    // ...其他模块
)
```

### 2. 在对象/模块中依赖事件总线
```go
import "yourmodule/core/event"

type MyService struct {
    bus *event.EventBus
}

func NewMyService(bus *event.EventBus) *MyService {
    return &MyService{bus: bus}
}
```

### 3. 注册、触发、注销事件（支持泛型）
```go
// 注册 string 类型事件
id := event.Register[string](bus, "test_event", func(data string) {
    fmt.Println("收到事件:", data)
})
// 触发事件
event.Emit[string](bus, "test_event", "Hello Event!")
// 注销监听器
event.Unregister[string](bus, "test_event", id)

// 注册 int 类型事件
id2 := event.Register[int](bus, "int_event", func(data int) {
    fmt.Println("int event:", data)
})
event.Emit[int](bus, "int_event", 42)
event.Unregister[int](bus, "int_event", id2)
```

### 4. 线程安全
事件系统内部已加锁，支持并发注册、触发、注销。

---

如需更多高级用法，请参考源码注释。 