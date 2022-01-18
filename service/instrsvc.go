package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/pkg/nodeact"
	"github.com/galaxy-future/schedulx/pkg/tool"
	"github.com/galaxy-future/schedulx/register/config"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/repository"
	"github.com/galaxy-future/schedulx/repository/model/db"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

type InstrSvc struct {
	BridgXExpand    types.Action
	BridgXShrink    types.Action
	NodeActInitBase types.Action
	NodeActInitSvc  types.Action
	MountSLB        types.Action
	UmountSLB       types.Action
	MountNginx      types.Action
	UmountNginx     types.Action
}

var instrSvcSvc *InstrSvc
var instrSvcOnce sync.Once

func GetInstrSvcInst() *InstrSvc {
	instrSvcOnce.Do(func() {
		instrSvcSvc = &InstrSvc{
			"bridgx.expand",
			"bridgx.shrink",
			"nodeact.initbase",
			"nodeact.initsvc",
			"mount.slb",
			"umount.slb",
			"mount.nginx",
			"umount.nginx",
		}
	})

	return instrSvcSvc
}

type InstrSvcReq struct {
	ServiceName      string
	ServiceClusterId int64
	ScheduleTaskId   int64
	InstrId          int64
	Instruction      *db.Instruction
	BridgXSvcReq     *BridgXSvcReq
	NodeActSvcReq    *NodeActSvcReq
}

type InstrSvcResp struct {
	BridgXSvcResp  *BridgXSvcResp
	NodeActSvcResp *NodeActSvcResp
}

func (s *InstrSvc) entryLog(ctx context.Context, act string, req interface{}) {
	log.Logger.Infof("entry log | act[%s] | req:%s", act, tool.ToJson(req))
}

func (s *InstrSvc) exitLog(ctx context.Context, act string, req, resp interface{}, err error) {
	log.Logger.Infof("exit log | act[%s] | req:%s | resp:%s | err:%v", act, tool.ToJson(req), tool.ToJson(resp), err)
}

func (s *InstrSvc) ExecAct(ctx context.Context, args interface{}, act types.Action) (svcResp interface{}, err error) {
	svcReq, ok := args.(*InstrSvcReq)
	if !ok {
		return nil, errors.New("init service request err")
	}
	s.entryLog(ctx, string(act), svcReq)
	defer func() {
		s.exitLog(ctx, string(act), svcReq, svcResp, err)
		instrStatus := types.InstrRecStatusSucc
		msg := ""
		if err != nil {
			instrStatus = types.InstrRecStatusFail
			msg = err.Error()
		}
		_ = s.updateInstrRecord(ctx, svcReq.ScheduleTaskId, svcReq.InstrId, instrStatus, msg)
	}()
	err = s.createInstrRecord(ctx, svcReq.ScheduleTaskId, svcReq.InstrId)
	if err != nil {
		return nil, err
	}

	switch act {
	case s.BridgXExpand:
		svcResp, err = s.bridgXExpandAction(ctx, svcReq.ScheduleTaskId, svcReq.BridgXSvcReq.Count, svcReq.BridgXSvcReq.ClusterName)
	case s.BridgXShrink:
		svcResp, err = s.bridgXShrinkAction(ctx, svcReq.ScheduleTaskId, svcReq.BridgXSvcReq, svcReq.ServiceClusterId)
	case s.NodeActInitBase:
		wt := 25 * time.Second
		log.Logger.Infof("准备机器环境初始化...等待%v", wt)
		time.Sleep(wt)
		svcResp, err = s.nodeActInitBaseAction(ctx, svcReq.ScheduleTaskId, svcReq.NodeActSvcReq.InstGroup, svcReq.NodeActSvcReq.Auth, svcReq.ServiceClusterId)
	case s.NodeActInitSvc:
		svcResp, err = s.nodeActInitSvcAction(ctx, svcReq.ScheduleTaskId, svcReq.NodeActSvcReq.InstGroup, svcReq.NodeActSvcReq.Auth, svcReq.Instruction)
	case s.MountSLB:
		svcResp, err = s.nodeActMountInstAction(ctx, svcReq.ScheduleTaskId, svcReq.NodeActSvcReq.InstGroup, svcReq.Instruction)
	case s.UmountSLB:
		svcResp, err = s.nodeActUmountInstAction(ctx, svcReq.ScheduleTaskId, svcReq.NodeActSvcReq.TaskId, svcReq.NodeActSvcReq.UmountSlbSvcReq, svcReq.Instruction, svcReq.ServiceClusterId)
	default:
		err = errors.New("no act matched")
	}
	return svcResp, err
}

