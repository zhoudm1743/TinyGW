package boot

import (
	"context"
	"tinyGW/app/api"
	"tinyGW/pkg/plugin/io"
	"tinyGW/pkg/service/cache"
	"tinyGW/pkg/service/cloud"
	"tinyGW/pkg/service/collect"
	"tinyGW/pkg/service/conf"
	"tinyGW/pkg/service/event"
	"tinyGW/pkg/service/http"
	"tinyGW/pkg/service/listener"
	"tinyGW/pkg/service/logger"
	"tinyGW/pkg/service/ntp"
	"tinyGW/pkg/service/orm"
	"tinyGW/pkg/service/script"
	"tinyGW/pkg/service/subscription"
	"tinyGW/pkg/service/task"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Module = fx.Options(
	conf.Module,
	logger.Module,
	ntp.Module,
	cache.Module,
	event.Module,
	orm.Module,
	http.Module,
	api.Module,
	script.Module,
	collect.Module,
	task.Module,

	// 运行时间监听模块
	fx.Invoke(io.StartTime),
	// 启动订阅服务器
	fx.Invoke(subscription.InitMqttServer),
	// 启动云端客户端
	fx.Invoke(cloud.InitMqttClient),
	// 启动采集接口服务器
	fx.Invoke(collect.InitCollectorServer),
	// 启动采集任务服务器
	fx.Invoke(task.InitCollectTaskServer),
	// 取消启动上报任务服务器，由采集任务执行完成后上报
	//fx.Invoke(task.InitReportTaskServer),
	fx.Invoke(setup),
)

func setup(
	lifecycle fx.Lifecycle,
	server *http.Service,
	db *gorm.DB,
	ntpSrv *ntp.NtpService,
	eventBus *event.EventService,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				err := ntpSrv.SyncTime()
				if err != nil {
					zap.S().Errorln("NTP同步失败！", err)
				}
				zap.S().Infoln("启动Web服务器...", server.Server.Addr)
				err = server.Server.ListenAndServe()
				if err != nil {
					_ = out(db)
					return
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			zap.S().Error("停止Web服务器")
			_ = out(db)
			eventBus.Shutdown()
			return server.Server.Shutdown(ctx)
		},
	})
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			zap.S().Infoln("启动监听线程...")
			go cloud.MClient.Connect()
			go listener.Start()
			go subscription.Mclient.Connect()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			zap.S().Infoln("停止监听线程")
			cloud.MClient.Disconnect()
			listener.Stop()
			subscription.Mclient.Disconnect()
			return nil
		},
	})
}

func out(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		zap.S().Errorln("数据库实例获取失败！", err)
		return err
	}
	err = sqlDB.Close()
	if err != nil {
		zap.S().Errorln("无法关闭数据库", err)
		return err
	}
	return nil
}
