package service

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/pkg/tool"
	"github.com/galaxy-future/schedulx/register/config"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/repository"
	"github.com/galaxy-future/schedulx/template"
)

const (
	_remoteBaseEnvScript    = "/root/init_base.sh"
	_remoteServiceEnvScript = "/root/init_service.sh"
)

type EnvService struct {
}

var envServiceInstance *EnvService
var envOnce sync.Once

func GetEnvOpsSvcInst() *EnvService {
	envOnce.Do(func() {
		envServiceInstance = &EnvService{}
	})
	return envServiceInstance
}

type BaseEnvInitAsyncSvcReq struct {
	ServiceClusterId int64                 `json:"service_cluster_id"`
	TaskId           int64                 `json:"task_id"`
	Cmd              string                `json:"cmd"`
	InstanceList     []*types.InstanceInfo `json:"instance_list"`
	Auth             *types.InstanceAuth   `json:"auth"`
}

type DeployAsyncReq struct {
	ServiceClusterId int64                 `json:"service_cluster_id"`
	TaskId           int64                 `json:"task_id"`
	Cmd              string                `json:"cmd"`
	InstanceList     []*types.InstanceInfo `json:"instance_list"`
	Auth             *types.InstanceAuth   `json:"auth"`
}

type SvcEnvInitAsyncSvcReq struct {
	ServiceClusterId int64                   `json:"service_cluster_id"`
	TaskId           int64                   `json:"task_id"`
	InstanceList     []*types.InstanceInfo   `json:"instance_list"`
	Auth             *types.InstanceAuth     `json:"auth"`
	Params           *types.ParamsServiceEnv `json:"params"`
	Cmd              string                  `json:"string"`
}

func (s *EnvService) entryLog(ctx context.Context, method string, req interface{}) {
	log.Logger.Infof("entry log | method[%s] | req:%s", method, tool.ToJson(req))
}

func (s *EnvService) exitLog(ctx context.Context, method string, req, resp interface{}, err error) {
	log.Logger.Infof("exit log | method[%s] | req:%s | resp:%s | err:%v", method, tool.ToJson(req), tool.ToJson(resp), err)
}

func (s *EnvService) BaseEnvInitAsync(ctx context.Context, svcReq *BaseEnvInitAsyncSvcReq) error {
	var err error
	s.entryLog(ctx, "BaseEnvInitAsync", svcReq) // todo 日志脱敏
	defer func() {
		s.exitLog(ctx, "BaseEnvInitAsync", svcReq, nil, err)
	}()
	// 机器信息入库
	if err = s.NodeUpdateStore(ctx, svcReq.InstanceList, svcReq.TaskId, svcReq.ServiceClusterId); err != nil {
		return err
	}
	//异步初始化
	ipList := svcReq.InstanceList
	taskId := svcReq.TaskId
	log.Logger.Info("start init base env async")
	for _, instInfo := range ipList {
		log.Logger.Infof("async initbase instanceid:%s", instInfo.InstanceId)
		go func(instance *types.InstanceInfo) {
			_ = s.BaseEnvInitSingle(ctx, taskId, instance, svcReq.Auth)
		}(instInfo)
	}
	log.Logger.Info("end init base env async")
	return nil
}

func (s *EnvService) DeployBeforeDownloadInitAsync(ctx context.Context, svcReq *BaseEnvInitAsyncSvcReq) error {
	var err error
	s.entryLog(ctx, "DeployBeforeDownloadInitAsync", svcReq)
	defer func() {
		s.exitLog(ctx, "DeployBeforeDownloadInitAsync", svcReq, nil, err)
	}()
	// 机器信息入库
	if err = s.NodeUpdateDeploy(ctx, svcReq.InstanceList, svcReq.TaskId, svcReq.ServiceClusterId); err != nil {
		return err
	}
	//异步初始化
	ipList := svcReq.InstanceList
	taskId := svcReq.TaskId
	log.Logger.Info("start DeployBeforeDownloadInitAsync async")
	var wg sync.WaitGroup
	for _, instInfo := range ipList {
		wg.Add(1)
		log.Logger.Infof("async DeployBeforeDownloadInitAsync instanceid:%s", instInfo.InstanceId)
		go func(instance *types.InstanceInfo) {
			_ = s.DeployBeforeDownloadInitSingle(ctx, taskId, svcReq.Cmd, instance, svcReq.Auth)
			wg.Done()
		}(instInfo)
	}
	wg.Wait()
	log.Logger.Info("end DeployBeforeDownloadInitAsync async")
	return nil
}

