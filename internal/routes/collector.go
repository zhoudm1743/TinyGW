package routes

import (
	"TinyGW/internal/service"
	"TinyGW/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterCollectorRoutes(r *gin.Engine, svc service.CollectorService) {
	group := r.Group("/collector")
	{
		group.POST("", func(c *gin.Context) {
			var collector models.Collector
			if err := c.ShouldBindJSON(&collector); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := svc.Save(&collector); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		group.GET("", func(c *gin.Context) {
			collectors, err := svc.FindAll()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, collectors)
		})

		group.GET(":name", func(c *gin.Context) {
			name := c.Param("name")
			collector, err := svc.Find(name)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, collector)
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
