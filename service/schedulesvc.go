package service

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/galaxy-future/schedulx/client/healthcheckcli"
	"github.com/galaxy-future/schedulx/pkg/nodeact"
	"github.com/galaxy-future/schedulx/register/config"

	"github.com/spf13/cast"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/pkg/tool"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/register/constant"
	"github.com/galaxy-future/schedulx/repository"
	jsoniter "github.com/json-iterator/go"
)

type ScheduleSvc struct {
	Expand types.Action
	Shrink types.Action
	Deploy types.Action
}

var scheduleSvc *ScheduleSvc
var scheduleOnce sync.Once

func GetScheduleSvcInst() *ScheduleSvc {
	scheduleOnce.Do(func() {
		scheduleSvc = &ScheduleSvc{
			"Expand",
			"Shrink",
			"Deploy",
		}
	})
	return scheduleSvc
}

type ScheduleSvcReq struct {
	ServiceExpandSvcReq *ServiceExpandSvcReq
	ServiceShrinkSvcReq *ServiceShrinkSvcReq
	ServiceDeploySvcReq *ServiceDeploySvcReq
	InstrId             int64
}

type ServiceExpandSvcReq struct {
	ServiceClusterId int64  `json:"service_cluster_id'"`
	Count            int64  `json:"count"`
	ExecType         string `json:"exec_type"`
}

type ServiceShrinkSvcReq struct {
	ServiceClusterId int64  `json:"service_cluster_id'"`
	Count            int64  `json:"count"`
	ExecType         string `json:"exec_type"` // manual | auto
}

type ServiceDeploySvcReq struct {
	ServiceClusterId int64              `json:"service_cluster_id'"`
	DownloadFileUrl  string             `json:"download_file_url"`
	Count            int64              `json:"count"`
	DeployType       string             `json:"deploy_type"` // The Type of deploy : all or scroll
	FailSurge        int                `json:"fail_surge"`  // percent, 20 means 20%, valid [1, 100] . The deployment will be terminated,If the proportion of failed instances more than fail surge.
	MaxSurge         string             `json:"max_surge"`   // percent, 20 means 20%, valid [1, 100] . The ratio of rolling deployments, use ',' to separate each round.
	ExecType         string             `json:"exec_type"`
	HealthCheck      *types.HealthCheck `json:"health_check"`
	Rollback         bool               `json:"rollback"`
}

type ServiceExpandSvcResp struct {
	TaskId int64 `json:"task_id"`
}

type ServiceShrinkSvcResp struct {
	TaskId int64 `json:"task_id"`
}

type ServiceDeploySvcResp struct {
	TaskId int64 `json:"task_id"`
}

type ScheduleSvcResp struct {
	ServiceExpandSvcResp *ServiceExpandSvcResp
	ServiceShrinkSvcResp *ServiceShrinkSvcResp
	ServiceDeploySvcResp *ServiceDeploySvcResp
}

func (s *ScheduleSvc) entryLog(ctx context.Context, act string, req interface{}) {
	log.Logger.Infof("entry log | act[%s] | req:%s", act, tool.ToJson(req))
}

func (s *ScheduleSvc) exitLog(ctx context.Context, act string, req, resp interface{}, err error) {
	log.Logger.Infof("exit log | act[%s] | req:%s | resp:%s | err:%v", act, tool.ToJson(req), tool.ToJson(resp), err)
}

func (s *ScheduleSvc) ExecAct(ctx context.Context, args interface{}, act types.Action) (svcResp interface{}, err error) {
	svcReq, ok := args.(*ScheduleSvcReq)
	if !ok {
		return nil, errors.New("init service request err")
	}
	s.entryLog(ctx, string(act), svcReq)
	defer func() {
		s.exitLog(ctx, string(act), svcReq, svcResp, err)
	}()
	switch act {
	case s.Expand:
		svcResp, err = s.expandAction(ctx, svcReq.ServiceExpandSvcReq)
	case s.Shrink:
		svcResp, err = s.shrinkAction(ctx, svcReq.ServiceShrinkSvcReq)
	case s.Deploy:
		svcResp, err = s.deployAction(ctx, svcReq.ServiceDeploySvcReq)
	default:
		err = errors.New("no act matched")
	}
	return svcResp, err
}