func (s *EnvService) doDeploy(ctx context.Context, svcReq *DeployAsyncReq, method string, instanceStatus types.InstanceStatus) error {
	var err error
	s.entryLog(ctx, method, svcReq)
	defer func() {
		s.exitLog(ctx, method, svcReq, nil, err)
	}()

	//异步初始化
	ipList := svcReq.InstanceList
	taskId := svcReq.TaskId
	log.Logger.Infof("start %v async", method)
	var wg sync.WaitGroup
	for _, instInfo := range ipList {
		wg.Add(1)
		log.Logger.Infof("async %v instanceid:%s", method, instInfo.InstanceId)
		go func(instance *types.InstanceInfo) {
			e := s.ExecCmdWithUpdateInstanceStatus(ctx, taskId, svcReq.Cmd, instance, svcReq.Auth, instanceStatus)
			if err == nil && e != nil {
				err = e
			}
			wg.Done()
		}(instInfo)
	}
	wg.Wait()
	log.Logger.Infof("end %v async", method)
	return err
}

func (s *EnvService) DeployDownloadAsync(ctx context.Context, svcReq *DeployAsyncReq) error {
	return s.doDeploy(ctx, svcReq, "DeployDownloadAsync", types.InstanceStatusDownload)
}

func (s *EnvService) BeforeDeployAsync(ctx context.Context, svcReq *DeployAsyncReq) error {
	return s.doDeploy(ctx, svcReq, "BeforeDeployAsync", types.InstanceStatusBeforeDeploy)
}

func (s *EnvService) DeployAsync(ctx context.Context, svcReq *DeployAsyncReq) error {
	return s.doDeploy(ctx, svcReq, "DeployAsync", types.InstanceStatusDeploy)
}

func (s *EnvService) AfterDeployAsync(ctx context.Context, svcReq *DeployAsyncReq) error {
	return s.doDeploy(ctx, svcReq, "AfterDeployAsync", types.InstanceStatusAfterDeploy)
}

func (s *EnvService) NodeUpdateStore(ctx context.Context, instanceList []*types.InstanceInfo, taskId, serviceClusterId int64) error {
	var err error
	repo := repository.GetInstanceRepoIns()
	// todo UpInsertTask()
	if err = repo.UpInsertInstanceBatch(ctx, instanceList, taskId, serviceClusterId); err != nil {
		return err
	}
	return nil
}
func (s *EnvService) NodeUpdateDeploy(ctx context.Context, instanceList []*types.InstanceInfo, taskId, serviceClusterId int64) error {
	var err error
	repo := repository.GetInstanceRepoIns()
	if err = repo.UpInsertInstanceBatchByCluster(ctx, instanceList, taskId, serviceClusterId); err != nil {
		return err
	}
	return nil
}

func (s *EnvService) BaseEnvInitSingle(ctx context.Context, taskId int64, inst *types.InstanceInfo, auth *types.InstanceAuth) error {
	var err error
	instanceStatus := types.InstanceStatusBase
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Errorf("%s", debug.Stack())
			err = config.ErrSysPanic
		}
		var msg string
		if err != nil {
			msg = err.Error()
		}
		_ = s.CallBackSvc(ctx, taskId, inst.InstanceId, instanceStatus, msg)
	}()
	localCmd := template.GetInitBaseCmd()
	res, err := RemoteCmdExec(ctx, localCmd, _remoteBaseEnvScript, inst.IpInner, auth.UserName, auth.Pwd)
	if err != nil {
		instanceStatus = types.InstanceStatusFail
		log.Logger.Error("RemoteCmdExec", err)
		return err
	}
	if s.IsNotRemoteCmdOk(res) {
		instanceStatus = types.InstanceStatusFail
		err = errors.New("run init_base.sh err")
		log.Logger.Errorf("RemoteCmdExec:%s", res)
		return err
	}
	return nil
}

func (s *EnvService) DeployBeforeDownloadInitSingle(ctx context.Context, taskId int64, cmd string, inst *types.InstanceInfo, auth *types.InstanceAuth) error {
	var err error
	instanceStatus := types.InstanceStatusBase
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Errorf("%s", debug.Stack())
			err = config.ErrSysPanic
		}
		var msg string
		if err != nil {
			msg = err.Error()
		}
		_ = s.CallBackSvc(ctx, taskId, inst.InstanceId, instanceStatus, msg)
	}()
	if cmd == "" {
		return nil
	}
	_, err = RemoteCmdExec(ctx, cmd, _remoteBaseEnvScript, inst.IpInner, auth.UserName, auth.Pwd)
	if err != nil {
		instanceStatus = types.InstanceStatusFail
		return err
	}
	return nil
}

