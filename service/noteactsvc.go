package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/pkg/nodeact"
	"github.com/galaxy-future/schedulx/pkg/tool"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/spf13/cast"
)

type NodeActSvc struct {
	InitBase             types.Action
	InitService          types.Action
	MountSlb             types.Action
	UmountSlb            types.Action
	PollQueryInitBase    types.Action
	PollQueryInitService types.Action
}

var nodeActSvc *NodeActSvc
var nodeActOnce sync.Once

func GetNodeActSvcInst() *NodeActSvc {
	nodeActOnce.Do(func() {
		nodeActSvc = &NodeActSvc{
			"init_base",
			"init_service",
			"mount_slb",
			"umount_slb",
			"poll_query_node_base",
			"poll_query_node_svc",
		}
	})
	return nodeActSvc
}

type NodeActSvcReq struct {
	InstGroup        *nodeact.InstanceGroup
	TaskId           int64
	ServiceClusterId int64
	Auth             *types.InstanceAuth
	//HarborRegisterUrl string // harbor 镜像服务地址，用于加入 docker 的配置允许 http 链接
	InitServicSvcReq *InitServicSvcReq
	//Params          *nodeact.ParamsServiceEnv
	SlbMountInfo    *nodeact.ParamsMountInfo // 用于挂载 slb 的服务器组 id
	UmountSlbSvcReq *UmountSlbSvcReq
}
type InitServicSvcReq struct {
	Cmd    string
	Params *types.ParamsServiceEnv
}
type NodeActSvcResp struct {
	InstGroup *nodeact.InstanceGroup
}

type UmountSlbSvcReq struct {
	UmountInstCnt int64                    `json:"umount_inst_cnt"`
	SlbInfo       *nodeact.ParamsMountInfo `json:"slb_info"`
}

func (s *NodeActSvc) entryLog(ctx context.Context, act string, req interface{}) {
	log.Logger.Infof("entry log | act[%s] | req:%s", act, tool.ToJson(req))
}

func (s *NodeActSvc) exitLog(ctx context.Context, act string, req, resp interface{}, err error) {
	log.Logger.Infof("exit log | act[%s] | req:%s | resp:%s | err:%v", act, tool.ToJson(req), tool.ToJson(resp), err)
}

func (s *NodeActSvc) ExecAct(ctx context.Context, args interface{}, act types.Action) (svcResp interface{}, err error) {
	svcReq, ok := args.(*NodeActSvcReq)
	if !ok {
		return nil, errors.New("init service request err")
	}
	s.entryLog(ctx, string(act), svcReq)
	defer func() {
		s.exitLog(ctx, string(act), svcReq, svcResp, err)
	}()
	switch act {
	case s.InitBase:
		svcResp, err = s.InitBaseAction(ctx, svcReq.InstGroup, svcReq.Auth, svcReq.ServiceClusterId)
	case s.InitService:
		svcResp, err = s.InitServiceAction(ctx, svcReq.InstGroup, svcReq.Auth, svcReq.InitServicSvcReq, svcReq.ServiceClusterId)
	case s.MountSlb:
		svcResp, err = s.MountInstAction(ctx, svcReq.InstGroup, svcReq.SlbMountInfo)
	case s.UmountSlb:
		svcResp, err = s.UmountInstAction(ctx, svcReq.TaskId, svcReq.ServiceClusterId, svcReq.UmountSlbSvcReq)
	case s.PollQueryInitBase:
		svcResp, err = s.PollQueryBaseNode(ctx, svcReq.TaskId)
	case s.PollQueryInitService:
		svcResp, err = s.PoolQueryServiceNode(ctx, svcReq.TaskId)
	default:
		err = errors.New("no act matched")
	}
	return svcResp, err
}

