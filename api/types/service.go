package types

type ServiceInfo struct {
	ServiceName string `json:"service_name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Domain      string `json:"domain"`
	Port        string `json:"port"`
	GitRepo     string `json:"git_repo"`
}

const (
	ProductEnv   = "PRODUCT"
	PreOnlineEnv = "PRE_ONLINE"
	TestEnv      = "TEST"
	PressEnv     = "PRESS"
	DevelopEnv   = "DEVELOP"
)

const (
	CloudCluster = "CLOUD_CLUSTER"
	K8sCluster   = "K8S_CLUSTER"
)

type RunningEnv struct {
	ServiceId   int64               `json:"service_id"`
	Id          int64               `json:"env_id"`
	Name        string              `json:"env_name"`
	Type        string              `json:"env_type"`
	Domain      string              `json:"domain"`
	Port        int                 `json:"port"`
	HealthCheck HealthCheck         `json:"health_check"`
	Resources   []ComputingResource `json:"computing_resources"`
}

type ComputingResource struct {
	Id            int64  `json:"resource_id"`
	ComputingType string `json:"computing_type"`
	ResourceInfo
}

type ResourceInfo struct {
	ClusterId          int64  `json:"cluster_id"`
	BridgxCluster      string `json:"bridgx_cluster"`
	InstanceCount      int64  `json:"instance_count"`
	InstanceTypeDesc   string `json:"instance_type_desc"`
	Provider           string `json:"provider"`
	ComputingPowerType string `json:"computing_power_type"`
	ChargeType         string `json:"charge_type"`
}

type HealthCheck struct {
	Mode               string `json:"mode'"`
	Path               string `json:"path"`
	Port               int    `json:"port"`
	InitTime           int    `json:"init_time"`
	TimeoutTime        int    `json:"timeout_time"`
	HealthThreshold    int    `json:"health_threshold"`
	UnhealthyThreshold int    `json:"unhealthy_threshold"`
	CheckPeriod        int    `json:"check_period"`
}