func (s *EnvService) ExecCmdWithUpdateInstanceStatus(ctx context.Context, taskId int64, cmd string, inst *types.InstanceInfo, auth *types.InstanceAuth, instanceStatus types.InstanceStatus) error {
	var err error
	var outs []byte
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Errorf("%s", debug.Stack())
			err = config.ErrSysPanic
		}
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		_ = s.CallBackSvc(ctx, taskId, inst.InstanceId, instanceStatus, msg)
	}()
	if cmd == "" {
		return nil
	}
	outs, err = RemoteCmdExec(ctx, cmd, _remoteBaseEnvScript, inst.IpInner, auth.UserName, auth.Pwd)
	if err != nil {
		instanceStatus = types.InstanceStatusFail
		err = fmt.Errorf("error:%v, outs:%v", err, string(outs))
		return err
	}
	return nil
}

func (s *EnvService) ServiceEnvInitAsync(ctx context.Context, svcReq *SvcEnvInitAsyncSvcReq) error {
	var err error
	s.entryLog(ctx, "ServiceEnvInitAsync", svcReq) // todo 日志脱敏
	defer func() {
		s.exitLog(ctx, "ServiceEnvInitAsync", svcReq, nil, err)
	}()
	// 机器信息入库
	if err = s.NodeUpdateStore(ctx, svcReq.InstanceList, svcReq.TaskId, svcReq.ServiceClusterId); err != nil {
		return err
	}
	// 异步
	ipList := svcReq.InstanceList
	taskId := svcReq.TaskId
	for _, nodeInfo := range ipList {
		go func(instance *types.InstanceInfo) {
			_ = s.ServiceEnvInitSingle(ctx, taskId, instance, svcReq.Auth, svcReq.Params, svcReq.Cmd)
		}(nodeInfo)
	}

	return nil
}

func (s *EnvService) ServiceEnvInitSingle(ctx context.Context, taskId int64, inst *types.InstanceInfo, auth *types.InstanceAuth, params *types.ParamsServiceEnv, cmd string) error {
	var err error
	instanceStatus := types.InstanceStatusSvc
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Errorf("%s", debug.Stack())
			err = config.ErrSysPanic
		}
		var msg string
		if err != nil {
			msg = err.Error()
		}
		_ = s.CallBackSvc(ctx, taskId, inst.InstanceId, instanceStatus, msg)
	}()

	localCmd, err := template.GetInitServiceCmd(params, cmd)
	if err != nil {
		instanceStatus = types.InstanceStatusFail
		log.Logger.Error("GetInitServiceCmd", err)
		return err
	}

	res, err := RemoteCmdExec(ctx, localCmd, _remoteServiceEnvScript, inst.IpInner, auth.UserName, auth.Pwd)
	if err != nil {
		instanceStatus = types.InstanceStatusFail
		log.Logger.Error("RemoteCmdExec", err)
		return err
	}
	if s.IsNotRemoteCmdOk(res) {
		instanceStatus = types.InstanceStatusFail
		err = errors.New("run init_base.sh err") // todo 从 res 中抽出具体问题
		log.Logger.Errorf("RemoteCmdExec:%s", res)
		return err
	}
	return nil
}

func (s *EnvService) IsNotHttpOk(res []byte) bool {
	if tool.Bytes2str(res) == "ok" {
		return true
	}
	return false
}

func (s *EnvService) IsNotRemoteCmdOk(res []byte) bool {
	if !strings.Contains(tool.Bytes2str(res), "success") {
		return true
	}
	return false
}

func (s *EnvService) CallBackSvc(ctx context.Context, taskId int64, instId string, instStatus types.InstanceStatus, msg string) error {
	var err error
	svcReq := &CallBackNodeInitSvcReq{
		Instance: &types.InstanceMeta{
			TaskId:         taskId,
			InstanceId:     instId,
			InstanceStatus: instStatus,
		},
		Msg: msg,
	}
	svc := GetNodeSvcInst()
	log.Logger.Infof("CallBackSvc req:%+v", tool.ToJson(svcReq))
	err = svc.UpdateNode(ctx, svcReq)
	if err != nil {
		log.Logger.Errorf("CallBackSvc %v", err)
		return err
	}
	return nil
}
