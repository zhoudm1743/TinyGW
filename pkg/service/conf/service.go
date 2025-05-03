package conf

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

var Version = "v20250503"

func NewConfig() *Config {
	defaultConfig := &Config{
		Server: Server{
			Port: 8080,
			Host: "0.0.0.0",
			Mode: "debug",
		},
		Cloud: Cloud{
			Host:     "mqtt.zsxakj.com",
			Port:     1883,
			Username: "energy-cloud",
			Password: "xakj1PnGD8y#s$",
			ClientId: "zsxagwB001",
		},
		Serial: Serial{
			OpenEveryTime: false,
		},
		Logger: Logger{
			Level:      "debug",
			Filename:   "logs/energy.log",
			MaxSize:    10,
			MaxBackups: 10,
			MaxAge:     3,
		},
	}
	conf := &Config{}

	viper.SetConfigType(configType)
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return defaultConfig
	}

	if err := viper.Unmarshal(conf); err != nil {
		return defaultConfig
	}

	return conf
}

func SaveConfig(config *Config) {
	//fmt.Println(config)
	viper.SetConfigType(configType)
	viper.SetConfigFile(configFile)
	viper.Set("cloud.ip", config.Cloud.Host)
	viper.Set("cloud.clientId", config.Cloud.ClientId)
	viper.Set("cloud.port", config.Cloud.Port)

	err := viper.WriteConfig()

	if err != nil {
		fmt.Println(err.Error())
	}
}

var Module = fx.Provide(
	NewConfig,
)
