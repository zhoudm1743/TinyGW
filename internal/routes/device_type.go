package routes

import (
	"TinyGW/internal/service"
	"TinyGW/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterDeviceTypeRoutes(r *gin.Engine, svc service.DeviceTypeService) {
	group := r.Group("/device_type")
	{
		group.POST("", func(c *gin.Context) {
			var deviceType models.DeviceType
			if err := c.ShouldBindJSON(&deviceType); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := svc.Save(&deviceType); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		group.GET("", func(c *gin.Context) {
			deviceTypes, err := svc.FindAll()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, deviceTypes)
		})

		group.GET(":name", func(c *gin.Context) {
			name := c.Param("name")
			deviceType, err := svc.Find(name)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, deviceType)
		})

		group.DELETE(":name", func(c *gin.Context) {
			name := c.Param("name")
			if err := svc.Delete(name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		// 属性相关接口
		group.POST(":name/property", func(c *gin.Context) {
			name := c.Param("name")
			var prop models.DeviceProperty
			if err := c.ShouldBindJSON(&prop); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := svc.AddProperties(name, &prop); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		group.PUT(":name/property/:id", func(c *gin.Context) {
			name := c.Param("name")
			id, _ := strconv.Atoi(c.Param("id"))
			var prop models.DeviceProperty
			if err := c.ShouldBindJSON(&prop); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := svc.UpdateProperties(name, id, &prop); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		group.DELETE(":name/property/:id", func(c *gin.Context) {
			name := c.Param("name")
			id, _ := strconv.Atoi(c.Param("id"))
			if err := svc.DeleteProperties(name, id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		group.GET(":name/property/:id", func(c *gin.Context) {
			name := c.Param("name")
			id, _ := strconv.Atoi(c.Param("id"))
			prop, err := svc.FindProperty(name, id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, prop)
		})

		group.GET(":name/property", func(c *gin.Context) {
			name := c.Param("name")
			props, err := svc.FindAllProperties(name)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, props)
		})
	}
}
