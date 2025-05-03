package models

type User struct {
	Model
	Name     string `gorm:"type:varchar(255);primary_key;" json:"name"`
	Password string `json:"password"`
}
