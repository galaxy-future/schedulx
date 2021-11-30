package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/spf13/cast"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/pkg/tool"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/register/constant"
	"github.com/galaxy-future/schedulx/repository"
	"github.com/galaxy-future/schedulx/repository/model/db"
	jsoniter "github.com/json-iterator/go"
)

type ScheduleSvc struct {
	Expand types.Action
	Shrink types.Action
}

var scheduleSvc *ScheduleSvc
var scheduleOnce sync.Once

func GetScheduleSvcInst() *ScheduleSvc {
	scheduleOnce.Do(func() {
		scheduleSvc = &ScheduleSvc{
			"Expand",
			"Shrink",
		}
	})
	return scheduleSvc
}

type ScheduleSvcReq struct {
	ServiceExpandSvcReq *ServiceExpandSvcReq
	ServiceShrinkSvcReq *ServiceShrinkSvcReq
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

type ServiceExpandSvcResp struct {
	TaskId int64 `json:"task_id"`
}

type ServiceShrinkSvcResp struct {
	TaskId int64 `json:"task_id"`
}

type ScheduleSvcResp struct {
	ServiceExpandSvcResp *ServiceExpandSvcResp
	ServiceShrinkSvcResp *ServiceShrinkSvcResp
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
	schedTaskId, err := taskRepo.CreateTask(ctx, schedTmpl.Id, svcReq.Count, cast.ToString(userName), svcReq.ExecType)
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
		ServiceName:    schedTmpl.ServiceName,
		ScheduleTaskId: schedTaskId,
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
	schedTaskId, err := taskRepo.CreateTask(ctx, schedReverseTmpl.Id, svcReq.Count, cast.ToString(userName), svcReq.ExecType)
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
	// 获取扩容模板对应的最后一个成功的任务 task_id
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
	}
	// 依次执行指令集
	instrSvcReq := &InstrSvcReq{
		ServiceName:    schedReverseTmpl.ServiceName,
		ScheduleTaskId: schedTaskId,
		BridgXSvcReq: &BridgXSvcReq{
			TaskId:      relationTaskid.BridgxTaskId,
			ClusterName: schedReverseTmpl.BridgxClusname,
		},
		NodeActSvcReq: &NodeActSvcReq{
			TaskId: relationTaskid.NodeactTaskId,
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
	switch instrument.InstrAction {
	case instrSvc.BridgXExpand: // 给下一轮赋值参数
		instrSvcReq.NodeActSvcReq.InstGroup = instrSvcResp.BridgXSvcResp.InstGroup
		instrSvcReq.NodeActSvcReq.Auth = instrSvcResp.BridgXSvcResp.Auth
		err = taskRepo.UpdateTaskRelationTaskId(ctx, instrSvcReq.ScheduleTaskId, "bridgx_task_id", instrSvcResp.BridgXSvcResp.TaskId)
	case instrSvc.NodeActInitBase:
		instrSvcReq.NodeActSvcReq.InstGroup = instrSvcResp.NodeActSvcResp.InstGroup
		err = taskRepo.UpdateTaskRelationTaskId(ctx, instrSvcReq.ScheduleTaskId, "nodeact_task_id", instrSvcResp.NodeActSvcResp.InstGroup.TaskId)
	case instrSvc.NodeActInitSvc:
		instrSvcReq.NodeActSvcReq.InstGroup = instrSvcResp.NodeActSvcResp.InstGroup
	case instrSvc.MountSLB:
	case instrSvc.UmountSLB:
		instrSvcReq.BridgXSvcReq.InstGroup = instrSvcResp.NodeActSvcResp.InstGroup
	case instrSvc.BridgXShrink:
	case instrSvc.MountNginx:
	case instrSvc.UmountNginx:
	default:
		err = fmt.Errorf("instr.action invalid:%s", instrument.InstrAction)
	}
	return err
}
