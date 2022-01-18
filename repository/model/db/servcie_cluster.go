package db

import "time"

type ServiceCluster struct {
	Id            int64      `gorm:"primaryKey;column:id" json:"id"`
	ServiceName   string     `gorm:"column:service_name" json:"service_name"`
	ClusterName   string     `gorm:"column:cluster_name" json:"cluster_name"`
	BridgxCluster string     `gorm:"bridgx_cluster" json:"bridgx_cluster"`
	AutoDecision  string     `gorm:"column:auto_decision" json:"auto_decision"`
	CreateAt      *time.Time `gorm:"column:create_at" json:"create_at"`
	UpdateAt      *time.Time `gorm:"column:update_at" json:"update_at"`
}

func (t *ServiceCluster) TableName() string {
	return "service_cluster"
}
