package bridgx

type InstStatus string

const (
	Undefined InstStatus = "UNDEFINED"
	Pending   InstStatus = "PENDING"
	Timeout   InstStatus = "TIMEOUT"
	Starting  InstStatus = "STARTING"
	Running   InstStatus = "RUNNING"
	Deleted   InstStatus = "DELETED"
	Deleting  InstStatus = "DELETING"
)

type TaskDescribe struct {
	TaskName    string `json:"task_name"`
	RunNum      int64  `json:"run_num"`
	SuspendNum  int64  `json:"suspend_num"`
	SuccessNum  int64  `json:"success_num"`
	FailNum     int64  `json:"fail_num"`
	SuccessRate string `json:"success_rate"`
	ExecuteTime int64  `json:"execute_time"`
	//CreateAt    string   `json:"create_at"`
}

type TaskInstancesData struct {
	InstanceList []Instance `json:"instance_list"`
	Pager        Pager      `json:"pager"`
}

type InstancesData struct {
	InstanceList []Instance `json:"instance_list"`
	Pager        Pager      `json:"pager"`
}

type Instance struct {
	InstanceId string `json:"instance_id"`
	IpInner    string `json:"ip_inner"`
	IpOuter    string `json:"ip_outer"`
	Provider   string `json:"provider"`
	CreateAt   string `json:"create_at"`
	Status     string `json:"status"`
}

type Pager struct {
	PagerNum  int `json:"pager_num"`
	PagerSize int `json:"pager_size"`
	Total     int `json:"total"`
}

type Account struct {
	UserName string `json:"user_name"`
}

type ChargeConfig struct {
	ChargeType string `json:"charge_type"`
}

type ClusterInfo struct {
	InstanceType string        `json:"instance_type"`
	Pwd          string        `json:"password"`
	UserName     string        `json:"username"`
	Provider     string        `json:"provider"`
	ChargeConfig *ChargeConfig `json:"charge_config"`
	ExtendConfig *ExtendConfig `json:"extend_config"`
}

type ExtendConfig struct {
	InstanceCore   int    `json:"core"`
	InstanceMemory int    `json:"memory"`
	CpuType        string `json:"cpu_type"`
}

type ClusterInstanceStat struct {
	InstanceTypeDesc string `json:"instance_type_desc"`
	InstanceCount    int    `json:"instance_count"`
}
