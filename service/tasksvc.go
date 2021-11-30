package service

import (
	"context"
	"sync"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/pkg/tool"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/repository"
	"github.com/galaxy-future/schedulx/repository/model/db"
)

type TaskService struct {
}

var taskServiceInstance *TaskService
var taskOnce sync.Once

func GetTaskSvcInst() *TaskService {
	taskOnce.Do(func() {
		taskServiceInstance = &TaskService{}
	})
	return taskServiceInstance
}

type TaskDescribeSvcReq struct {
	TaskId         int64                `json:"task_id"`
	InstanceStatus types.InstanceStatus `json:"instance_status"`
}

type TaskDescribeSvcResp struct {
	TaskDescribe *types.TaskDescribe `json:"task_describe"`
}

type TaskInstancesSvcReq struct {
	TaskId         int64                `json:"task_id"`
	InstanceStatus types.InstanceStatus `json:"instance_status"`
	PageNumber     int                  `json:"page_number"`
	PageSize       int                  `json:"page_size"`
}

type TaskInstancesSvcResp struct {
	InstancesList []*types.Instance `json:"instances_list"`
	Pager         *types.Pager      `json:"pager"`
}

type TaskInfoSvcReq struct {
	TaskId int64 `json:"task_id"`
}

type TaskInfoSvcResp struct {
	TaskInfo *types.TaskInfo `json:"task_info"`
}

func (s *TaskService) entryLog(ctx context.Context, method string, req interface{}) {
	log.Logger.Infof("entry log | method[%s] | req:%s", method, tool.ToJson(req))
}

func (s *TaskService) exitLog(ctx context.Context, method string, req, resp interface{}, err error) {
	log.Logger.Infof("exit log | method[%s] | req:%s | resp:%s | err:%v", method, tool.ToJson(req), tool.ToJson(resp), err)
}

func (s *TaskService) Describe(ctx context.Context, svcReq *TaskDescribeSvcReq) (*TaskDescribeSvcResp, error) {
	var err error
	resp := &TaskDescribeSvcResp{
		TaskDescribe: &types.TaskDescribe{},
	}
	s.entryLog(ctx, "Describe", svcReq) // todo 日志脱敏
	defer func() {
		s.exitLog(ctx, "Describe", svcReq, resp, err)
	}()
	repo := repository.GetInstanceRepoIns()
	fields := []string{
		"id",
		"instance_status",
		"ip_inner",
	}
	insts, err := repo.InstsQueryByTaskId(ctx, svcReq.TaskId, "", fields) //todo 改为查 nodeact_task 表
	if err != nil {
		log.Logger.Error("InstsQueryByTaskId", err)
		return nil, err
	}
	if len(insts) != 0 {
		row := &db.Instance{}
		err = db.Get(insts[0].Id, row)
		if err != nil {
			log.Logger.Error("db Get", err)
			return nil, err
		}
		resp.TaskDescribe.FoundTime = row.CreateAt
	}

	var successNum, failNum int64
	for _, inst := range insts {
		if inst.InstanceStatus == svcReq.InstanceStatus {
			successNum++
			continue
		}
		if inst.InstanceStatus == types.InstanceStatusFail {
			failNum++
			continue
		}
	}
	totalNum := int64(len(insts))
	resp.TaskDescribe.FailNum = failNum
	resp.TaskDescribe.SuccessNum = successNum
	resp.TaskDescribe.TotalNum = totalNum
	resp.TaskDescribe.SuccessRate = tool.FormatFloat(float64(successNum/totalNum), 2)
	return resp, nil
}

func (s *TaskService) Instances(ctx context.Context, svcReq *TaskInstancesSvcReq) (*TaskInstancesSvcResp, error) {
	var err error
	resp := &TaskInstancesSvcResp{}
	s.entryLog(ctx, "Instances", svcReq) // todo 日志脱敏
	defer func() {
		s.exitLog(ctx, "Instances", svcReq, resp, err)
	}()
	repo := repository.GetInstanceRepoIns()
	fields := []string{
		"instance_id",
		"ip_inner",
		"ip_outer",
	}
	insts, count, err := repo.InstsQueryByPage(ctx, svcReq.TaskId, svcReq.InstanceStatus, svcReq.PageSize, svcReq.PageNumber, fields)
	if err != nil {
		log.Logger.Error("InstsQueryByTaskId", err)
		return nil, err
	}
	instances := make([]*types.Instance, 0, len(insts))
	for _, item := range insts {
		ins := &types.Instance{
			InstanceId: item.InstanceId,
			IpInner:    item.IpInner,
			IpOuter:    item.IpOuter,
			Status:     svcReq.InstanceStatus,
		}
		instances = append(instances, ins)
	}
	pager := &types.Pager{
		PagerNum:  svcReq.PageNumber,
		PagerSize: svcReq.PageSize,
		Total:     int(count),
	}
	resp.InstancesList = instances
	resp.Pager = pager
	return resp, nil
}

func (s *TaskService) Info(ctx context.Context, svcReq *TaskInfoSvcReq) (*TaskInfoSvcResp, error) {
	var err error
	repo := repository.GetTaskRepoInst()
	task, err := repo.GetTask(ctx, svcReq.TaskId)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	return &TaskInfoSvcResp{
		TaskInfo: &types.TaskInfo{
			TaskStatus: task.TaskStatus,
			TaskStep:   task.TaskStep,
			InstCnt:    task.InstCnt,
			Msg:        task.Msg,
			Operator:   task.Operator,
			ExecType:   task.ExecType,
		},
	}, nil
}

func (s *TaskService) InstanceList(ctx context.Context, page, pageSize, taskId int, taskStatus types.InstanceStatus) (int64, []types.InstInfoResp, error) {
	fields := []string{"instance_id", "ip_inner", "ip_outer", "instance_status"}
	if page == 0 {
		page = 1
	}
	if pageSize == 0 || pageSize > 500 {
		pageSize = 500
	}
	list, count, err := repository.GetInstanceRepoIns().InstsQueryByPage(ctx, int64(taskId), taskStatus, pageSize, page, fields)
	if err != nil {
		log.Logger.Errorf("error:%v", err)
		return 0, nil, err
	}

	instanceInfo := make([]types.InstInfoResp, len(list))
	for _, item := range list {
		info := types.InstInfoResp{
			InstanceId: item.InstanceId,
			IpInner:    item.IpInner,
			IpOuter:    item.IpOuter,
			Status:     item.InstanceStatus,
		}
		instanceInfo = append(instanceInfo, info)
	}
	return count, instanceInfo, nil
}
