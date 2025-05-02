package ntp

import (
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os/exec"
	"strings"
)

type NtpConfig struct {
	Servers []string
}

type NtpService struct {
	logger *zap.Logger
	config *NtpConfig
}

func NewNtpService(log *zap.Logger) *NtpService {
	return &NtpService{
		logger: log,
		config: &NtpConfig{
			Servers: []string{"cn.ntp.org.cn"},
		},
	}
}

func (s *NtpService) detectSystem() (string, error) {
	out, err := exec.Command("cat", "/etc/os-release").CombinedOutput()
	if err != nil {
		return "", err
	}

	if strings.Contains(string(out), "debian") {
		return "debian", nil
	}
	if strings.Contains(string(out), "centos") {
		return "centos", nil
	}
	return "", fmt.Errorf("unsupported system")
}

func (s *NtpService) SyncTime() error {
	sysType, err := s.detectSystem()
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	switch sysType {
	case "debian":
		cmd = exec.Command("timedatectl", "set-ntp", "true")
	case "centos":
		cmd = exec.Command("chronyc", "sources")
	default:
		return fmt.Errorf("unsupported system type")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("时间同步失败",
			zap.String("系统类型", sysType),
			zap.String("输出", string(output)),
			zap.Error(err))
		return err
	}

	s.logger.Info("时间同步成功",
		zap.String("系统类型", sysType),
		zap.String("服务器", strings.Join(s.config.Servers, ",")))
	return nil
}

var Module = fx.Provide(
	NewNtpService,
)
