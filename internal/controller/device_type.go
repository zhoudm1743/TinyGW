package controller

import (
	"TinyGW/internal/repo"
	"TinyGW/internal/schemas/req"
	"TinyGW/internal/schemas/resp"
	"TinyGW/models"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type DeviceTypeController struct {
	deviceTypeRepo *repo.DeviceTypeRepo
	deviceRepo     *repo.DeviceRepo
}

func NewDeviceTypeController(deviceTypeRepo *repo.DeviceTypeRepo, deviceRepo *repo.DeviceRepo) *DeviceTypeController {
	return &DeviceTypeController{
		deviceTypeRepo: deviceTypeRepo,
		deviceRepo:     deviceRepo,
	}
}

// Create 创建设备类型
func (c *DeviceTypeController) Create(ctx *gin.Context) {
	var typeReq req.DeviceTypeReq
	if err := ctx.ShouldBindJSON(&typeReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceType := &models.DeviceType{
		Name:       typeReq.Name,
		Driver:     typeReq.Driver,
		Properties: typeReq.Properties,
	}

	if err := c.deviceTypeRepo.Create(deviceType); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "设备类型创建成功"})
}

// Update 更新设备类型
func (c *DeviceTypeController) Update(ctx *gin.Context) {
	var typeReq req.DeviceTypeReq
	if err := ctx.ShouldBindJSON(&typeReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceType, err := c.deviceTypeRepo.Get(typeReq.Name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "设备类型不存在"})
		return
	}

	deviceType.Driver = typeReq.Driver
	deviceType.Properties = typeReq.Properties

	if err := c.deviceTypeRepo.Update(deviceType); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "设备类型更新成功"})
}

// Delete 删除设备类型
func (c *DeviceTypeController) Delete(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "设备类型名不能为空"})
		return
	}

	deviceType, err := c.deviceTypeRepo.Get(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "设备类型不存在"})
		return
	}

	if err := c.deviceTypeRepo.Delete(deviceType); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "设备类型删除成功"})
}

// Get 获取设备类型
func (c *DeviceTypeController) Get(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "设备类型名不能为空"})
		return
	}

	deviceType, err := c.deviceTypeRepo.Get(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "设备类型不存在"})
		return
	}

	typeResp := resp.DeviceTypeResp{
		Name:       deviceType.Name,
		Driver:     deviceType.Driver,
		Properties: deviceType.Properties,
		CreatedAt:  deviceType.CreatedAt,
		UpdatedAt:  deviceType.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, typeResp)
}

// List 获取设备类型列表
func (c *DeviceTypeController) List(ctx *gin.Context) {
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
	if driver := ctx.Query("driver"); driver != "" {
		conds["driver = ?"] = driver
	}

	total, types, err := c.deviceTypeRepo.Find(page, pageSize, conds)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var typeResps []resp.DeviceTypeResp
	for _, deviceType := range types {
		typeResps = append(typeResps, resp.DeviceTypeResp{
			Name:       deviceType.Name,
			Driver:     deviceType.Driver,
			Properties: deviceType.Properties,
			CreatedAt:  deviceType.CreatedAt,
			UpdatedAt:  deviceType.UpdatedAt,
		})
	}

	ctx.JSON(http.StatusOK, resp.PageResp{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     typeResps,
	})
}

// GetAll 获取所有设备类型
func (c *DeviceTypeController) GetAll(ctx *gin.Context) {
	types, err := c.deviceTypeRepo.FindAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var typeResps []resp.DeviceTypeResp
	for _, deviceType := range types {
		typeResps = append(typeResps, resp.DeviceTypeResp{
			Name:       deviceType.Name,
			Driver:     deviceType.Driver,
			Properties: deviceType.Properties,
			CreatedAt:  deviceType.CreatedAt,
			UpdatedAt:  deviceType.UpdatedAt,
		})
	}

	ctx.JSON(http.StatusOK, typeResps)
}

// Count 获取设备类型数量
func (c *DeviceTypeController) Count(ctx *gin.Context) {
	conds := make(map[string]interface{})

	// 可选过滤条件
	if driver := ctx.Query("driver"); driver != "" {
		conds["driver = ?"] = driver
	}

	count, err := c.deviceTypeRepo.Count(conds)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"count": count})
}

// GetProperties 获取设备类型属性
func (c *DeviceTypeController) GetProperties(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "设备类型名不能为空"})
		return
	}

	deviceType, err := c.deviceTypeRepo.Get(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "设备类型不存在"})
		return
	}

	ctx.JSON(http.StatusOK, deviceType.Properties)
}

// UpdateProperties 更新设备类型属性
func (c *DeviceTypeController) UpdateProperties(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "设备类型名不能为空"})
		return
	}

	var properties []models.DeviceProperty
	if err := ctx.ShouldBindJSON(&properties); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deviceType, err := c.deviceTypeRepo.Get(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "设备类型不存在"})
		return
	}

	deviceType.Properties = properties

	if err := c.deviceTypeRepo.Update(deviceType); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "设备类型属性更新成功"})
}

