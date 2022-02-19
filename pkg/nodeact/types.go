package nodeact

import (
	"time"

	"github.com/galaxy-future/schedulx/api/types"
)

//const (
//	StatusInit    = "INIT"
//	StatusBase    = "BASE"
//	StatusService = "SVC"
//	StatusFail    = "FAIL"
//)

type InitBaseRequest struct {
	InstanceGroup  *InstanceGroup      `json:"instance_group'"`
	Auth           *types.InstanceAuth `json:"auth"`
	HarborRegistry string              `json:"harbor_registry"`
}

type InitServiceRequest struct {
	InstanceGroup *InstanceGroup      `json:"instance_group"`
	Auth          *types.InstanceAuth `json:"auth"`
	Params        *ParamsServiceEnv   `json:"params"`
}

type InstanceGroup struct {
	TaskId       int64                 `json:"task_id"` // relationTaskId
	InstanceList []*types.InstanceInfo `json:"instance_list"`
}

//type InstanceInfo struct {
//	IpInner    string `json:"ip_inner"`
//	IpOuter    string `json:"ip_outer"`
//	InstanceId string `json:"instance_id"`
//	//InstanceStatus string `json:"instance_status"`
//}

//type InstanceAuth struct {
//	UserName string `json:"user_name"`
//	Pwd      string `json:"pwd"`
//}

type InstanceMeta struct {
	TaskId         int64          `json:"task_id"`
	InstanceId     string         `json:"instance_id"`
	InstanceStatus InstanceStatus `json:"instance_status"`
}

type InstanceStatus string

const (
	InstanceStatusInit InstanceStatus = "INIT" //初始
	InstanceStatusBase InstanceStatus = "BASE" // base 环境已完成
	InstanceStatusSvc  InstanceStatus = "SVC"  // service 环境已完成
	InstanceStatusFail InstanceStatus = "FAIL" // 异常、失败
)

type ParamsServiceEnv struct {
	ImageStorageType string `json:"image_storage_type"`
	ImageUrl         string `json:"image_url"`
	ServiceName      string `json:"service_name"`
	Port             int64  `json:"port"`
}

type ParamsBaseEnv struct {
	IsContainer bool `json:"is_container"`
}

type ParamsMountInfo struct {
	MountType  string `json:"mount_type"` // alb / nginx
	MountValue string `json:"mount_value"`
}

type TaskDescribe struct {
	FoundTime   *time.Time `json:"found_time"`
	TotalNum    int64      `json:"total_num"`
	SuccessNum  int64      `json:"success_num"`
	FailNum     int64      `json:"fail_num"`
	SuccessRate string     `json:"success_rate"`
}

type TaskInstancesData struct {
	InstancesList []*types.InstanceInfo `json:"instances_list"`
	Pager         Pager                 `json:"pager"`
}

type Pager struct {
	PagerNum  int `json:"pager_num"`
	PagerSize int `json:"pager_size"`
	Total     int `json:"total"`
}
