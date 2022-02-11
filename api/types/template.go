package types

const (
	MountValueALB   = "slb"
	MountValueNginx = "nginx"

	DeployModeInstance  = "instance"
	DeployModeContainer = "container"
)

const (
	ENVInt       = "nodeact.initbase" //环境初始化
	SERVICEInt   = "nodeact.initsvc"  // 服务初始化
	MountTypeSLB = "mount.slb"        //挂载slb
)

type TmpInfo struct {
	TmplName         string `json:"tmpl_name"`
	ServiceClusterId int64  `json:"service_cluster_id"`
	Describe         string `json:"describe"`
	BridgxClusname   string `json:"bridgx_clusname"`
	DeployMode       string `json:"deploy_mode"`
}

type TmplAttrs struct {
	RepoPath     string `json:"repo_path"`
	RepoUser     string `json:"repo_user"`
	RepoPassword string `json:"repo_password"`
}

type BaseEnv struct {
	IsContainer bool `json:"is_container"`
}

type DownloadExec struct {
	DeployFilePath string `json:"deploy_file_path"`
	DeployFileName string `json:"deploy_file_name"`
}

type DeployParams struct {
	EnvVariables []string `json:"env_variables"` //KEY=VAL
}

type ServiceEnv struct {
	ImageStorageType string `json:"image_storage_type"`
	ImageUrl         string `json:"image_url"`
	Port             int64  `json:"port"`
	Account          string `json:"account"`
	Password         string `json:"password"`
	Cmd              string `json:"cmd"`
	ServiceName      string `json:"service_name"`
}

type DeployInfo struct {
	Strategy          string   `json:"strategy"`  //only support [rolling_update] now
	InPlace           bool     `json:"in_place"`  //upgrade app in the original instance
	MaxSurge          int      `json:"max_surge"` //percent, 20 means 20%, valid [1, 100]
	RepoPath          string   `json:"repo_path"` //download params
	RepoUser          string   `json:"repo_user"`
	RepoPassword      string   `json:"repo_password"`
	ExecutablePath    string   `json:"executable_path"`     //download params
	DeployFilePath    string   `json:"deploy_file_path"`    //file path to save
	DeployFileName    string   `json:"deploy_file_name"`    //file name to save
	EnvVariables      []string `json:"env_variables"`       //KEY=VAL
	BeforeDownloadCmd string   `json:"before_download_cmd"` //e.g. install dep before download
	BeforeDeployCmd   string   `json:"before_deploy_cmd"`   //install dep, service discovery
	DeployCmd         string   `json:"deploy_cmd"`          //download + run
	AfterDeployCmd    string   `json:"after_deploy_cmd"`    //service discovery
}

type ParamsMount struct {
	MountType  string `json:"mount_type"`
	MountValue string `json:"mount_value"`
}