func (s *ScheduleSvc) expandAction(ctx context.Context, svcReq *ServiceExpandSvcReq) (*ScheduleSvcResp, error) {
	var err error
	resp := &ServiceExpandSvcResp{}
	// 获取 tmpl
	tmplRepo := repository.GetScheduleTemplateRepoInst()
	schedTmpl, err := tmplRepo.GetSchedTmplBySvcClusterId(svcReq.ServiceClusterId, constant.ScheduleTypeExpand)
	if err != nil {
		return nil, err
	}
	// 创建 task
	taskRepo := repository.TaskRepo{}
	userName := ctx.Value(constant.CtxUserNameKey)
	schedTaskId, err := taskRepo.CreateTask(ctx, schedTmpl.Id, svcReq.Count, cast.ToString(userName), svcReq.ExecType, "", false)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = taskRepo.UpdateTaskStatus(ctx, schedTaskId, types.TaskStatusFail, err.Error())
			return
		}
	}()
	// 获取指令集
	var instrGroup []int64
	err = jsoniter.Unmarshal([]byte(schedTmpl.InstrGroup), &instrGroup)
	if err != nil {
		log.Logger.Error(err.Error())
		err = errors.New("instr_group unmarshal exception")
		return nil, err
	}
	// 依次执行指令集
	if err != nil {
		return nil, err
	}
	instrSvcReq := &InstrSvcReq{
		ServiceName:      schedTmpl.ServiceName,
		ServiceClusterId: schedTmpl.ServiceClusterId,
		ScheduleTaskId:   schedTaskId,
		BridgXSvcReq: &BridgXSvcReq{
			Count:       svcReq.Count,
			ClusterName: schedTmpl.BridgxClusname,
		},
		NodeActSvcReq: &NodeActSvcReq{},
	}
	go func() {
		var asyncErr error
		defer func() {
			if asyncErr != nil {
				_ = taskRepo.UpdateTaskStatus(ctx, schedTaskId, types.TaskStatusFail, asyncErr.Error())
				return
			}
			_ = taskRepo.UpdateTaskStatus(ctx, schedTaskId, types.TaskStatusSuccess, "")
		}()
		for _, instrId := range instrGroup {
			instrSvcReq.InstrId = instrId
			if asyncErr = s.doInstr(ctx, instrSvcReq); asyncErr != nil {
				log.Logger.Error("doInstr err, ", asyncErr)
				break
			}
		}
	}()
	resp.TaskId = schedTaskId
	return &ScheduleSvcResp{ServiceExpandSvcResp: resp}, nil
}

