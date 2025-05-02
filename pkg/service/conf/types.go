package conf

const (
	configFile = "config/conf.yml"
	configType = "yml"
)

type Config struct {
	Server    Server    `yaml:"server" json:"server"`
	Cloud     Cloud     `yaml:"cloud" json:"cloud"`
	Serial    Serial    `yaml:"serial" json:"serial"`
	Logger    Logger    `yaml:"logger" json:"logger"`
	Subscribe Subscribe `yaml:"subscribe" json:"subscribe"`
}

// Logger 配置日志文件
type Logger struct {
	Level      string `json:"level" yaml:"level"`
	Filename   string `json:"filename" yaml:"filename"`
	MaxSize    int    `json:"maxsize" yaml:"maxsize"`
	MaxBackups int    `json:"maxbackups" yaml:"maxbackups"`
	MaxAge     int    `json:"maxage" yaml:"maxage"`
}

type Serial struct {
	OpenEveryTime bool `json:"openEveryTime" yaml:"openEveryTime"`
}

type Server struct {
	Port int    `yaml:"port" json:"port"`
	Host string `yaml:"host" json:"host"`
	Mode string `yaml:"module" json:"module"`
}

type Cloud struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	ClientId string `yaml:"clientId" json:"clientId"`
}

type Subscribe struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	ClientId string `yaml:"clientId" json:"clientId"`
}
