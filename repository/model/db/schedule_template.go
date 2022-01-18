package db

import (
	"time"

	"github.com/galaxy-future/schedulx/register/constant"
	"gorm.io/plugin/soft_delete"
)

type ScheduleTemplate struct {
	Id                 int64                 `gorm:"primaryKey;column:id" json:"id"`
	TmplName           string                `gorm:"column:tmpl_name" json:"tmpl_name"`
	ServiceName        string                `gorm:"column:service_name" json:"service_name"`
	ServiceClusterId   int64                 `gorm:"column:service_cluster_id" json:"service_cluster_id"`
	BridgxClusname     string                `gorm:"column:bridgx_clusname" json:"bridgx_clusname"` // bridgx 的 cluster name
	Description        string                `gorm:"column:description" json:"description"`
	InstrGroup         string                `gorm:"column:instr_group" json:"instr_group"`
	ScheduleType       constant.ScheduleType `gorm:"column:schedule_type" json:"schedule_type"`
	ReverseSchedTmplId int64                 `gorm:"column:reverse_sched_tmpl_id" json:"reverse_sched_tmpl_id"`
	IsDeleted          soft_delete.DeletedAt `gorm:"softDelete:flag;column:is_deleted" json:"is_deleted"`
	CreateAt           *time.Time            `gorm:"column:create_at" json:"create_at"`
	UpdateAt           *time.Time            `gorm:"column:update_at" json:"update_at"`
}

func (t *ScheduleTemplate) TableName() string {
	return "schedule_template"
}
