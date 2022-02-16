package db

import "time"

type Service struct {
	Id          int64      `gorm:"column:id" json:"id"`
	ServiceName string     `gorm:"column:service_name" json:"service_name"`
	Description string     `gorm:"column:description" json:"description"`
	Language    string     `gorm:"column:language" json:"language"`
	Domain      string     `gorm:"column:domain" json:"domain"`
	Port        string     `gorm:"column:port" json:"port"`
	GitRepo     string     `gorm:"column:git_repo" json:"git_repo"`
	IsDeleted   int8       `gorm:"column:is_deleted" json:"is_deleted"`
	CreateAt    *time.Time `gorm:"column:create_at" json:"create_at"` // 加 * 是为类触 mysql NOT NULL DEFAULT CURRENT_TIMESTAMP 属性
	UpdateAt    *time.Time `gorm:"column:update_at" json:"update_at"`
}

func (t *Service) TableName() string {
	return "service"
}
