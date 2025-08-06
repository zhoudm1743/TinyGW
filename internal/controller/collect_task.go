package controller

import (
	"TinyGW/internal/repo"
	"TinyGW/internal/schemas/req"
	"TinyGW/internal/schemas/resp"
	"TinyGW/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CollectTaskController struct {
	collectTaskRepo *repo.CollectTaskRepo
}

func NewCollectTaskController(collectTaskRepo *repo.CollectTaskRepo) *CollectTaskController {
	return &CollectTaskController{
		collectTaskRepo: collectTaskRepo,
	}
}

// Create 创建采集任务
func (c *CollectTaskController) Create(ctx *gin.Context) {
	var taskReq req.CollectTaskReq
	if err := ctx.ShouldBindJSON(&taskReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &models.CollectTask{
		Name:       taskReq.Name,
		Cron:       taskReq.Cron,
		Status:     taskReq.Status,
		DeviceList: taskReq.DeviceList,
	}

	if err := c.collectTaskRepo.Create(task); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "采集任务创建成功"})
}

// Update 更新采集任务
func (c *CollectTaskController) Update(ctx *gin.Context) {
	var taskReq req.CollectTaskReq
	if err := ctx.ShouldBindJSON(&taskReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := c.collectTaskRepo.Get(taskReq.Name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "采集任务不存在"})
		return
	}

	task.Cron = taskReq.Cron
	task.Status = taskReq.Status
	task.DeviceList = taskReq.DeviceList

	if err := c.collectTaskRepo.Update(task); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "采集任务更新成功"})
}

// Delete 删除采集任务
func (c *CollectTaskController) Delete(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "采集任务名不能为空"})
		return
	}

	task, err := c.collectTaskRepo.Get(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "采集任务不存在"})
		return
	}

	if err := c.collectTaskRepo.Delete(task); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "采集任务删除成功"})
}

// Get 获取采集任务
func (c *CollectTaskController) Get(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "采集任务名不能为空"})
		return
	}

	task, err := c.collectTaskRepo.Get(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "采集任务不存在"})
		return
	}

	taskResp := resp.CollectTaskResp{
		Name:       task.Name,
		Cron:       task.Cron,
		Status:     task.Status,
		DeviceList: task.DeviceList,
		CreatedAt:  task.CreatedAt,
		UpdatedAt:  task.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, taskResp)
}

// List 获取采集任务列表
func (c *CollectTaskController) List(ctx *gin.Context) {
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
	if status := ctx.Query("status"); status != "" {
		statusInt, err := strconv.Atoi(status)
		if err == nil {
			conds["status = ?"] = int8(statusInt)
		}
	}

	total, tasks, err := c.collectTaskRepo.Find(page, pageSize, conds)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var taskResps []resp.CollectTaskResp
	for _, task := range tasks {
		taskResps = append(taskResps, resp.CollectTaskResp{
			Name:       task.Name,
			Cron:       task.Cron,
			Status:     task.Status,
			DeviceList: task.DeviceList,
			CreatedAt:  task.CreatedAt,
			UpdatedAt:  task.UpdatedAt,
		})
	}

	ctx.JSON(http.StatusOK, resp.PageResp{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     taskResps,
	})
}

// Count 获取采集任务数量
func (c *CollectTaskController) Count(ctx *gin.Context) {
	conds := make(map[string]interface{})

	// 可选过滤条件
	if status := ctx.Query("status"); status != "" {
		statusInt, err := strconv.Atoi(status)
		if err == nil {
			conds["status = ?"] = int8(statusInt)
		}
	}

	count, err := c.collectTaskRepo.Count(conds)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// Start 启动采集任务
func (c *CollectTaskController) Start(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "采集任务名不能为空"})
		return
	}

	task, err := c.collectTaskRepo.Get(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "采集任务不存在"})
		return
	}

	task.Status = 1 // 假设1表示启动状态
	if err := c.collectTaskRepo.Update(task); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "采集任务已启动"})
}

// Stop 停止采集任务
func (c *CollectTaskController) Stop(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "采集任务名不能为空"})
		return
	}

	task, err := c.collectTaskRepo.Get(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "采集任务不存在"})
		return
	}

	task.Status = 0 // 假设0表示停止状态
	if err := c.collectTaskRepo.Update(task); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "采集任务已停止"})
}