func (s *InstrSvc) bridgXExpandAction(ctx context.Context, schedTaskId, count int64, clusterName string) (*InstrSvcResp, error) {
	var err error
	taskRepo := repository.GetTaskRepoInst()
	_ = taskRepo.UpdateTaskStep(ctx, schedTaskId, types.TaskStepBridgxExpandInit, "")
	resp := &InstrSvcResp{}
	defer func() {
		if err == nil {
			_ = taskRepo.UpdateTaskStep(ctx, schedTaskId, types.TaskStepBridgxExpandSucc, "")
		}
	}()
	bridgXSvc = GetBridgXSvcInst()
	bridgXSvcReq := &BridgXSvcReq{
		Count:       count,
		ClusterName: clusterName,
	}
	var svcResp interface{}
	bridgXSvcResp := &BridgXSvcResp{}
	svcResp, err = bridgXSvc.ExecAct(ctx, bridgXSvcReq, bridgXSvc.Expand)
	if err != nil {
		return nil, err
	}
	bridgXSvcResp.TaskId = svcResp.(*BridgXSvcResp).TaskId
	if bridgXSvcResp.TaskId == 0 {
		err = errors.New("bridgx expand resp.taskid is 0")
		log.Logger.Error(err)
		return nil, err
	}
	bridgXSvcReq = &BridgXSvcReq{
		TaskId: bridgXSvcResp.TaskId,
	}
	svcResp, err = bridgXSvc.ExecAct(ctx, bridgXSvcReq, bridgXSvc.PoolQueryExpand) // 循环等待和查询,有超时标准和
	if err != nil {
		return nil, err
	}
	bridgXSvcResp.InstGroup = svcResp.(*BridgXSvcResp).InstGroup
	log.Logger.Infof("[PoolQueryExpand] bridgXSvcResp:%+v", tool.ToJson(bridgXSvcResp))
	if bridgXSvcResp.InstGroup == nil || len(bridgXSvcResp.InstGroup.InstanceList) == 0 {
		err = errors.New("no instances found!")
		log.Logger.Error(err)
		return nil, err
	}
	resp.BridgXSvcResp = bridgXSvcResp

	bridgXSvcReq = &BridgXSvcReq{
		ClusterName: clusterName,
	}
	// 获取登录用户名和密码 (ssh 用)
	svcResp, err = bridgXSvc.ExecAct(ctx, bridgXSvcReq, bridgXSvc.GetCluster)
	bridgXSvcResp = svcResp.(*BridgXSvcResp)
	resp.BridgXSvcResp.Auth = bridgXSvcResp.Auth
	return resp, err
}

func (s *InstrSvc) bridgXShrinkAction(ctx context.Context, schedTaskId int64, bridgXSvcReq *BridgXSvcReq, serviceClusterId int64) (*InstrSvcResp, error) {
	var err error
	taskRepo := repository.GetTaskRepoInst()
	_ = taskRepo.UpdateTaskStep(ctx, schedTaskId, types.TaskStepBridgxShrinkInit, "")
	defer func() {
		if err == nil {
			_ = taskRepo.UpdateTaskStep(ctx, schedTaskId, types.TaskStepBridgxShrinkSucc, "")
		}
	}()
	if bridgXSvcReq.InstGroup == nil || len(bridgXSvcReq.InstGroup.InstanceList) == 0 {
		// 主动获取 instances
		log.Logger.Infof("fetch InstanceList")
		instRepo := repository.GetInstanceRepoIns()
		insts, err := instRepo.NotDeletedInstsWithLimit(ctx, serviceClusterId, int(bridgXSvcReq.Count))
		//insts, err := instRepo.InstsQueryByPage(ctx, bridgXSvcReq.TaskId, "", nil)
		if err != nil {
			log.Logger.Error(err)
			return nil, err
		}
		instanceList := make([]*types.InstanceInfo, 0, len(insts))
		for _, inst := range insts {
			item := &types.InstanceInfo{
				IpInner:    inst.IpInner,
				IpOuter:    inst.IpOuter,
				InstanceId: inst.InstanceId,
			}
			instanceList = append(instanceList, item)
		}
		bridgXSvcReq.InstGroup = &nodeact.InstanceGroup{
			TaskId:       bridgXSvcReq.TaskId,
			InstanceList: instanceList,
		}
		bridgXSvcReq.Count = int64(len(instanceList))
	}
	resp := &InstrSvcResp{}
	bridgXSvc = GetBridgXSvcInst()
	svcResp, err := bridgXSvc.ExecAct(ctx, bridgXSvcReq, bridgXSvc.Shrink)
	if err != nil {
		return nil, err
	}
	var bridgXSvcResp *BridgXSvcResp
	bridgXSvcResp = svcResp.(*BridgXSvcResp)
	taskId := bridgXSvcResp.TaskId
	bridgXSvcReq.TaskId = taskId

	svcResp, err = bridgXSvc.ExecAct(ctx, bridgXSvcReq, bridgXSvc.PoolQueryShrink) // 循环等待和查询,有超时标准和
	if err != nil {
		return nil, err
	}
	bridgXSvcResp = svcResp.(*BridgXSvcResp)
	log.Logger.Debugf("[bridgXShrinkAction] bridgXSvcResp:%+v", tool.ToJson(bridgXSvcResp))
	resp.BridgXSvcResp = bridgXSvcResp
	return resp, nil
}

