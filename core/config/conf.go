package config

import "go.uber.org/fx"

type Config struct {
	Server  Server  `yaml:"server"`
	Cloud   Cloud   `yaml:"cloud"`
	SubMqtt SubMqtt `yaml:"subMqtt"`
}

type SubMqtt struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Server struct {
	APPID string `yaml:"appid"`
	Port  string `yaml:"port"`
}

type Cloud struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func NewConfig() *Config {
	return &Config{
		Server: Server{
			APPID: "tinyGW001",
			Port:  "8080",
		},
		Cloud: Cloud{
			Host:     "mqtt.zsxakj.com",
			Port:     "1883",
			User:     "bgscs",
			Password: "123456",
		},
		SubMqtt: SubMqtt{
			Host:     "mqtt.zsxakj.com",
			Port:     "1883",
			User:     "tinyGW001",
			Password: "123456",
		},
	}
}

var Module = fx.Provide(
	NewConfig,
)