func (s *ScheduleSvc) shrinkAction(ctx context.Context, svcReq *ServiceShrinkSvcReq) (*ScheduleSvcResp, error) {
	var err error
	resp := &ServiceShrinkSvcResp{}
	// 获取 tmpl
	tmplRepo := repository.GetScheduleTemplateRepoInst()
	schedReverseTmpl, err := tmplRepo.GetSchedTmplBySvcClusterId(svcReq.ServiceClusterId, constant.ScheduleTypeShrink)
	log.Logger.Infof("shrinkAction | %+v", schedReverseTmpl)
	if err != nil {
		return nil, err
	}
	// 创建 task
	userName := ctx.Value(constant.CtxUserNameKey)
	taskRepo := repository.TaskRepo{}
	schedTaskId, err := taskRepo.CreateTask(ctx, schedReverseTmpl.Id, svcReq.Count, cast.ToString(userName), svcReq.ExecType, "", false)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = taskRepo.UpdateTaskStatus(ctx, schedTaskId, types.TaskStatusFail, err.Error())
			return
		}
	}()
	// 获取指令集
	var instrGroup []int64
	err = jsoniter.Unmarshal([]byte(schedReverseTmpl.InstrGroup), &instrGroup)
	if err != nil {
		log.Logger.Error(err.Error())
		err = errors.New("schedule_template.instr_group Unmarshal exception")
		return nil, err
	}
	/*获取扩容模板对应的最后一个成功的任务 task_id
	task, err := taskRepo.GetLastExpandSuccTask(ctx, schedReverseTmpl.ReverseSchedTmplId)
	if err != nil {
		log.Logger.Error(err.Error())
		return nil, err
	}
	relationTaskid := &db.RelationTaskId{}
	err = jsoniter.Unmarshal([]byte(task.RelationTaskId), relationTaskid)
	if err != nil {
		log.Logger.Error("relationTaskid unmarshal error | ", task.RelationTaskId, err.Error())
		return nil, err
	}*/
	// 依次执行指令集
	instrSvcReq := &InstrSvcReq{
		ServiceName:      schedReverseTmpl.ServiceName,
		ServiceClusterId: schedReverseTmpl.ServiceClusterId,
		ScheduleTaskId:   schedTaskId,
		BridgXSvcReq: &BridgXSvcReq{
			//TaskId:      relationTaskid.BridgxTaskId,
			ClusterName: schedReverseTmpl.BridgxClusname,
			Count:       svcReq.Count,
		},
		NodeActSvcReq: &NodeActSvcReq{
			//TaskId: relationTaskid.NodeactTaskId,
			UmountSlbSvcReq: &UmountSlbSvcReq{
				UmountInstCnt: svcReq.Count,
			},
		},
	}
	go func() {
		var asyncErr error
		defer func() {
			if asyncErr != nil {
				_ = taskRepo.UpdateTaskStatus(ctx, schedTaskId, types.TaskStatusFail, asyncErr.Error())
				return
			}
			_ = taskRepo.UpdateTaskStatus(ctx, schedTaskId, types.TaskStatusSuccess, "")
		}()
		for _, instrId := range instrGroup {
			instrSvcReq.InstrId = instrId
			if asyncErr = s.doInstr(ctx, instrSvcReq); asyncErr != nil {
				log.Logger.Error("doInstr err, ", asyncErr)
				break
			}
		}
	}()
	resp.TaskId = schedTaskId
	return &ScheduleSvcResp{ServiceShrinkSvcResp: resp}, nil
}

func (s *ScheduleSvc) deployAction(ctx context.Context, svcReq *ServiceDeploySvcReq) (svcResp interface{}, err error) {
	switch svcReq.DeployType {
	case constant.TaskDeployTypeAll:
		svcResp, err = s.deployActionForAll(ctx, svcReq)
	case constant.TaskDeployTypeScroll:
		svcResp, err = s.deployActionForScroll(ctx, svcReq)
	default:
		err = errors.New("no deploy type matched")
	}
	return svcResp, err
}