func (s *InstrSvc) nodeActInitBaseAction(ctx context.Context, schedTaskId int64, instGroup *nodeact.InstanceGroup, auth *types.InstanceAuth, serviceClusterId int64) (*InstrSvcResp, error) {
	var err error
	taskRepo := repository.GetTaskRepoInst()
	_ = taskRepo.UpdateTaskStep(ctx, schedTaskId, types.TaskStepBaseEnvInit, "")
	resp := &InstrSvcResp{}
	defer func() {
		if err == nil {
			_ = taskRepo.UpdateTaskStep(ctx, schedTaskId, types.TaskStepBaseEnvSucc, "")
		}
	}()
	nodeXSvcReq := &NodeActSvcReq{
		InstGroup:        instGroup,
		Auth:             auth,
		ServiceClusterId: serviceClusterId,
	}
	nodeActSvc = GetNodeActSvcInst()
	// 执行初始化
	_, err = nodeActSvc.ExecAct(ctx, nodeXSvcReq, nodeActSvc.InitBase)
	if err != nil {
		return nil, err
	}
	nodeXSvcReq.TaskId = instGroup.TaskId
	// 轮询执行结果
	svcResp, err := nodeActSvc.ExecAct(ctx, nodeXSvcReq, nodeActSvc.PollQueryInitBase)
	if err != nil {
		return nil, err
	}
	nodeActSvcResp := svcResp.(*NodeActSvcResp)
	resp.NodeActSvcResp = nodeActSvcResp
	return resp, err
}

func (s *InstrSvc) nodeActInitSvcAction(ctx context.Context, schedTaskId int64, instGroup *nodeact.InstanceGroup, auth *types.InstanceAuth, instruction *db.Instruction) (*InstrSvcResp, error) {
	var err error
	taskRepo := repository.GetTaskRepoInst()
	_ = taskRepo.UpdateTaskStep(ctx, schedTaskId, types.TaskStepSvcEnvInit, "")
	defer func() {
		if err == nil {
			_ = taskRepo.UpdateTaskStep(ctx, schedTaskId, types.TaskStepSvcEnvSucc, "")
		}
	}()
	params := &types.ParamsServiceEnv{}
	err = jsoniter.Unmarshal([]byte(instruction.Params), params)
	if err != nil {
		log.Logger.Errorf("instrParams:%s:%v", instruction.Params, err.Error())
		return nil, err
	}
	resp := &InstrSvcResp{}
	nodeXSvcReq := &NodeActSvcReq{
		InstGroup: instGroup,
		Auth:      auth,
		InitServicSvcReq: &InitServicSvcReq{
			Cmd:    instruction.Cmd,
			Params: params,
		},
	}
	_, err = nodeActSvc.ExecAct(ctx, nodeXSvcReq, nodeActSvc.InitService)
	if err != nil {
		return nil, err
	}
	// 3.4 轮询 initService 是否完成
	nodeXSvcReq.TaskId = instGroup.TaskId
	svcResp, err := nodeActSvc.ExecAct(ctx, nodeXSvcReq, nodeActSvc.PollQueryInitService)
	if err != nil {
		return nil, err
	}
	nodeActSvcResp := svcResp.(*NodeActSvcResp)
	resp.NodeActSvcResp = nodeActSvcResp
	return resp, nil
}

