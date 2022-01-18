package types

type InstanceInfo struct {
	IpInner    string `json:"ip_inner"`
	IpOuter    string `json:"ip_outer"`
	InstanceId string `json:"instance_id"`
}

type InstanceAuth struct {
	UserName string `json:"user_name"`
	Pwd      string `json:"pwd"`
}

type InstanceMeta struct {
	TaskId         int64          `json:"task_id"`
	InstanceId     string         `json:"instance_id"`
	InstanceStatus InstanceStatus `json:"instance_status"`
}

type InstanceStatus string

var (
	InstanceStatusInit    InstanceStatus = "INIT"    //初始
	InstanceStatusBase    InstanceStatus = "BASE"    // base 环境已完成
	InstanceStatusSvc     InstanceStatus = "SVC"     // service 环境已完成
	InstanceStatusALB     InstanceStatus = "ALB"     // 后端挂载alb
	InstanceStatusUNALB   InstanceStatus = "UNALB"   // 后端卸载alb
	InstanceStatusFail    InstanceStatus = "FAIL"    // 异常、失败
	InstanceStatusDeleted InstanceStatus = "DELETED" // 已缩容
)
