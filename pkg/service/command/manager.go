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
	// pendingCommands 待发送指令映射表: 设备地址 -> 指令队列
	pendingCommands     map[string][]*PendingCommand
	pendingCommandMutex sync.RWMutex
}

// PendingCommand 待发送命令结构
type PendingCommand struct {
	DeviceAddr  string    // 设备地址
	CommandData []byte    // 指令数据
	CreateTime  time.Time // 创建时间
	ExpireTime  time.Time // 过期时间
	CommandID   string    // 指令ID，可选，用于标识特定指令
}

// GetManager 获取命令管理器单例
func GetManager() *Manager {
	once.Do(func() {
		commandManager = &Manager{
			pendingCommands: make(map[string][]*PendingCommand),
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

	cmd := &PendingCommand{
		DeviceAddr:  deviceAddr,
		CommandData: commandData,
		CreateTime:  time.Now(),
		ExpireTime:  time.Now().Add(time.Duration(expireSeconds) * time.Second),
	}

	// 如果该设备地址不存在队列，创建新队列
	if _, exists := m.pendingCommands[deviceAddr]; !exists {
		m.pendingCommands[deviceAddr] = make([]*PendingCommand, 0)
	}

	// 将指令添加到队列末尾
	m.pendingCommands[deviceAddr] = append(m.pendingCommands[deviceAddr], cmd)

	zap.S().Infof("存储设备 %s 的待发送指令，队列长度: %d，过期时间: %v",
		deviceAddr, len(m.pendingCommands[deviceAddr]), cmd.ExpireTime)
}

// Get 获取设备待发送的指令（返回最早入队的指令）
// 如果存在未过期的指令，返回指令数据和true
// 如果不存在或已过期，返回nil和false
func (m *Manager) Get(deviceAddr string) ([]byte, bool) {
	m.pendingCommandMutex.Lock()
	defer m.pendingCommandMutex.Unlock()

	cmds, exists := m.pendingCommands[deviceAddr]
	if !exists || len(cmds) == 0 {
		return nil, false
	}

	// 获取队列中第一个指令
	cmd := cmds[0]

	// 检查指令是否过期
	if time.Now().After(cmd.ExpireTime) {
		// 移除过期指令
		m.pendingCommands[deviceAddr] = cmds[1:]
		zap.S().Warnf("设备 %s 的指令已过期，创建时间: %v, 过期时间: %v",
			deviceAddr, cmd.CreateTime, cmd.ExpireTime)

		// 如果队列为空，删除该设备的队列
		if len(m.pendingCommands[deviceAddr]) == 0 {
			delete(m.pendingCommands, deviceAddr)
		}

		// 递归调用获取下一个有效指令
		return m.Get(deviceAddr)
	}

	// 返回指令数据，但不从队列中移除
	// 指令将在调用Remove后移除
	return cmd.CommandData, true
}

// Remove 移除设备的最早待发送指令
func (m *Manager) Remove(deviceAddr string) {
	m.pendingCommandMutex.Lock()
	defer m.pendingCommandMutex.Unlock()

	cmds, exists := m.pendingCommands[deviceAddr]
	if !exists || len(cmds) == 0 {
		return
	}

	// 移除队列中第一个指令
	m.pendingCommands[deviceAddr] = cmds[1:]
	zap.S().Infof("移除设备 %s 的待发送指令，剩余队列长度: %d", deviceAddr, len(m.pendingCommands[deviceAddr]))

	// 如果队列为空，删除该设备的队列
	if len(m.pendingCommands[deviceAddr]) == 0 {
		delete(m.pendingCommands, deviceAddr)
	}
}

// GetQueueLength 获取设备指令队列长度
func (m *Manager) GetQueueLength(deviceAddr string) int {
	m.pendingCommandMutex.RLock()
	defer m.pendingCommandMutex.RUnlock()

	cmds, exists := m.pendingCommands[deviceAddr]
	if !exists {
		return 0
	}
	return len(cmds)
}

// HasPendingCommands 检查设备是否有待发送指令
func (m *Manager) HasPendingCommands(deviceAddr string) bool {
	return m.GetQueueLength(deviceAddr) > 0
}

// CleanupExpired 清理所有过期的指令
func (m *Manager) CleanupExpired() {
	m.pendingCommandMutex.Lock()
	defer m.pendingCommandMutex.Unlock()

	now := time.Now()
	expiredCount := 0
	devicesToClean := []string{}

	// 对每个设备的命令队列进行清理
	for addr, cmds := range m.pendingCommands {
		validCmds := make([]*PendingCommand, 0)

		// 过滤出未过期的指令
		for _, cmd := range cmds {
			if !now.After(cmd.ExpireTime) {
				validCmds = append(validCmds, cmd)
			} else {
				expiredCount++
			}
		}

		if len(validCmds) > 0 {
			// 更新为过滤后的队列
			m.pendingCommands[addr] = validCmds
		} else {
			// 如果没有有效指令，标记为待清理
			devicesToClean = append(devicesToClean, addr)
		}
	}

	// 清理没有有效指令的设备
	for _, addr := range devicesToClean {
		delete(m.pendingCommands, addr)
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

// GetQueueLength 获取队列长度（便捷函数）
func GetQueueLength(deviceAddr string) int {
	return GetManager().GetQueueLength(deviceAddr)
}

// HasPendingCommands 检查是否有待发送指令（便捷函数）
func HasPendingCommands(deviceAddr string) bool {
	return GetManager().HasPendingCommands(deviceAddr)
}
