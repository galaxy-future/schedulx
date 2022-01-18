package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/client/bridgxcli"
	"github.com/galaxy-future/schedulx/pkg/bridgx"
	"github.com/galaxy-future/schedulx/pkg/nodeact"
	"github.com/galaxy-future/schedulx/pkg/tool"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/repository"
	"github.com/spf13/cast"
)

type BridgXSvc struct {
	Expand          types.Action
	PoolQueryExpand types.Action
	Shrink          types.Action
	PoolQueryShrink types.Action
	GetCluster      types.Action
}

var bridgXSvc *BridgXSvc
var bridgXOnce sync.Once

func GetBridgXSvcInst() *BridgXSvc {
	bridgXOnce.Do(func() {
		bridgXSvc = &BridgXSvc{
			"expand",
			"pool_query_expand",
			"shrink",
			"pool_query_shrink",
			"get_cluster",
		}
	})
	return bridgXSvc
}

type BridgXSvcReq struct {
	Count       int64
	ClusterName string
	TaskId      int64
	InstGroup   *nodeact.InstanceGroup
}

type BridgXSvcResp struct {
	TaskId    int64
	InstGroup *nodeact.InstanceGroup
	Auth      *types.InstanceAuth
}

func (s *BridgXSvc) entryLog(ctx context.Context, method string, req interface{}) {
	log.Logger.Infof("entry log | method[%s] | req:%s", method, tool.ToJson(req))
}

func (s *BridgXSvc) exitLog(ctx context.Context, method string, req, resp interface{}, err error) {
	log.Logger.Infof("exit log | method[%s] | req:%s | resp:%s | err:%v", method, tool.ToJson(req), tool.ToJson(resp), err)
}

func (s *BridgXSvc) ExecAct(ctx context.Context, args interface{}, act types.Action) (resp interface{}, err error) {
	log.Logger.Infof("args:%+v", args)
	svcReq, ok := args.(*BridgXSvcReq)
	if !ok {
		return nil, errors.New("init service request err")
	}
	s.entryLog(ctx, string(act), svcReq)
	defer func() {
		s.exitLog(ctx, string(act), svcReq, resp, err)
	}()
	switch act {
	case s.Expand:
		resp, err = s.expandAction(ctx, svcReq.ClusterName, svcReq.Count)
	case s.PoolQueryExpand:
		resp, err = s.pollQueryExpandAction(ctx, svcReq.TaskId)
	case s.Shrink:
		resp, err = s.shrinkAction(ctx, svcReq.TaskId, svcReq.ClusterName, svcReq.InstGroup)
	case s.PoolQueryShrink:
		resp, err = s.pollQueryShrinkAction(ctx, svcReq.TaskId, svcReq.InstGroup)
	case s.GetCluster:
		resp, err = s.getClusterAction(ctx, svcReq.ClusterName)
	default:
		err = errors.New("no act matched")
	}

	return resp, err
}

func (s *BridgXSvc) expandAction(ctx context.Context, clusterName string, count int64) (*BridgXSvcResp, error) {
	var err error
	resp := &BridgXSvcResp{}
	bCli := bridgxcli.GetBridgXCli(ctx)
	cliReq := &bridgxcli.ClusterExpandReq{
		TaskName:    fmt.Sprintf("schedulx 扩容 %v 台", count),
		ClusterName: clusterName,
		Count:       count,
	}
	httpResp, err := bCli.ClusterExpand(ctx, cliReq)
	if err != nil {
		return nil, err
	}
	resp.TaskId = httpResp.Data
	return resp, err
}

