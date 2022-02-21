package db

import "time"

type Task struct {
	Id             int64      `gorm:"column:id" json:"id"`
	SchedTmplId    int64      `gorm:"column:sched_tmpl_id" json:"sched_tmpl_id"`
	Operator       string     `gorm:"column:operator" json:"operator"`
	RelationTaskId string     `gorm:"column:relation_task_id" json:"relation_task_id"`
	TaskStatus     string     `gorm:"column:task_status" json:"task_status"`
	TaskStep       string     `gorm:"column:task_step" json:"task_step"`
	InstCnt        int64      `gorm:"column:inst_cnt" json:"inst_cnt"`    // 本次任务操作的实例数量
	ExecType       string     `gorm:"column:exec_type" json:"exec_type'"` // 执行方式 manual | auto
	Msg            string     `gorm:"column:msg" json:"msg"`
	TaskInfo       string     `gorm:"column:task_info" json:"task_info"`
	BeginAt        time.Time  `gorm:"column:begin_at" json:"begin_at"`
	FinishAt       *time.Time `gorm:"column:finish_at" json:"finish_at"`
	CreateAt       *time.Time `gorm:"column:create_at" json:"create_at"`
	UpdateAt       *time.Time `gorm:"column:update_at" json:"update_at"`
}

func (t *Task) TableName() string {
	return "task"
}

type RelationTaskId struct {
	BridgxTaskId  int64 `json:"bridgx_task_id"`
	NodeactTaskId int64 `json:"nodeact_task_id"`
}
