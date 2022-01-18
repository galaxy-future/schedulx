package db

import (
	"time"

	"github.com/galaxy-future/schedulx/api/types"
)

// Instance 内实例信息
type Instance struct {
	Id               int64                `gorm:"column:id" json:"id"`
	TaskId           int64                `gorm:"column:task_id" json:"task_id"`
	InstanceId       string               `gorm:"column:instance_id" json:"instance_id"`
	ServiceClusterId int64                `gorm:"column:service_cluster_id" json:"service_cluster_id"`
	InstanceStatus   types.InstanceStatus `gorm:"column:instance_status" json:"instance_status"`
	IpInner          string               `gorm:"column:ip_inner"  json:"ip_inner"`
	IpOuter          string               `gorm:"column:ip_outer" json:"ip_outer"`
	Msg              string               `gorm:"column:msg" json:"msg"`
	CreateAt         *time.Time           `gorm:"column:create_at" json:"create_at"`
	UpdateAt         *time.Time           `gorm:"column:update_at" json:"update_at"`
}

func (t *Instance) TableName() string {
	return "instance"
}

const (
	InstanceStatusInit  = "INIT"  //初始
	InstanceStatusBase  = "BASE"  // base 环境已完成
	InstanceStatusSvc   = "SVC"   // service 环境已完成
	InstanceStatusALB   = "ALB"   // 后端挂载alb
	InstanceStatusUNALB = "UNALB" // 后端挂载alb
	InstanceStatusFail  = "FAIL"  // 异常、失败
)
