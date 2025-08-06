package controller

import (
	"TinyGW/internal/repo"
	"TinyGW/internal/schemas/req"
	"TinyGW/internal/schemas/resp"
	"TinyGW/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type DeviceLogController struct {
	deviceLogRepo *repo.DeviceLogRepo
	deviceRepo    *repo.DeviceRepo
}

func NewDeviceLogController(deviceLogRepo *repo.DeviceLogRepo, deviceRepo *repo.DeviceRepo) *DeviceLogController {
	return &DeviceLogController{
		deviceLogRepo: deviceLogRepo,
		deviceRepo:    deviceRepo,
	}
}

// Create 创建设备日志
func (c *DeviceLogController) Create(ctx *gin.Context) {
	var logReq req.DeviceLogReq
	if err := ctx.ShouldBindJSON(&logReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证设备是否存在
	device, err := c.deviceRepo.Get(logReq.DeviceID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "设备不存在"})
		return
	}

	log := &models.DeviceLog{
		DeviceID:  logReq.DeviceID,
		Device:    device,
		Time:      logReq.Time,
		OriginCmd: logReq.OriginCmd,
		OriginRes: logReq.OriginRes,
		Res:       logReq.Res,
		Property:  logReq.Property,
		Value:     logReq.Value,
		Status:    logReq.Status,
		Error:     logReq.Error,
	}

	if err := c.deviceLogRepo.Create(log); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "设备日志创建成功"})
}

// Delete 删除设备日志
func (c *DeviceLogController) Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的日志ID"})
		return
	}

	log := &models.DeviceLog{
		Model: models.Model{
			ID: uint(id),
		},
	}

	if err := c.deviceLogRepo.Delete(log); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "设备日志删除成功"})
}

// List 获取设备日志列表
func (c *DeviceLogController) List(ctx *gin.Context) {
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
	if deviceID := ctx.Query("deviceID"); deviceID != "" {
		conds["device_id = ?"] = deviceID
	}

	if status := ctx.Query("status"); status != "" {
		statusInt, err := strconv.Atoi(status)
		if err == nil {
			conds["status = ?"] = statusInt
		}
	}

	if property := ctx.Query("property"); property != "" {
		conds["property = ?"] = property
	}

	if startTimeStr := ctx.Query("startTime"); startTimeStr != "" {
		startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err == nil {
			conds["time >= ?"] = startTime
		}
	}

	if endTimeStr := ctx.Query("endTime"); endTimeStr != "" {
		endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err == nil {
			conds["time <= ?"] = endTime
		}
	}

	total, logs, err := c.deviceLogRepo.Find(page, pageSize, conds)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var logResps []resp.DeviceLogResp
	for _, log := range logs {
		var deviceName string
		if log.Device != nil {
			deviceName = log.Device.Name
		}

		logResps = append(logResps, resp.DeviceLogResp{
			ID:         log.ID,
			DeviceID:   log.DeviceID,
			DeviceName: deviceName,
			Time:       log.Time,
			OriginCmd:  log.OriginCmd,
			OriginRes:  log.OriginRes,
			Res:        log.Res,
			Property:   log.Property,
			Value:      log.Value,
			Status:     log.Status,
			Error:      log.Error,
			CreatedAt:  log.CreatedAt,
			UpdatedAt:  log.UpdatedAt,
		})
	}

	ctx.JSON(http.StatusOK, resp.PageResp{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     logResps,
	})
}

// Count 获取设备日志数量
func (c *DeviceLogController) Count(ctx *gin.Context) {
	conds := make(map[string]interface{})

	// 可选过滤条件
	if deviceID := ctx.Query("deviceID"); deviceID != "" {
		conds["device_id = ?"] = deviceID
	}

	if status := ctx.Query("status"); status != "" {
		statusInt, err := strconv.Atoi(status)
		if err == nil {
			conds["status = ?"] = statusInt
		}
	}

	count, err := c.deviceLogRepo.Count(conds)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// ClearBefore 清除指定时间之前的日志
func (c *DeviceLogController) ClearBefore(ctx *gin.Context) {
	timeStr := ctx.Query("time")
	if timeStr == "" {
		// 默认清除7天前的日志
		timeStr = strconv.FormatInt(time.Now().AddDate(0, 0, -7).Unix(), 10)
	}

	clearTime, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的时间参数"})
		return
	}

	if err := c.deviceLogRepo.ClearBefore(clearTime); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "日志清除成功"})
}