func (s *ScheduleSvc) deployActionForAll(ctx context.Context, svcReq *ServiceDeploySvcReq) (*ScheduleSvcResp, error) {
	var err error
	resp := &ServiceDeploySvcResp{}
	// 获取 tmpl
	tmplRepo := repository.GetScheduleTemplateRepoInst()
	tmpl, err := tmplRepo.GetSchedTmplBySvcClusterId(svcReq.ServiceClusterId, constant.ScheduleTypeDeploy)
	log.Logger.Infof("deployAction | %+v", tmpl)
	if err != nil {
		return nil, err
	}
	// 创建 task
	userName := ctx.Value(constant.CtxUserNameKey)
	taskRepo := repository.TaskRepo{}
	taskInfo, _ := jsoniter.MarshalToString(svcReq)
	schedTaskId, err := taskRepo.CreateTask(ctx, tmpl.Id, svcReq.Count, cast.ToString(userName), svcReq.ExecType, taskInfo, svcReq.Rollback)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = taskRepo.UpdateTaskStatus(ctx, schedTaskId, types.TaskStatusFail, err.Error())
			return
		}
	}()
	// 获取指令集
	var instrGroup []int64
	err = jsoniter.Unmarshal([]byte(tmpl.InstrGroup), &instrGroup)
	if err != nil {
		log.Logger.Error(err.Error())
		err = errors.New("schedule_template.instr_group Unmarshal exception")
		return nil, err
	}

	// 依次执行指令集
	instrSvcReq := &InstrSvcReq{
		ServiceName:      tmpl.ServiceName,
		ServiceClusterId: tmpl.ServiceClusterId,
		ScheduleTaskId:   schedTaskId,
		NodeActSvcReq: &NodeActSvcReq{
			TaskId:           schedTaskId,
			ServiceClusterId: svcReq.ServiceClusterId,
			DownloadFileUrl:  svcReq.DownloadFileUrl,
			InstanceCount:    svcReq.Count,
		},
	}
	go func() {
		var asyncErr error
		defer func() {
			if r := recover(); r != nil {
				log.Logger.Errorf("%s", debug.Stack())
				asyncErr = config.ErrSysPanic
			}
			if asyncErr != nil {
				_ = taskRepo.UpdateTaskStatus(ctx, schedTaskId, types.TaskStatusFail, asyncErr.Error())
				return
			}
			_ = taskRepo.UpdateTaskStatus(ctx, schedTaskId, types.TaskStatusSuccess, "")
		}()
		for _, instrId := range instrGroup {
			instrSvcReq.InstrId = instrId
			if asyncErr = s.doInstr(ctx, instrSvcReq); asyncErr != nil {
				log.Logger.Error("doInstr err: ", asyncErr)
				break
			}
		}
	}()
	resp.TaskId = schedTaskId
	return &ScheduleSvcResp{ServiceDeploySvcResp: resp}, nil
}

