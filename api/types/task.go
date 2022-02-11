package types

import "time"

const (
	TaskStatusInit        = "INIT"
	TaskStatusRunning     = "RUNNING"
	TaskStatusRollingBack = "ROLLING_BACK"
	TaskStatusSuccess     = "SUCC"
	TaskStatusFail        = "FAIL"
)

const (
	//扩容
	TaskExpand = "expand"
	//缩容
	TaskShrink = "shrink"
)

var TaskStatusDescMap = map[string]string{
	TaskStatusInit:        "已创建",
	TaskStatusRunning:     "进行中",
	TaskStatusRollingBack: "回滚中",
	TaskStatusSuccess:     "成功",
	TaskStatusFail:        "失败",
}

var TaskStatusDesc = func(TaskStatus string) string {
	if v, ok := TaskStatusDescMap[TaskStatus]; ok {
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

	TaskStepDeployBeforeDownloadInit = "DEPLOY_BEFORE_DOWNLOAD_INIT"
	TaskStepDeployBeforeDownloadSucc = "DEPLOY_BEFORE_DOWNLOAD_SUCC"
	TaskStepDeployBeforeDownloadFail = "DEPLOY_BEFORE_DOWNLOAD_FAIL"
	TaskStepDeployDownloadInit       = "DEPLOY_DOWNLOAD_INIT"
	TaskStepDeployDownloadSucc       = "DEPLOY_DOWNLOAD_SUCC"
	TaskStepDeployDownloadFail       = "DEPLOY_DOWNLOAD_FAIL"
	TaskStepDeployBeforeDeployInit   = "DEPLOY_BEFORE_DEPLOY_INIT"
	TaskStepDeployBeforeDeploySucc   = "DEPLOY_BEFORE_DEPLOY_SUCC"
	TaskStepDeployBeforeDeployFail   = "DEPLOY_BEFORE_DEPLOY_FAIL"
	TaskStepDeployDeployInit         = "DEPLOY_DEPLOY_INIT"
	TaskStepDeployDeploySucc         = "DEPLOY_DEPLOY_SUCC"
	TaskStepDeployDeployFail         = "DEPLOY_DEPLOY_FAIL"
	TaskStepDeployAfterDeployInit    = "DEPLOY_AFTER_DEPLOY_INIT"
	TaskStepDeployAfterDeploySucc    = "DEPLOY_AFTER_DEPLOY_SUCC"
	TaskStepDeployAfterDeployFail    = "DEPLOY_AFTER_DEPLOY_FAIL"
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

	TaskStepDeployBeforeDownloadInit: "环境初始化中",
	TaskStepDeployBeforeDownloadSucc: "环境初始化成功",
	TaskStepDeployBeforeDownloadFail: "环境初始化失败",
	TaskStepDeployDownloadInit:       "下载中",
	TaskStepDeployDownloadSucc:       "下载成功",
	TaskStepDeployDownloadFail:       "下载失败",
	TaskStepDeployBeforeDeployInit:   "部署前置任务执行中",
	TaskStepDeployBeforeDeploySucc:   "部署前置任务执行成功",
	TaskStepDeployBeforeDeployFail:   "部署前置任务执行失败",
	TaskStepDeployDeployInit:         "部署中",
	TaskStepDeployDeploySucc:         "部署成功",
	TaskStepDeployDeployFail:         "部署失败",
	TaskStepDeployAfterDeployInit:    "部署后置任务执行中",
	TaskStepDeployAfterDeploySucc:    "部署后置任务执行成功",
	TaskStepDeployAfterDeployFail:    "部署后置任务执行失败",
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
	TaskStatus     string `json:"task_status"`
	TaskStatusDesc string `json:"task_status_desc"`
	TaskStep       string `json:"task_step"`
	TaskStepDesc   string `json:"task_step_desc"`
	InstCnt        int64  `json:"inst_cnt"`
	Msg            string `json:"msg"`
	Operator       string `json:"operator"`
	ExecType       string `json:"exec_type"`
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