// AddProperty 添加设备类型属性
func (c *DeviceTypeController) AddProperty(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "设备类型名不能为空"})
		return
	}

	var propertyReq req.DevicePropertyReq
	if err := ctx.ShouldBindJSON(&propertyReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	property := models.DeviceProperty{
		Name:        propertyReq.Name,
		Description: propertyReq.Description,
		Type:        propertyReq.Type,
		Length:      propertyReq.Length,
		Decimal:     propertyReq.Decimal,
		Unit:        propertyReq.Unit,
		Value:       propertyReq.Value,
		Reported:    propertyReq.Reported,
		IsAlarm:     propertyReq.IsAlarm,
		Threshold:   propertyReq.Threshold,
		Used:        propertyReq.Used,
		AutoCalc:    propertyReq.AutoCalc,
		Scale:       propertyReq.Scale,
	}

	if err := c.deviceTypeRepo.AddProperty(name, &property); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 通知设备类型已更改
	if c.deviceRepo != nil {
		// 获取更新后的设备类型，但忽略错误，因为我们已经确认设备类型存在
		_, _ = c.deviceTypeRepo.Get(name)
		// TODO: 实现设备类型变更通知
		// c.deviceRepo.DeviceTypeChanged(deviceType)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "属性添加成功"})
}

// UpdateProperty 更新设备类型属性
func (c *DeviceTypeController) UpdateProperty(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "设备类型名不能为空"})
		return
	}

	propertyName := ctx.Param("propertyName")
	if propertyName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "属性名不能为空"})
		return
	}

	var propertyReq req.DevicePropertyReq
	if err := ctx.ShouldBindJSON(&propertyReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	property := models.DeviceProperty{
		Name:        propertyReq.Name,
		Description: propertyReq.Description,
		Type:        propertyReq.Type,
		Length:      propertyReq.Length,
		Decimal:     propertyReq.Decimal,
		Unit:        propertyReq.Unit,
		Value:       propertyReq.Value,
		Reported:    propertyReq.Reported,
		IsAlarm:     propertyReq.IsAlarm,
		Threshold:   propertyReq.Threshold,
		Used:        propertyReq.Used,
		AutoCalc:    propertyReq.AutoCalc,
		Scale:       propertyReq.Scale,
	}

	if err := c.deviceTypeRepo.UpdateProperty(name, propertyName, &property); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 通知设备类型已更改
	if c.deviceRepo != nil {
		// 获取更新后的设备类型，但忽略错误，因为我们已经确认设备类型存在
		_, _ = c.deviceTypeRepo.Get(name)
		// TODO: 实现设备类型变更通知
		// c.deviceRepo.DeviceTypeChanged(deviceType)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "属性更新成功"})
}

// DeleteProperty 删除设备类型属性
func (c *DeviceTypeController) DeleteProperty(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "设备类型名不能为空"})
		return
	}

	propertyName := ctx.Param("propertyName")
	if propertyName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "属性名不能为空"})
		return
	}

	if err := c.deviceTypeRepo.DeleteProperty(name, propertyName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 通知设备类型已更改
	if c.deviceRepo != nil {
		// 获取更新后的设备类型，但忽略错误，因为我们已经确认设备类型存在
		_, _ = c.deviceTypeRepo.Get(name)
		// TODO: 实现设备类型变更通知
		// c.deviceRepo.DeviceTypeChanged(deviceType)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "属性删除成功"})
}

// GetProperty 获取设备类型属性
func (c *DeviceTypeController) GetProperty(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "设备类型名不能为空"})
		return
	}

	propertyName := ctx.Param("propertyName")
	if propertyName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "属性名不能为空"})
		return
	}

	property, err := c.deviceTypeRepo.GetProperty(name, propertyName)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, property)
}

// Upload 上传设备类型驱动
func (c *DeviceTypeController) Upload(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "设备类型名不能为空"})
		return
	}

	deviceType, err := c.deviceTypeRepo.Get(name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("设备类型 %s 不存在", name)})
		return
	}

	// 确保插件目录存在
	if err := ensureDir("plugin"); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("创建插件目录失败: %s", err.Error())})
		return
	}

	// 获取当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取当前目录失败: %s", err.Error())})
		return
	}

	// 获取上传的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("上传失败: %s", err.Error())})
		return
	}

	// 保存文件
	pluginDir := path.Join(dir, "plugin")
	filename := path.Join(pluginDir, file.Filename)
	if err := ctx.SaveUploadedFile(file, filename); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存文件失败: %s", err.Error())})
		return
	}
	defer os.Remove(filename) // 解压后删除原始文件

	// 解压文件
	if err := unzip(filename, pluginDir); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("解压失败: %s", err.Error())})
		return
	}

	// 提取插件名称
	filenameWithSuffix := path.Base(filename)
	fileType := path.Ext(filename)
	plugin := strings.TrimSuffix(filenameWithSuffix, fileType)
	deviceType.Driver = plugin

	// 尝试读取属性文件
	propertiesFile := path.Join("plugin", plugin, "properties.json")
	if fileExists(propertiesFile) {
		data, err := os.ReadFile(propertiesFile)
		if err == nil {
			var properties []models.DeviceProperty
			if err := json.Unmarshal(data, &properties); err == nil {
				deviceType.Properties = properties
			}
		}
	}

	// 保存设备类型
	if err := c.deviceTypeRepo.Update(deviceType); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存失败: %s", err.Error())})
		return
	}

	typeResp := resp.DeviceTypeResp{
		Name:       deviceType.Name,
		Driver:     deviceType.Driver,
		Properties: deviceType.Properties,
		CreatedAt:  deviceType.CreatedAt,
		UpdatedAt:  deviceType.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, typeResp)
}

// 确保目录存在
func ensureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// 解压文件
func unzip(src, dest string) error {
	// TODO: 实现解压功能
	// 这里需要实现一个解压函数，可以使用标准库或第三方库
	return nil
}