func (s *ScheduleSvc) deployActionForScroll(ctx context.Context, svcReq *ServiceDeploySvcReq) (*ScheduleSvcResp, error) {
	var err error
	resp := &ServiceDeploySvcResp{}
	tmplRepo := repository.GetScheduleTemplateRepoInst()
	tmpl, err := tmplRepo.GetSchedTmplBySvcClusterId(svcReq.ServiceClusterId, constant.ScheduleTypeDeploy)
	log.Logger.Infof("deployAction | %+v", tmpl)
	if err != nil {
		return nil, err
	}

	userName := ctx.Value(constant.CtxUserNameKey)
	taskRepo := repository.TaskRepo{}
	taskInfo, _ := jsoniter.MarshalToString(svcReq)
	scheduleTaskId, err := taskRepo.CreateTask(ctx, tmpl.Id, svcReq.Count, cast.ToString(userName), svcReq.ExecType, taskInfo, svcReq.Rollback)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = taskRepo.UpdateTaskStatus(ctx, scheduleTaskId, types.TaskStatusFail, err.Error())
			return
		}
	}()

	// Create subtask for the given task.
	ret, err := getBridgxInstances(ctx, &InstrSvcReq{ServiceClusterId: tmpl.ServiceClusterId})
	if err != nil {
		return nil, err
	}
	subTaskRepo := repository.SubTaskRepo{}
	instanceList := ret.NodeActSvcResp.InstGroup.InstanceList
	total := uint64(len(instanceList))
	scheduleSubTaskList, err := subTaskRepo.CreateSubTask(instanceList, svcReq.MaxSurge, scheduleTaskId)
	if err != nil {
		return nil, err
	}

	var instrGroup []int64
	err = jsoniter.Unmarshal([]byte(tmpl.InstrGroup), &instrGroup)
	if err != nil {
		log.Logger.Error(err.Error())
		err = errors.New("schedule_template.instr_group Unmarshal exception")
		return nil, err
	}
	instrSvcReq := &InstrSvcReq{
		ServiceName:      tmpl.ServiceName,
		ServiceClusterId: tmpl.ServiceClusterId,
		NodeActSvcReq: &NodeActSvcReq{
			ServiceClusterId: svcReq.ServiceClusterId,
			DownloadFileUrl:  svcReq.DownloadFileUrl,
			Auth:             ret.NodeActSvcResp.Auth,
		},
	}

	go func() {
		var asyncErr error
		var failNum uint64 = 0
		needRollBack := false
		taskStatus := types.TaskStatusSuccess
		defer func() {
			if r := recover(); r != nil {
				log.Logger.Errorf("%s", debug.Stack())
				asyncErr = config.ErrSysPanic
			}
			if asyncErr != nil {
				taskStatus = types.TaskStatusFail
				if needRollBack {
					taskStatus = types.TaskStatusRollingBack
				}
				_ = taskRepo.UpdateTaskStatus(ctx, scheduleTaskId, taskStatus, asyncErr.Error())
				return
			}
			_ = taskRepo.UpdateTaskStatus(ctx, scheduleTaskId, taskStatus, "")
		}()

		// batch processing subtasks
		log.Logger.Info("start deploy instance async")
		for _, subTask := range scheduleSubTaskList {
			var wg sync.WaitGroup
			var instanceList []*types.InstanceInfo
			_ = jsoniter.UnmarshalFromString(subTask.InstanceList, &instanceList)
			subTaskId := subTask.Id
			instrSvcReq.NodeActSvcReq.InstanceCount = int64(len(instanceList))
			instrSvcReq.NodeActSvcReq.TaskId = subTaskId
			instrSvcReq.ScheduleTaskId = subTaskId
			_ = subTaskRepo.UpdateSubTaskStatus(ctx, subTaskId, types.TaskStatusRunning, "")

			// process all instances under subtasks
			instanceGroup := &nodeact.InstanceGroup{TaskId: ret.NodeActSvcResp.InstGroup.TaskId}
			for _, instInfo := range instanceList {
				wg.Add(1)
				instanceGroup.InstanceList = []*types.InstanceInfo{instInfo}
				instanceId := instInfo.InstanceId
				instrSvcReq.InstanceTaskId = instanceId
				instrSvcReq.NodeActSvcReq.InstGroup = instanceGroup
				log.Logger.Infof("async scroll deploy instance instanceid:%s", instanceId)

				// Execute all instruction sets of each machine in turn
				instInfo := instInfo
				go func(instance *types.InstanceInfo) {
					defer func() {
						wg.Done()
					}()
					for _, instrId := range instrGroup {
						instrSvcReq.InstrId = instrId
						if asyncErr = s.doInstrForScrollDeploy(ctx, instrSvcReq); asyncErr != nil {
							atomic.AddUint64(&failNum, 1)
							log.Logger.Error("doInstr err: ", asyncErr)
							break
						}
					}
					if asyncErr == nil {
						if checkErr := healthcheckcli.GetHealthCheckXCli(ctx).HealthCheck(ctx, svcReq.HealthCheck, instInfo); checkErr != nil {
							_, asyncErr = repository.GetInstanceRepoIns().UpdateStatus(ctx, types.InstanceStatusHealthCheckFail, subTaskId, instance.IpInner)
							return
						}
						_, asyncErr = repository.GetInstanceRepoIns().UpdateStatus(ctx, types.InstanceStatusHealthCheck, subTaskId, instance.IpInner)
					}
				}(instInfo)
			}
			// wait for all instances in a subtask to be processed
			wg.Wait()
			if r := recover(); r != nil {
				log.Logger.Errorf("%s", debug.Stack())
				asyncErr = config.ErrSysPanic
			}
			if asyncErr != nil {
				_ = subTaskRepo.UpdateSubTaskStatus(ctx, subTaskId, types.TaskStatusFail, asyncErr.Error())
			} else {
				_ = subTaskRepo.UpdateSubTaskStatus(ctx, subTaskId, taskStatus, "")
			}
			if failNum*100/total >= uint64(svcReq.FailSurge) {
				needRollBack = true
				break
			}
		}
		log.Logger.Info("end scroll deploy instance async")
	}()

	resp.TaskId = scheduleTaskId
	return &ScheduleSvcResp{ServiceDeploySvcResp: resp}, nil
}

