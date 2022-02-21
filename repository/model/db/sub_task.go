package db

import "time"

type SubTask struct {
	Id             int64      `gorm:"column:id" json:"id"`
	SuperTaskId    int64      `gorm:"column:super_task_id" json:"super_task_id"`
	RelationTaskId string     `gorm:"column:relation_task_id" json:"relation_task_id"`
	TaskStatus     string     `gorm:"column:task_status" json:"task_status"`
	TaskStep       string     `gorm:"column:task_step" json:"task_step"`
	MaxSurge       int        `gorm:"column:max_surge" json:"max_surge"`         //percent, 20 means 20%, valid [1, 100]
	InstanceList   string     `gorm:"column:instance_list" json:"instance_list"` // 本次任务操作的实例信息
	Msg            string     `gorm:"column:msg" json:"msg"`
	TaskInfo       string     `gorm:"column:task_info" json:"task_info"`
	BeginAt        time.Time  `gorm:"column:begin_at" json:"begin_at"`
	FinishAt       *time.Time `gorm:"column:finish_at" json:"finish_at"`
	CreateAt       *time.Time `gorm:"column:create_at" json:"create_at"`
	UpdateAt       *time.Time `gorm:"column:update_at" json:"update_at"`
}

func (t *SubTask) TableName() string {
	return "task"
}
