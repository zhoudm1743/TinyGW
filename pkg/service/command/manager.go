package command

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	commandManager *Manager
	once           sync.Once
)

// Manager 命令管理器
type Manager struct {
	// pendingCommands 待发送指令映射表: 设备地址 -> 指令数据
	pendingCommands     map[string]*PendingCommand
	pendingCommandMutex sync.RWMutex
}

// PendingCommand 待发送命令结构
type PendingCommand struct {
	DeviceAddr  string    // 设备地址
	CommandData []byte    // 指令数据
	CreateTime  time.Time // 创建时间
	ExpireTime  time.Time // 过期时间
}

// GetManager 获取命令管理器单例
func GetManager() *Manager {
	once.Do(func() {
		commandManager = &Manager{
			pendingCommands: make(map[string]*PendingCommand),
		}
		// 启动定期清理任务
		go commandManager.startPeriodicCleanup()
	})
	return commandManager
}

// Store 存储待发送的指令
// deviceAddr: 设备地址
// commandData: 指令数据
// expireSeconds: 过期时间(秒)，指令在该时间后将被视为过期，默认24小时
func (m *Manager) Store(deviceAddr string, commandData []byte, expireSeconds int64) {
	if expireSeconds <= 0 {
		expireSeconds = 86400 // 默认24小时
	}

	m.pendingCommandMutex.Lock()
	defer m.pendingCommandMutex.Unlock()

	m.pendingCommands[deviceAddr] = &PendingCommand{
		DeviceAddr:  deviceAddr,
		CommandData: commandData,
		CreateTime:  time.Now(),
		ExpireTime:  time.Now().Add(time.Duration(expireSeconds) * time.Second),
	}

	zap.S().Infof("存储设备 %s 的待发送指令，过期时间: %v", deviceAddr, time.Now().Add(time.Duration(expireSeconds)*time.Second))
}

// Get 获取设备待发送的指令
// 如果存在未过期的指令，返回指令数据和true
// 如果不存在或已过期，返回nil和false
func (m *Manager) Get(deviceAddr string) ([]byte, bool) {
	m.pendingCommandMutex.RLock()
	defer m.pendingCommandMutex.RUnlock()

	cmd, exists := m.pendingCommands[deviceAddr]
	if !exists {
		return nil, false
	}

	// 检查指令是否过期
	if time.Now().After(cmd.ExpireTime) {
		zap.S().Warnf("设备 %s 的指令已过期，创建时间: %v, 过期时间: %v",
			deviceAddr, cmd.CreateTime, cmd.ExpireTime)
		return nil, false
	}

	return cmd.CommandData, true
}

// Remove 移除设备的待发送指令
func (m *Manager) Remove(deviceAddr string) {
	m.pendingCommandMutex.Lock()
	defer m.pendingCommandMutex.Unlock()

	if _, exists := m.pendingCommands[deviceAddr]; exists {
		delete(m.pendingCommands, deviceAddr)
		zap.S().Infof("移除设备 %s 的待发送指令", deviceAddr)
	}
}

// CleanupExpired 清理所有过期的指令
func (m *Manager) CleanupExpired() {
	m.pendingCommandMutex.Lock()
	defer m.pendingCommandMutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for addr, cmd := range m.pendingCommands {
		if now.After(cmd.ExpireTime) {
			delete(m.pendingCommands, addr)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		zap.S().Infof("清理了 %d 条过期指令", expiredCount)
	}
}

// startPeriodicCleanup 启动定期清理任务
func (m *Manager) startPeriodicCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		m.CleanupExpired()
	}
}

// 以下是提供给外部包使用的便捷函数

// Store 存储命令（便捷函数）
func Store(deviceAddr string, commandData []byte, expireSeconds int64) {
	GetManager().Store(deviceAddr, commandData, expireSeconds)
}

// Get 获取命令（便捷函数）
func Get(deviceAddr string) ([]byte, bool) {
	return GetManager().Get(deviceAddr)
}

// Remove 移除命令（便捷函数）
func Remove(deviceAddr string) {
	GetManager().Remove(deviceAddr)
}
