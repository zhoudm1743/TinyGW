package models

type User struct {
	Model
	Username string `json:"username" gorm:"type:varchar(255);unique;not null;index:idx_username"`
	Password string `json:"password"`
}