func (s *InstrSvc) nodeActMountInstAction(ctx context.Context, schedTaskId int64, instGroup *nodeact.InstanceGroup, instr *db.Instruction) (*InstrSvcResp, error) {
	var err error
	taskRepo := repository.GetTaskRepoInst()
	_ = taskRepo.UpdateTaskStep(ctx, schedTaskId, types.TaskStepMountInit, "")
	defer func() {
		if err == nil {
			_ = taskRepo.UpdateTaskStep(ctx, schedTaskId, types.TaskStepMountSucc, "")
		}
	}()
	params := instr.Params
	slbMountInfo := &nodeact.ParamsMountInfo{}
	err = jsoniter.Unmarshal([]byte(params), slbMountInfo)
	if err != nil {
		log.Logger.Errorf("params:%s:%v", params, err.Error())
		return nil, err
	}
	resp := &InstrSvcResp{}
	nodeXSvcReq := &NodeActSvcReq{
		InstGroup:    instGroup,
		SlbMountInfo: slbMountInfo,
	}
	svcResp, err := nodeActSvc.ExecAct(ctx, nodeXSvcReq, nodeActSvc.MountSlb)
	if err != nil {
		return nil, err
	}

	nodeActSvcResp := svcResp.(*ExposeMountSvcResp)
	resp.NodeActSvcResp = &NodeActSvcResp{}
	resp.NodeActSvcResp.InstGroup = &nodeact.InstanceGroup{}
	resp.NodeActSvcResp.InstGroup.InstanceList = make([]*types.InstanceInfo, 0)
	resp.NodeActSvcResp.InstGroup.InstanceList = nodeActSvcResp.InstanceList
	return resp, nil
}

func (s *InstrSvc) nodeActUmountInstAction(ctx context.Context, schedTaskId, nodeActTaskId int64, umountSlbSvcReq *UmountSlbSvcReq, instr *db.Instruction, serviceClusterId int64) (*InstrSvcResp, error) {
	var err error
	taskRepo := repository.GetTaskRepoInst()
	_ = taskRepo.UpdateTaskStep(ctx, schedTaskId, types.TaskStepUmountInit, "")
	defer func() {
		if err == nil {
			_ = taskRepo.UpdateTaskStep(ctx, schedTaskId, types.TaskStepUmountSucc, "")
		}
	}()
	params := instr.Params
	slbMountInfo := &nodeact.ParamsMountInfo{}
	err = jsoniter.Unmarshal([]byte(params), slbMountInfo)
	if err != nil {
		log.Logger.Errorf("params:%s:%v", params, err.Error())
		return nil, err
	}
	umountSlbSvcReq.SlbInfo = slbMountInfo
	resp := &InstrSvcResp{
		NodeActSvcResp: &NodeActSvcResp{
			InstGroup: &nodeact.InstanceGroup{},
		},
	}
	nodeXSvcReq := &NodeActSvcReq{
		ServiceClusterId: serviceClusterId,
		TaskId:           nodeActTaskId,   // 原扩容任务 id
		UmountSlbSvcReq:  umountSlbSvcReq, //  卸载所需的信息
	}
	svc := GetNodeActSvcInst()
	svcResp, err := svc.ExecAct(ctx, nodeXSvcReq, svc.UmountSlb)
	if err != nil {
		return nil, err
	}
	nodeActSvcResp := svcResp.(*ExposeUmountSvcResp)
	resp.NodeActSvcResp.InstGroup.InstanceList = nodeActSvcResp.InstanceList
	return resp, nil

}

func (s *InstrSvc) createInstrRecord(ctx context.Context, taskId, instrId int64) error {
	var err error
	obj := &db.InstrRecord{
		TaskId:      taskId,
		InstrStatus: types.InstrRecStatusRunning,
		InstrId:     instrId,
	}
	err = db.Create(obj, nil)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	return nil
}

func (s *InstrSvc) updateInstrRecord(ctx context.Context, taskId, instrId int64, instrStatus, msg string) error {
	var err error
	where := map[string]interface{}{
		"task_id":  taskId,
		"instr_id": instrId,
	}
	data := map[string]interface{}{
		"instr_status": instrStatus,
		"msg":          tool.SubStr(msg, 100),
	}
	rowsAffected, err := db.Updates(&db.InstrRecord{}, where, data, nil)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	if rowsAffected != 1 {
		err = config.ErrRowsAffectedInvalid
		log.Logger.Error(err)
		return err
	}
	return nil
}

func (s *InstrSvc) CreateBridgxExpandInstr(ctx context.Context, tmplId int64, revTmplId int64, needReverse bool, dbo *gorm.DB) (int64, int64, error) {
	//创建 instruction
	var err error
	obj := &db.Instruction{
		Cmd:         "",
		Params:      "",
		InstrAction: s.BridgXExpand,
		TmplId:      tmplId,
	}
	err = db.Create(obj, dbo)
	if err != nil {
		log.Logger.Error(err)
		return 0, 0, err
	}
	var reverseInstrId int64
	if needReverse {
		reverseInstrId, err = s.CreateBridgxShrinkInstr(ctx, revTmplId, dbo)
		if err != nil {
			log.Logger.Error(err)
			return 0, 0, err
		}
		return obj.Id, reverseInstrId, err
	}
	return obj.Id, 0, nil
}

