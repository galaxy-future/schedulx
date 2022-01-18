package types

import "time"

const (
	TaskStatusInit    = "INIT"
	TaskStatusRunning = "RUNNING"
	TaskStatusSuccess = "SUCC"
	TaskStatusFail    = "FAIL"
)

const (
	//扩容
	TaskExpand = "expand"
	//缩容
	TaskShrink = "shrink"
)

var TaskStatusDescMap = map[string]string{
	TaskStatusInit:    "已创建",
	TaskStatusRunning: "进行中",
	TaskStatusSuccess: "成功",
	TaskStatusFail:    "失败",
}

var TaskStatusDesc = func(TaskStep string) string {
	if v, ok := TaskStatusDescMap[TaskStep]; ok {
		return v
	}
	return "UnKnown"
}

const (
	TaskStepInit             = "INIT"
	TaskStepBridgxExpandInit = "BRIDGX_EXPAND_INIT"
	TaskStepBridgxShrinkInit = "BRIDGX_SHRINK_INIT"
	TaskStepBridgxExpandSucc = "BRIDGX_EXPAND_SUCC"
	TaskStepBridgxShrinkSucc = "BRIDGX_SHRINK_SUCC"
	TaskStepBaseEnvInit      = "BASE_ENV_INIT"
	TaskStepBaseEnvSucc      = "BASE_ENV_SUCC"
	TaskStepSvcEnvInit       = "SVC_ENV_INIT"
	TaskStepSvcEnvSucc       = "SVC_ENV_SUCC"
	TaskStepMountInit        = "MOUNT_INIT"
	TaskStepUmountInit       = "UMOUNT_INIT"
	TaskStepMountSucc        = "MOUNT_SUCC"
	TaskStepUmountSucc       = "UMOUNT_SUCC"
)

var TaskStepDescMap = map[string]string{
	TaskStepInit:             "待执行",
	TaskStepBridgxExpandInit: "计算资源扩容",
	TaskStepBridgxShrinkInit: "计算资源缩容",
	TaskStepBridgxExpandSucc: "计算资源已获取",
	TaskStepBridgxShrinkSucc: "计算资源已缩容",
	TaskStepBaseEnvInit:      "基础环境搭建",
	TaskStepBaseEnvSucc:      "基础环境搭建成功",
	TaskStepSvcEnvInit:       "服务搭建",
	TaskStepSvcEnvSucc:       "服务搭建成功",
	TaskStepMountInit:        "执行实例挂载",
	TaskStepUmountInit:       "执行实例卸载",
	TaskStepMountSucc:        "实例挂载成功",
	TaskStepUmountSucc:       "实例卸载成功",
}

var TaskStepDesc = func(TaskStep string) string {
	if v, ok := TaskStepDescMap[TaskStep]; ok {
		return v
	}
	return "UnKnown"
}

type TaskDescribe struct {
	FoundTime   *time.Time `json:"found_time"`
	TotalNum    int64      `json:"total_num"`
	SuccessNum  int64      `json:"success_num"`
	FailNum     int64      `json:"fail_num"`
	SuccessRate string     `json:"success_rate"`
}

type Pager struct {
	PagerNum  int `json:"pager_num"`
	PagerSize int `json:"pager_size"`
	Total     int `json:"total"`
}

type Instance struct {
	InstanceId string         `json:"instance_id"`
	IpInner    string         `json:"ip_inner"`
	IpOuter    string         `json:"ip_outer"`
	CreateAt   string         `json:"create_at"`
	Status     InstanceStatus `json:"status"`
}

type TaskInfo struct {
	TaskStatus string `json:"task_status"`
	TaskStep   string `json:"task_step"`
	InstCnt    int64  `json:"inst_cnt"`
	Msg        string `json:"msg"`
	Operator   string `json:"operator"`
	ExecType   string `json:"exec_type"`
}

type InstInfoResp struct {
	InstanceId string         `json:"instance_id"`
	IpInner    string         `json:"ip_inner"`
	IpOuter    string         `json:"ip_outer"`
	Status     InstanceStatus `json:"instance_status"`
}

type RelationTaskId struct {
	NodeActTaskId int64 `json:"nodeact_task_id"`
	BridgXTaskId  int64 `json:"bridgx_task_id"`
}

const (
	NodeactTaskId = "nodeact_task_id"
	BridgXTaskId  = "bridgx_task_id"
)
