package routes

import (
	"TinyGW/internal/service"
	"TinyGW/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterDeviceRoutes(r *gin.Engine, svc service.DeviceService) {
	group := r.Group("/device")
	{
		group.POST("", func(c *gin.Context) {
			var device models.Device
			if err := c.ShouldBindJSON(&device); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := svc.Save(&device); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		group.GET("", func(c *gin.Context) {
			devices, err := svc.FindAll()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, devices)
		})

		group.GET(":name", func(c *gin.Context) {
			name := c.Param("name")
			device, err := svc.Find(name)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, device)
		})

		group.DELETE(":name", func(c *gin.Context) {
			name := c.Param("name")
			if err := svc.Delete(name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})
	}
}