func (s *InstrSvc) CreateBridgxShrinkInstr(ctx context.Context, tmplId int64, dbo *gorm.DB) (int64, error) {
	//创建 instruction
	var err error
	obj := &db.Instruction{
		Cmd:         "",
		Params:      "",
		InstrAction: s.BridgXShrink,
		TmplId:      tmplId,
	}
	err = db.Create(obj, dbo)
	if err != nil {
		log.Logger.Error(err)
		return 0, err
	}
	return obj.Id, nil
}

func (s *InstrSvc) CreateBaseEnvInstr(ctx context.Context, args *types.BaseEnv, tmpId int64, needReverse bool, dbo *gorm.DB) (int64, int64, error) {
	//创建 instruction
	var err error
	params, _ := jsoniter.MarshalToString(&types.BaseEnv{
		IsContainer: args.IsContainer,
	})
	obj := &db.Instruction{
		Cmd:         "",
		Params:      params,
		InstrAction: s.NodeActInitBase,
		TmplId:      tmpId,
	}
	err = db.Create(obj, dbo)
	if err != nil {
		log.Logger.Error(err)
		return 0, 0, err
	}
	return obj.Id, 0, nil
}

func (s *InstrSvc) CreateServiceEnvInstr(ctx context.Context, args *types.ServiceEnv, tmplId int64, needReverse bool, dbo *gorm.DB) (int64, int64, error) {
	pass, err := tool.AesEncrypt([]byte(args.Password), []byte(args.Account))
	if err != nil {
		return 0, 0, err
	}

	//创建 instruction
	params, _ := jsoniter.MarshalToString(&types.ServiceEnv{
		ImageStorageType: args.ImageStorageType,
		ImageUrl:         args.ImageUrl,
		Port:             args.Port,
		ServiceName:      args.ServiceName,
		Account:          args.Account,
		Password:         pass,
	})
	obj := &db.Instruction{
		Cmd:         args.Cmd,
		Params:      params,
		InstrAction: s.NodeActInitSvc,
		TmplId:      tmplId,
	}
	err = db.Create(obj, dbo)
	if err != nil {
		log.Logger.Error(err)
		return 0, 0, err
	}
	return obj.Id, 0, nil
}

func (s *InstrSvc) CreateMountSlbInstr(ctx context.Context, args *types.ParamsMount, tmplId, revTmplId int64, needReverse bool, dbo *gorm.DB) (int64, int64, error) {
	//创建 instruction
	var err error
	if ok := IsAlibabaCloudAccountValid(config.GlobalConfig.AlibabaCloudAccount); !ok {
		err = errors.New("invalid AlibabaCloudAccount Config")
		log.Logger.Error(err)
		return 0, 0, err
	}
	params, _ := jsoniter.MarshalToString(&nodeact.ParamsMountInfo{
		MountType:  args.MountType,
		MountValue: args.MountValue,
	})
	obj := &db.Instruction{
		Cmd:         "",
		Params:      params,
		InstrAction: s.MountSLB,
		TmplId:      tmplId,
	}
	err = db.Create(obj, dbo)
	if err != nil {
		log.Logger.Error(err)
		return 0, 0, err
	}
	var reverseInstrId int64
	if needReverse {
		reverseInstrId, err = s.CreateUMountSlbInstr(ctx, args, revTmplId, dbo)
		if err != nil {
			log.Logger.Error(err)
			return 0, 0, err
		}
		return obj.Id, reverseInstrId, nil
	}
	return obj.Id, 0, nil
}

func (s *InstrSvc) CreateUMountSlbInstr(ctx context.Context, args *types.ParamsMount, tmplId int64, dbo *gorm.DB) (int64, error) {
	//创建 instruction
	var err error
	params, _ := jsoniter.MarshalToString(&nodeact.ParamsMountInfo{
		MountType:  args.MountType,
		MountValue: args.MountValue,
	})
	obj := &db.Instruction{
		Cmd:         "",
		Params:      params,
		InstrAction: s.UmountSLB,
		TmplId:      tmplId,
	}
	err = db.Create(obj, dbo)
	if err != nil {
		log.Logger.Error(err)
		return 0, err
	}
	return obj.Id, nil
}
