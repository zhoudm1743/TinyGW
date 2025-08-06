package req

import "TinyGW/models"

type DeviceTypeReq struct {
	Name       string                  `json:"name" binding:"required"`
	Driver     string                  `json:"driver" binding:"required"`
	Properties []models.DeviceProperty `json:"properties"`
}
