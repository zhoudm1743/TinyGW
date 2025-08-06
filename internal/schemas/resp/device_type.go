package resp

import "TinyGW/models"

type DeviceTypeResp struct {
	Name       string                  `json:"name"`
	Driver     string                  `json:"driver"`
	Properties []models.DeviceProperty `json:"properties"`
	CreatedAt  int64                   `json:"createdAt"`
	UpdatedAt  int64                   `json:"updatedAt"`
}
