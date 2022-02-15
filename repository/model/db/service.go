package db

import (
	"time"

	"github.com/galaxy-future/schedulx/register/config/client"
	"github.com/galaxy-future/schedulx/register/config/log"
	"gorm.io/plugin/soft_delete"
)

type Service struct {
	Id          int64                 `gorm:"column:id" json:"id"`
	ServiceName string                `gorm:"column:service_name" json:"service_name"`
	Description string                `gorm:"column:description" json:"description"`
	Language    string                `gorm:"column:language" json:"language"`
	Domain      string                `gorm:"column:domain" json:"domain"`
	Port        string                `gorm:"column:port" json:"port"`
	GitRepo     string                `gorm:"column:git_repo" json:"git_repo"`
	IsDeleted   soft_delete.DeletedAt `gorm:"softDelete:flag;column:is_deleted" json:"is_deleted"`
	CreateAt    *time.Time            `gorm:"column:create_at" json:"create_at"` // 加 * 是为类触 mysql NOT NULL DEFAULT CURRENT_TIMESTAMP 属性
	UpdateAt    *time.Time            `gorm:"column:update_at" json:"update_at"`
}

func (t *Service) TableName() string {
	return "service"
}

func (t Service) GetRunningEnv() []RunningEnv {
	var envs []RunningEnv
	if err := client.ReadDBCli.Where("service_id=?", t.Id).Find(&envs).Error; err != nil {
		log.Logger.Errorf("get running env by service %v failed %v", t.Id, err)
		return nil
	}
	return envs
}

type RunningEnv struct {
	Base
	ServiceId   int64  `json:"service_id"`
	Name        string `json:"env_name"`
	Type        string `json:"env_type"`
	Domain      string `json:"domain"`
	Port        int    `json:"port"`
	HealthCheck string `json:"health_check"`
}

func (t *RunningEnv) TableName() string {
	return "running_env"
}

func (t RunningEnv) GetId() int64 {
	var env RunningEnv
	if err := client.ReadDBCli.Where("service_id=? and name=?", t.ServiceId, t.Name).First(&env).Error; err != nil {
		return 0
	}
	return env.Id
}

func (t RunningEnv) GetComputingRes() []ComputingResource {
	computingRes := make([]ComputingResource, 0)
	if err := client.ReadDBCli.Where("env_id=?", t.Id).Find(&computingRes).Error; err != nil {
		log.Logger.Errorf("get computing resource by running env %v failed %v", t.Id, err)
		return nil
	}
	return computingRes
}

type ComputingResource struct {
	Base
	EnvId         int64  `json:"env_id"`
	ComputingType string `json:"computing_type"`
	ClusterId     int64  `json:"cluster_id"`
	ClusterName   string `json:"cluster_name"`
}

func (t *ComputingResource) TableName() string {
	return "computing_resource"
}