func (s *ScheduleSvc) doInstr(ctx context.Context, instrSvcReq *InstrSvcReq) error {
	var err error
	defer func() {
		if err != nil {
			log.Logger.Errorf("doInstr instrSvcReq:%s, %v", tool.ToJson(instrSvcReq), err)
		}
	}()

	instrRepo := repository.GetInstrRepoInst()
	instrument, err := instrRepo.GetInstr(ctx, instrSvcReq.InstrId)
	if err != nil {
		return err
	}
	instrSvcReq.Instruction = instrument
	instrSvc := GetInstrSvcInst()
	svcResp, err := instrSvc.ExecAct(ctx, instrSvcReq, instrument.InstrAction)
	if err != nil {
		return err
	}
	instrSvcResp := svcResp.(*InstrSvcResp)
	taskRepo := repository.GetTaskRepoInst()
	instanceGroup := instrSvcResp.NodeActSvcResp.InstGroup
	switch instrument.InstrAction {
	case instrSvc.BridgXExpand: // 给下一轮赋值参数
		instrSvcReq.NodeActSvcReq.InstGroup = instrSvcResp.BridgXSvcResp.InstGroup
		instrSvcReq.NodeActSvcReq.Auth = instrSvcResp.BridgXSvcResp.Auth
		err = taskRepo.UpdateTaskRelationTaskId(ctx, instrSvcReq.ScheduleTaskId, types.BridgXTaskId, instrSvcResp.BridgXSvcResp.TaskId)
	case instrSvc.NodeActInitBase:
		instrSvcReq.NodeActSvcReq.InstGroup = instanceGroup
		err = taskRepo.UpdateTaskRelationTaskId(ctx, instrSvcReq.ScheduleTaskId, types.NodeactTaskId, instanceGroup.TaskId)
	case instrSvc.NodeActInitSvc:
		instrSvcReq.NodeActSvcReq.InstGroup = instanceGroup
	case instrSvc.MountSLB:
	case instrSvc.UmountSLB:
		instrSvcReq.BridgXSvcReq.InstGroup = instanceGroup
	case instrSvc.BridgXShrink:
	case instrSvc.MountNginx:
	case instrSvc.UmountNginx:
	case instrSvc.NodeActBeforeDownload:
		instrSvcReq.NodeActSvcReq.InstGroup = instrSvcResp.NodeActSvcResp.InstGroup
		instrSvcReq.NodeActSvcReq.Auth = instrSvcResp.NodeActSvcResp.Auth
	case instrSvc.NodeActDownload:
	case instrSvc.NodeActBeforeDeploy:
	case instrSvc.NodeActDeploy:
	case instrSvc.NodeActAfterDeploy:
	default:
		err = fmt.Errorf("instr.action invalid:%s", instrument.InstrAction)
	}
	return err
}

func (s *ScheduleSvc) doInstrForScrollDeploy(ctx context.Context, instrSvcReq *InstrSvcReq) error {
	var err error
	defer func() {
		if err != nil {
			log.Logger.Errorf("doInstrForScrollDeploy instrSvcReq:%s, %v", tool.ToJson(instrSvcReq), err)
		}
	}()

	instrRepo := repository.GetInstrRepoInst()
	instrument, err := instrRepo.GetInstr(ctx, instrSvcReq.InstrId)
	if err != nil {
		return err
	}
	instrSvcReq.Instruction = instrument
	instrSvc := GetInstrSvcInst()
	_, err = instrSvc.ExecActForScrollDeploy(ctx, instrSvcReq, instrument.InstrAction)
	if err != nil {
		return err
	}
	return err
}