func (s *NodeActSvc) InitBaseAction(ctx context.Context, instGroup *nodeact.InstanceGroup, auth *types.InstanceAuth, serviceClusterId int64) (*NodeActSvcResp, error) {
	var err error
	resp := &NodeActSvcResp{}

	baseEnvReq := &BaseEnvInitAsyncSvcReq{
		ServiceClusterId: serviceClusterId,
		TaskId:           instGroup.TaskId,
		InstanceList:     instGroup.InstanceList,
		Auth:             auth,
	}
	err = GetEnvOpsSvcInst().BaseEnvInitAsync(ctx, baseEnvReq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *NodeActSvc) InitServiceAction(ctx context.Context, instGroup *nodeact.InstanceGroup, auth *types.InstanceAuth, svcReq *InitServicSvcReq, serviceClusterId int64) (*NodeActSvcResp, error) {
	var err error
	resp := &NodeActSvcResp{}

	svcEnvReq := &SvcEnvInitAsyncSvcReq{
		TaskId:           instGroup.TaskId,
		ServiceClusterId: serviceClusterId,
		InstanceList:     instGroup.InstanceList,
		Auth:             auth,
		Params:           svcReq.Params,
		Cmd:              svcReq.Cmd,
	}
	err = GetEnvOpsSvcInst().ServiceEnvInitAsync(ctx, svcEnvReq)
	return resp, err
}

func (s *NodeActSvc) MountInstAction(ctx context.Context, instGroup *nodeact.InstanceGroup, slbMountInfo *nodeact.ParamsMountInfo) (*ExposeMountSvcResp, error) {
	var err error
	exposeResp := &ExposeMountSvcResp{}

	mountReq := &ExposeMountSvcReq{
		MountType:    slbMountInfo.MountType,
		MountValue:   slbMountInfo.MountValue,
		TaskId:       instGroup.TaskId,
		InstanceList: instGroup.InstanceList,
	}
	exposeResp, err = GetMountSvcInst().Mount(ctx, mountReq)
	return exposeResp, err
}

func (s *NodeActSvc) UmountInstAction(ctx context.Context, taskId, serviceClusterId int64, umountSlbSvcReq *UmountSlbSvcReq) (*ExposeUmountSvcResp, error) {
	var err error
	resp := &ExposeUmountSvcResp{}

	unmountReq := &ExposeUmountSvcReq{
		ServiceClusterId: serviceClusterId,
		TaskId:           taskId,
		Count:            umountSlbSvcReq.UmountInstCnt,
		MountType:        umountSlbSvcReq.SlbInfo.MountType,
		MountValue:       umountSlbSvcReq.SlbInfo.MountValue,
	}
	resp, err = GetMountSvcInst().Umount(ctx, unmountReq)
	return resp, err
}

func (s *NodeActSvc) PollQueryBaseNode(ctx context.Context, taskId int64) (*NodeActSvcResp, error) {
	return s.poolQueryTask(ctx, taskId, types.InstanceStatusBase)
}

func (s *NodeActSvc) PoolQueryServiceNode(ctx context.Context, taskId int64) (*NodeActSvcResp, error) {
	return s.poolQueryTask(ctx, taskId, types.InstanceStatusSvc)
}

func (s *NodeActSvc) poolQueryTask(ctx context.Context, taskId int64, insStatus types.InstanceStatus) (*NodeActSvcResp, error) {
	var err error
	resp := &NodeActSvcResp{}

	taskDescReq := &TaskDescribeSvcReq{
		TaskId:         taskId,
		InstanceStatus: insStatus,
	}
	queryTaskC := make(chan bool, 1)
	// 轮询 task 进度
	time.Sleep(20 * time.Second)
	//var taskDescribe *nodeact.TaskDescribe
	timeWait := 5 * time.Minute
	go func() {
		st := time.Now()
		errCnt := 0
		for {
			if errCnt >= 3 {
				log.Logger.Error("too many request err")
				queryTaskC <- false
				return
			}
			time.Sleep(3 * time.Second)

			taskDescRet, err := GetTaskSvcInst().Describe(ctx, taskDescReq)
			if err != nil {
				log.Logger.Errorf("func GetTaskSvcInst().Describe  error:%v", err)
			}
			if s.RatePass(taskDescRet.TaskDescribe.SuccessRate) {
				queryTaskC <- true
				return
			}
			end := time.Now()
			if end.After(st.Add(timeWait + 2*time.Second)) { // time out
				queryTaskC <- false
				return
			}
		}
	}()

	select {
	case taskDone := <-queryTaskC:
		if !taskDone {
			err = errors.New("something wrong in expand task")
		}
	case <-time.After(timeWait):
		err = errors.New("pool query timeout")
	}
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	InstGroup := &nodeact.InstanceGroup{
		TaskId: taskId,
	}
	var InstanceList []*types.InstanceInfo
	// 查询 ip
	instCnt := 0
	pageNum := 1
	for {
		taskInsReq := &TaskInstancesSvcReq{
			TaskId:         taskId,
			InstanceStatus: insStatus,
			PageNumber:     pageNum,
			PageSize:       50,
		}
		taskInstancesResp, err := GetTaskSvcInst().Instances(ctx, taskInsReq)
		if err != nil {
			log.Logger.Errorf("func GetTaskSvcInst().Instances params:%+v error:%v", *taskInsReq, err)
		}
		log.Logger.Infof("instances: taskInstancesData:%v", tool.ToJson(taskInstancesResp))
		if pageNum == 1 {
			InstanceList = make([]*types.InstanceInfo, 0, taskInstancesResp.Pager.Total)
		}
		if pageNum == 1 && len(taskInstancesResp.InstancesList) == 0 {
			err = errors.New("no instances found")
			log.Logger.Error(err)
			return nil, err
		}
		for _, item := range taskInstancesResp.InstancesList {
			inst := &types.InstanceInfo{
				IpInner:    item.IpInner,
				IpOuter:    item.IpOuter,
				InstanceId: item.InstanceId,
			}
			InstanceList = append(InstanceList, inst)
		}
		instCnt += len(taskInstancesResp.InstancesList)
		if instCnt >= taskInstancesResp.Pager.Total {
			break
		}
		pageNum++
	}
	InstGroup.InstanceList = InstanceList
	resp.InstGroup = InstGroup

	return resp, nil
}

func (s *NodeActSvc) RatePass(rate string) bool {
	if cast.ToFloat64(rate) >= 1 {
		return true
	}
	return false
}
