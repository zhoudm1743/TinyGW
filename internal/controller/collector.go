package controller

import (
	"TinyGW/internal/repo"
	"TinyGW/internal/schemas/req"
	"TinyGW/internal/schemas/resp"
	"TinyGW/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CollectorController struct {
	collectorRepo *repo.CollectorRepo
}

func NewCollectorController(collectorRepo *repo.CollectorRepo) *CollectorController {
	return &CollectorController{
		collectorRepo: collectorRepo,
	}
}

// Create 创建采集器
func (c *CollectorController) Create(ctx *gin.Context) {
	var collectorReq req.CollectorReq
	if err := ctx.ShouldBindJSON(&collectorReq); err != nil {
		FailWithParamError(ctx, err.Error())
		return
	}

	collector := &models.Collector{
		Name:      collectorReq.Name,
		Type:      collectorReq.Type,
		Address:   collectorReq.Address,
		Serial:    collectorReq.Serial,
		TcpClient: collectorReq.TcpClient,
		TcpServer: collectorReq.TcpServer,
		Mqtt:      collectorReq.Mqtt,
		Channel:   collectorReq.Channel,
		FourGPRS:  collectorReq.FourGPRS,
		Timeout:   collectorReq.Timeout,
		Interval:  collectorReq.Interval,
		Enable:    collectorReq.Enable,
	}

	if err := c.collectorRepo.Create(collector); err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	SuccessWithMsg(ctx, "采集器创建成功", nil)
}

// Update 更新采集器
func (c *CollectorController) Update(ctx *gin.Context) {
	var collectorReq req.CollectorReq
	if err := ctx.ShouldBindJSON(&collectorReq); err != nil {
		FailWithParamError(ctx, err.Error())
		return
	}

	collector, err := c.collectorRepo.Get(collectorReq.Name)
	if err != nil {
		FailWithNotFound(ctx, "采集器不存在")
		return
	}

	collector.Type = collectorReq.Type
	collector.Address = collectorReq.Address
	collector.Serial = collectorReq.Serial
	collector.TcpClient = collectorReq.TcpClient
	collector.TcpServer = collectorReq.TcpServer
	collector.Mqtt = collectorReq.Mqtt
	collector.Channel = collectorReq.Channel
	collector.FourGPRS = collectorReq.FourGPRS
	collector.Timeout = collectorReq.Timeout
	collector.Interval = collectorReq.Interval
	collector.Enable = collectorReq.Enable

	if err := c.collectorRepo.Update(collector); err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	SuccessWithMsg(ctx, "采集器更新成功", nil)
}

// Delete 删除采集器
func (c *CollectorController) Delete(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		FailWithParamError(ctx, "采集器名不能为空")
		return
	}

	collector, err := c.collectorRepo.Get(name)
	if err != nil {
		FailWithNotFound(ctx, "采集器不存在")
		return
	}

	if err := c.collectorRepo.Delete(collector); err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	SuccessWithMsg(ctx, "采集器删除成功", nil)
}

// Get 获取采集器
func (c *CollectorController) Get(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		FailWithParamError(ctx, "采集器名不能为空")
		return
	}

	collector, err := c.collectorRepo.Get(name)
	if err != nil {
		FailWithNotFound(ctx, "采集器不存在")
		return
	}

	collectorResp := resp.CollectorResp{
		Name:      collector.Name,
		Type:      collector.Type,
		Address:   collector.Address,
		Serial:    collector.Serial,
		TcpClient: collector.TcpClient,
		TcpServer: collector.TcpServer,
		Mqtt:      collector.Mqtt,
		Channel:   collector.Channel,
		FourGPRS:  collector.FourGPRS,
		Timeout:   collector.Timeout,
		Interval:  collector.Interval,
		Enable:    collector.Enable,
		CreatedAt: collector.CreatedAt,
		UpdatedAt: collector.UpdatedAt,
	}

	Success(ctx, collectorResp)
}

// List 获取采集器列表
func (c *CollectorController) List(ctx *gin.Context) {
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
	if collectorType := ctx.Query("type"); collectorType != "" {
		conds["type = ?"] = collectorType
	}

	if enabled := ctx.Query("enable"); enabled != "" {
		enableBool := enabled == "true"
		conds["enable = ?"] = enableBool
	}

	total, collectors, err := c.collectorRepo.Find(page, pageSize, conds)
	if err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	var collectorResps []resp.CollectorResp
	for _, collector := range collectors {
		collectorResps = append(collectorResps, resp.CollectorResp{
			Name:      collector.Name,
			Type:      collector.Type,
			Address:   collector.Address,
			Serial:    collector.Serial,
			TcpClient: collector.TcpClient,
			TcpServer: collector.TcpServer,
			Mqtt:      collector.Mqtt,
			Channel:   collector.Channel,
			FourGPRS:  collector.FourGPRS,
			Timeout:   collector.Timeout,
			Interval:  collector.Interval,
			Enable:    collector.Enable,
			CreatedAt: collector.CreatedAt,
			UpdatedAt: collector.UpdatedAt,
		})
	}

	SuccessWithPage(ctx, total, page, pageSize, collectorResps)
}

// Count 获取采集器数量
func (c *CollectorController) Count(ctx *gin.Context) {
	conds := make(map[string]interface{})

	// 可选过滤条件
	if collectorType := ctx.Query("type"); collectorType != "" {
		conds["type = ?"] = collectorType
	}

	if enabled := ctx.Query("enable"); enabled != "" {
		enableBool := enabled == "true"
		conds["enable = ?"] = enableBool
	}

	count, err := c.collectorRepo.Count(conds)
	if err != nil {
		FailWithServerError(ctx, err.Error())
		return
	}

	Success(ctx, gin.H{"count": count})
}
