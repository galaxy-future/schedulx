package service

import (
	"context"
	"errors"
	"sync"

	"github.com/galaxy-future/schedulx/register/config"

	"github.com/spf13/cast"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alb"
	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/pkg/cloud/aliyun"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/register/constant"
	"github.com/galaxy-future/schedulx/repository"
)

type MountService struct {
}

var (
	mountServiceInstance *MountService
	slbClient            aliyun.SLBClient
)
var mountOnce sync.Once

func GetMountSvcInst() *MountService {
	mountOnce.Do(func() {
		mountServiceInstance = &MountService{}
		slbClient, _ = aliyun.InitALB(config.GlobalConfig.AlibabaCloudAccount.Region,
			config.GlobalConfig.AlibabaCloudAccount.AccessKey,
			config.GlobalConfig.AlibabaCloudAccount.Secret) //
	})
	return mountServiceInstance
}

func (mou *MountService) Mount(ctx context.Context, svcReq *ExposeMountSvcReq) (*ExposeMountSvcResp, error) {
	var err error
	resp := &ExposeMountSvcResp{}

	vGroupId := svcReq.MountValue
	// 分页添加
	insListLen := len(svcReq.InstanceList)
	insListInfo := svcReq.InstanceList
	ipInner := []string{}
	serverList := []alb.AddServersToServerGroupServers{}
	if insListLen < constant.ALIYUNAddServerGroupLenMax {
		for _, item := range insListInfo {
			server := alb.AddServersToServerGroupServers{
				ServerId:    item.InstanceId,
				ServerIp:    item.IpInner,
				Port:        "80",
				ServerType:  "Ecs",
				Description: "slb挂载到后端服务，提供Http接口",
				Weight:      "100",
			}
			serverList = append(serverList, server)
			ipInner = append(ipInner, item.IpInner)
		}
		log.Logger.Infof("server conf:%+v\n", serverList)

		response, err := slbClient.CreateServers(ctx, vGroupId, &serverList)
		if err != nil {
			log.Logger.Errorf("errr:%v", err)
			return nil, err
		}

		// 更新数据库字段为 alb挂载完成更新数据库字段
		retCount, err := repository.GetInstanceRepoIns().BatchUpdateStatus(ctx, ipInner, svcReq.TaskId, types.InstanceStatusALB)
		// 返回对应的IP列表
		if err != nil {
			log.Logger.Errorf("BatchUpdates error:%v", err)
			return nil, err
		}
		log.Logger.Infof("records:%v resp:%v", retCount, response)
		resp.TaskId = svcReq.TaskId
		resp.InstanceList = svcReq.InstanceList // 临时认为全部操作成功 todo
		return resp, nil
	}
	/*
		// todo 超过40条分片处理
		tmpInsListInfo :=make([]interface{},len(insListInfo))
		for index,ins:=range insListInfo{
			tmpInsListInfo[index]=ins
		}
		chunk := util.ArraySplitChunk(tmpInsListInfo, constant.ALIYUNAddServerGroupLenMax)
		for _,info:=range chunk{
			if listInfo,ok := info.([]*types.InstanceInfo); ok{
				tmpServerList:=[]alb.AddServersToServerGroupServers{}
				for _,item:=range listInfo{
					server:=alb.AddServersToServerGroupServers{
						ServerId:    item.InstanceId, //"i-2zeh3bjq6rgpffhlkstu",
						ServerIp:    item.IpInner,    //"10.192.221.29",
						Port:        "80",
						ServerType:  "Ecs",
						Description: "slb挂载到后端服务，提供Http接口",
						Weight:      "100",
					}
					tmpServerList=append(serverList,server)
				}
			}
			// 调用服务
			response, err := slbClient.CreateServers(ctx, vGroupId, &serverList)
			fmt.Println("resp aa:", response)
		}
	*/
	err = errors.New("too many instances")
	return nil, err
}

func (mou *MountService) Umount(ctx context.Context, svcReq *ExposeUmountSvcReq) (*ExposeUmountSvcResp, error) {
	//用 taskid 获取要卸载的 groups
	var err error
	resp := &ExposeUmountSvcResp{}
	instRepo := repository.GetInstanceRepoIns()
	taskId, instanceList, err := instRepo.QueryInstsToUmount(ctx, svcReq.TaskId, types.InstanceStatusALB, cast.ToInt(svcReq.Count))
	if err != nil {
		return nil, err
	}
	vGroupId := svcReq.MountValue
	insListLen := len(instanceList)
	insListInfo := instanceList
	ipInner := []string{}
	serverList := []alb.RemoveServersFromServerGroupServers{}
	if insListLen < constant.ALIYUNAddServerGroupLenMax {
		for _, item := range insListInfo {
			server := alb.RemoveServersFromServerGroupServers{
				ServerId:   item.InstanceId, //"i-2zeh3bjq6rgpffhlkstu",
				ServerIp:   item.IpInner,    //"10.192.221.29",
				Port:       "80",
				ServerType: "Ecs",
			}
			serverList = append(serverList, server)
			ipInner = append(ipInner, item.IpInner)
		}
		log.Logger.Infof("server conf:%+v\n", serverList)
		response, err := slbClient.RemoveServer(ctx, vGroupId, &serverList)

		if err != nil {
			log.Logger.Error(err)
			return nil, err
		}
		// 更新数据库字段为 alb挂载完成更新数据库字段
		retCount, err := repository.GetInstanceRepoIns().BatchUpdateStatus(ctx, ipInner, taskId, types.InstanceStatusUNALB)
		// 返回对应的IP列表
		if err != nil {
			log.Logger.Errorf("BatchUpdates:%v", err)
			return nil, err
		}
		log.Logger.Infof("resp:%v records count:%d", response, retCount)
		resp.TaskId = taskId
		resp.InstanceList = instanceList // 临时认为全部操作成功 todo
		return resp, nil
	}
	err = errors.New("too many instances")
	return nil, err
}
