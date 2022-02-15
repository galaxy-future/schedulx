package db

import "time"

type Integration struct {
	Id       int64      `json:"id" gorm:"id"`
	Host     string     `json:"host" gorm:"host"`
	Account  string     `json:"account" gorm:"account"`
	Password string     `json:"password" gorm:"password"`
	Type     string     `json:"type" gorm:"type"`
	CreateAt *time.Time `json:"create_at" gorm:"create_at"`
	UpdateAt *time.Time `json:"update_at" gorm:"update_at"`
	CreateBy string     `json:"create_by" gorm:"create_by"`
	UpdateBy string     `json:"update_by" gorm:"update_by"`
}

func (t *Integration) TableName() string {
	return "integration"
}
