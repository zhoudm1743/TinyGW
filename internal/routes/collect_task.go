package routes

import (
	"TinyGW/internal/service"
	"TinyGW/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterCollectTaskRoutes(r *gin.Engine, svc service.CollectTaskService) {
	group := r.Group("/collect_task")
	{
		group.POST("", func(c *gin.Context) {
			var task models.CollectTask
			if err := c.ShouldBindJSON(&task); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := svc.Save(&task); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		group.GET("", func(c *gin.Context) {
			tasks, err := svc.FindAll()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, tasks)
		})

		group.GET(":name", func(c *gin.Context) {
			name := c.Param("name")
			task, err := svc.Find(name)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, task)
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