func (s *BridgXSvc) pollQueryExpandAction(ctx context.Context, taskId int64) (*BridgXSvcResp, error) {
	var err error
	resp := &BridgXSvcResp{}
	bCli := bridgxcli.GetBridgXCli(ctx)
	cliReq := &bridgxcli.TaskDescribeReq{
		TaskId: taskId,
	}
	queryTaskC := make(chan bool, 1)
	// 轮询 task 进度
	wt := 5 * time.Second
	log.Logger.Infof("wait %v to continue", wt)
	time.Sleep(wt)
	var taskDescribe *bridgx.TaskDescribe
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
			httpResp, err := bCli.TaskDescribe(ctx, cliReq)
			if err != nil {
				errCnt++
				continue
			}
			taskDescribe = httpResp.Data
			if s.RatePass(taskDescribe.SuccessRate) { //TODO DELETE
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
	// 查询扩容后的 ip 列表
	instCnt := 0
	pageNum := 1
	cliReq2 := &bridgxcli.TaskInstancesReq{
		TaskId:         taskId,
		InstanceStatus: bridgx.Running,
		PageSize:       50,
	}
	for {
		cliReq2.PageNum = int64(pageNum)
		httpResp, err := bCli.TaskInstances(ctx, cliReq2)
		if err != nil {
			if err != nil {
				log.Logger.Errorf("task/instances 信息查询异常:%v | task_id:%d", err, taskId)
				return nil, err
			}
		}
		taskInstancesData := httpResp.Data
		if pageNum == 1 {
			InstanceList = make([]*types.InstanceInfo, 0, taskInstancesData.Pager.Total)
		}
		if pageNum == 1 && len(taskInstancesData.InstanceList) == 0 {
			err = errors.New("no instances found")
			log.Logger.Error(err)
			return nil, err
		}
		for _, item := range taskInstancesData.InstanceList {
			inst := &types.InstanceInfo{
				IpInner:    item.IpInner,
				IpOuter:    item.IpOuter,
				InstanceId: item.InstanceId,
			}
			InstanceList = append(InstanceList, inst)
		}
		instCnt += len(taskInstancesData.InstanceList)
		if instCnt >= taskInstancesData.Pager.Total {
			break
		}
		pageNum++
	}
	InstGroup.InstanceList = InstanceList
	resp.InstGroup = InstGroup

	return resp, nil
}

func (s *BridgXSvc) shrinkAction(ctx context.Context, taskId int64, clusterName string, instGroup *nodeact.InstanceGroup) (*BridgXSvcResp, error) {
	var err error
	resp := &BridgXSvcResp{}
	var ips []string
	for _, inst := range instGroup.InstanceList {
		ips = append(ips, inst.IpInner)
	}
	bCli := bridgxcli.GetBridgXCli(ctx)
	cliReq := &bridgxcli.ClusterShrinkReq{
		TaskName:    fmt.Sprintf("schedulx 缩容 %v 台", len(ips)),
		ClusterName: clusterName,
		Ips:         ips,
		Count:       int64(len(instGroup.InstanceList)),
	}
	httpResp, err := bCli.ClusterShrink(ctx, cliReq)
	if err != nil {
		return nil, err
	}
	resp.TaskId = httpResp.Data

	return resp, err
}

func (s *BridgXSvc) pollQueryShrinkAction(ctx context.Context, taskId int64, instances *nodeact.InstanceGroup) (*BridgXSvcResp, error) {
	var err error
	resp := &BridgXSvcResp{}
	bCli := bridgxcli.GetBridgXCli(ctx)
	cliReq := &bridgxcli.TaskDescribeReq{
		TaskId: taskId,
	}
	queryTaskC := make(chan bool, 1)
	// 轮询 task 进度
	time.Sleep(10 * time.Second)
	var taskDescribe *bridgx.TaskDescribe
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
			time.Sleep(5 * time.Second)
			httpResp, err := bCli.TaskDescribe(ctx, cliReq)
			if err != nil {
				errCnt++
				continue
			}
			taskDescribe = httpResp.Data
			if s.RatePass(taskDescribe.SuccessRate) {
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
	var ids []string
	for _, instance := range instances.InstanceList {
		ids = append(ids, instance.InstanceId)
	}
	if len(ids) == 0 {
		return resp, nil
	}
	_, err = repository.GetInstanceRepoIns().BatchUpdateStatusByIds(ctx, ids, types.InstanceStatusDeleted)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *BridgXSvc) getClusterAction(ctx context.Context, clusterName string) (*BridgXSvcResp, error) {
	var err error
	resp := &BridgXSvcResp{}
	bCli := bridgxcli.GetBridgXCli(ctx)
	cliReq := &bridgxcli.GetClusterByNameReq{
		ClusterName: clusterName,
	}
	httpResp, err := bCli.GetClusterByName(ctx, cliReq)
	if err != nil {
		return nil, err
	}
	clusterInfo := httpResp.Data
	resp.Auth = &types.InstanceAuth{
		UserName: clusterInfo.UserName,
		Pwd:      clusterInfo.Pwd,
	}
	return resp, nil
}

func (s *BridgXSvc) RatePass(rate string) bool {
	if cast.ToFloat64(rate) >= 1 {
		return true
	}
	return false
}
