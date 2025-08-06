package controller

import (
	"TinyGW/internal/repo"
	"TinyGW/internal/schemas/req"
	"TinyGW/internal/schemas/resp"
	"TinyGW/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeviceController struct {
	deviceRepo *repo.DeviceRepo
}

func NewDeviceController(deviceRepo *repo.DeviceRepo) *DeviceController {
	return &DeviceController{
		deviceRepo: deviceRepo,
	}
}

// Create 创建设备
func (c *DeviceController) Create(ctx *gin.Context) {
	var deviceReq req.DeviceReq
	if err := ctx.ShouldBindJSON(&deviceReq); err != nil {
		FailWithParamError(ctx, err.Error())
		return
	}

	device := &models.Device{
		Name:        deviceReq.Name,
		TypeID:      deviceReq.TypeID,
		Address:     deviceReq.Address,
		CollectorID: deviceReq.CollectorID,
		InitialVal:  deviceReq.InitialVal,
		Scale:       deviceReq.Scale,
	}

	if err := c.deviceRepo.Create(device); err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	SuccessWithMsg(ctx, "设备创建成功", nil)
}

// Update 更新设备
func (c *DeviceController) Update(ctx *gin.Context) {
	var deviceReq req.DeviceReq
	if err := ctx.ShouldBindJSON(&deviceReq); err != nil {
		FailWithParamError(ctx, err.Error())
		return
	}

	device, err := c.deviceRepo.Get(deviceReq.Name)
	if err != nil {
		FailWithNotFound(ctx, "设备不存在")
		return
	}

	device.TypeID = deviceReq.TypeID
	device.Address = deviceReq.Address
	device.CollectorID = deviceReq.CollectorID
	device.InitialVal = deviceReq.InitialVal
	device.Scale = deviceReq.Scale

	if err := c.deviceRepo.Update(device); err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	SuccessWithMsg(ctx, "设备更新成功", nil)
}

// Delete 删除设备
func (c *DeviceController) Delete(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		FailWithParamError(ctx, "设备名不能为空")
		return
	}

	device, err := c.deviceRepo.Get(name)
	if err != nil {
		FailWithNotFound(ctx, "设备不存在")
		return
	}

	if err := c.deviceRepo.Delete(device); err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	SuccessWithMsg(ctx, "设备删除成功", nil)
}

// Get 获取设备
func (c *DeviceController) Get(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		FailWithParamError(ctx, "设备名不能为空")
		return
	}

	device, err := c.deviceRepo.Get(name)
	if err != nil {
		FailWithNotFound(ctx, "设备不存在")
		return
	}

	deviceResp := resp.DeviceResp{
		Name:           device.Name,
		TypeID:         device.TypeID,
		Address:        device.Address,
		CollectorID:    device.CollectorID,
		Online:         device.Online,
		CollectTime:    device.CollectTime,
		CollectTotal:   device.CollectTotal,
		CollectSuccess: device.CollectSuccess,
		ReportTime:     device.ReportTime,
		ReportTotal:    device.ReportTotal,
		ReportSuccess:  device.ReportSuccess,
		AlarmStatus:    device.AlarmStatus,
		AlarmReason:    device.AlarmReason,
		AlarmTime:      device.AlarmTime,
		AlarmTotal:     device.AlarmTotal,
		InitialVal:     device.InitialVal,
		Scale:          device.Scale,
		CreatedAt:      device.CreatedAt,
		UpdatedAt:      device.UpdatedAt,
	}

	Success(ctx, deviceResp)
}

// List 获取设备列表
func (c *DeviceController) List(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	conds := make(map[string]interface{})

	// 可选过滤条件
	if typeID := ctx.Query("typeID"); typeID != "" {
		conds["type_id = ?"] = typeID
	}

	if collectorID := ctx.Query("collectorID"); collectorID != "" {
		conds["collector_id = ?"] = collectorID
	}

	total, devices, err := c.deviceRepo.Find(page, pageSize, conds)
	if err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	var deviceResps []resp.DeviceResp
	for _, device := range devices {
		deviceResps = append(deviceResps, resp.DeviceResp{
			Name:           device.Name,
			TypeID:         device.TypeID,
			Address:        device.Address,
			CollectorID:    device.CollectorID,
			Online:         device.Online,
			CollectTime:    device.CollectTime,
			CollectTotal:   device.CollectTotal,
			CollectSuccess: device.CollectSuccess,
			ReportTime:     device.ReportTime,
			ReportTotal:    device.ReportTotal,
			ReportSuccess:  device.ReportSuccess,
			AlarmStatus:    device.AlarmStatus,
			AlarmReason:    device.AlarmReason,
			AlarmTime:      device.AlarmTime,
			AlarmTotal:     device.AlarmTotal,
			InitialVal:     device.InitialVal,
			Scale:          device.Scale,
			CreatedAt:      device.CreatedAt,
			UpdatedAt:      device.UpdatedAt,
		})
	}

	SuccessWithPage(ctx, total, page, pageSize, deviceResps)
}

// Count 获取设备数量
func (c *DeviceController) Count(ctx *gin.Context) {
	conds := make(map[string]interface{})

	// 可选过滤条件
	if typeID := ctx.Query("typeID"); typeID != "" {
		conds["type_id = ?"] = typeID
	}

	if collectorID := ctx.Query("collectorID"); collectorID != "" {
		conds["collector_id = ?"] = collectorID
	}

	count, err := c.deviceRepo.Count(conds)
	if err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	Success(ctx, gin.H{"count": count})
}
