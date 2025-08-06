package models

type Model struct {
	ID        uint  `gorm:"primaryKey" json:"id"`
	CreatedAt int64 `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt int64 `gorm:"autoUpdateTime" json:"updatedAt"`
}
