package db

import "time"

type InstrRecord struct {
	Id          int64      `gorm:"column:id" json:"id"`
	TaskId      string     `gorm:"column:task_id" json:"task_id"`
	InstrStatus string     `gorm:"column:instr_status" json:"instr_status"`
	Msg         string     `gorm:"column:msg" json:"msg"`
	InstrId     int64      `gorm:"column:instr_id" json:"instr_id"`
	CreateAt    *time.Time `gorm:"column:create_at" json:"create_at"`
	UpdateAt    *time.Time `gorm:"column:update_at" json:"update_at"`
}

func (t *InstrRecord) TableName() string {
	return "instr_record"
}
