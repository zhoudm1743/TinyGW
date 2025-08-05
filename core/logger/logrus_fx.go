package logger

import (
	"github.com/sirupsen/logrus"
	"go.uber.org/fx/fxevent"
)

type FxLogrusLogger struct {
	Log *logrus.Logger
}

func (l *FxLogrusLogger) LogEvent(ev fxevent.Event) {
	switch e := ev.(type) {
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.Log.Errorf("[fx] 启动失败: %s: %v", e.FunctionName, e.Err)
		}
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.Log.Errorf("[fx] 停止失败: %s: %v", e.FunctionName, e.Err)
		}
	case *fxevent.Run:
		if e.Err != nil {
			l.Log.Errorf("[fx] 运行出错: %v", e.Err)
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.Log.Errorf("[fx] 日志初始化失败: %v", e.Err)
		}
	case *fxevent.RollingBack:
		l.Log.Warnf("[fx] 回滚: %v", e.StartErr)
	}
	// 其他事件类型不输出
}

func NewFxLogger(log *logrus.Logger) fxevent.Logger {
	return &FxLogrusLogger{Log: log}
}
