package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/client/bridgxcli"
	"github.com/galaxy-future/schedulx/register/config/client"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/repository"
	"github.com/galaxy-future/schedulx/repository/model/db"
)

type ServiceSvc struct {
}

var (
	serviceSvc     *ServiceSvc
	serviceSvcOnce sync.Once
)

type ServiceCreateSvcRequest struct {
	ServiceInfo *types.ServiceInfo `json:"service_info"`
}

type ServiceCreateSvcResponse struct {
	ServiceClusterId int64 `json:"service_cluster_id"`
}

func GetServiceIns() *ServiceSvc {
	serviceSvcOnce.Do(
		func() {
			serviceSvc = &ServiceSvc{}
		})
	return serviceSvc
}

type ServiceClusterInfo struct {
	ServiceClusterId   int64  `json:"service_cluster_id"`
	ServiceCluster     string `json:"service_cluster"`
	BridgxCluster      string `json:"bridgx_cluster"`
	InstanceCount      int64  `json:"instance_count"`
	InstanceTypeDesc   string `json:"instance_type_desc"`
	Provider           string `json:"provider"`
	ComputingPowerType string `json:"computing_power_type"`
	ChargeType         string `json:"charge_type"`
}

type ClusterListResp struct {
	ClusterList []ServiceClusterInfo `json:"cluster_list"`
}

func (s *ServiceSvc) GetServiceClusterList(ctx context.Context, serviceName string) (*ClusterListResp, error) {
	clusters, err := repository.GetServiceRepoInst().GetServiceClusters(ctx, serviceName, "")
	if err != nil {
		return nil, err
	}
	if len(clusters) == 0 {
		return &ClusterListResp{}, nil
	}
	res := make([]ServiceClusterInfo, 0)
	for _, cluster := range clusters {
		resp, err := bridgxcli.GetBridgXCli(ctx).GetClusterByName(ctx, &bridgxcli.GetClusterByNameReq{ClusterName: cluster.BridgxCluster})
		if err != nil {
			return nil, err
		}
		serviceClusters, err := repository.GetInstanceRepoIns().GetInstanceCountByClusterIds(ctx, []int64{cluster.Id})
		if err != nil {
			return nil, err
		}
		var count int64
		if len(serviceClusters) > 0 {
			count = int64(serviceClusters[0].InstanceCount)
		}
		clusterInfo := resp.Data
		var chargeType string
		if clusterInfo.ChargeConfig != nil {
			chargeType = clusterInfo.ChargeConfig.ChargeType
		}
		res = append(res, ServiceClusterInfo{
			ServiceClusterId:   cluster.Id,
			ServiceCluster:     cluster.ClusterName,
			BridgxCluster:      cluster.BridgxCluster,
			InstanceCount:      count,
			InstanceTypeDesc:   genDesc(clusterInfo.InstanceType, clusterInfo.InstanceCore, clusterInfo.InstanceMemory),
			ComputingPowerType: getComputingPowerType(clusterInfo.Provider, clusterInfo.InstanceType),
			Provider:           clusterInfo.Provider,
			ChargeType:         chargeType,
		})
	}
	return &ClusterListResp{ClusterList: res}, nil
}

func getComputingPowerType(provider, instanceType string) string {
	switch provider {
	case "AlibabaCloud":
		if strings.Contains(instanceType, "gn") {
			return "GPU"
		}
	case "HuaweiCloud":
		if strings.HasPrefix(instanceType, "G") || strings.HasPrefix(instanceType, "P") {
			return "GPU"
		}
	}
	return "CPU"
}

func genDesc(instanceType string, core, memory int) string {
	return fmt.Sprintf("%d核%dG(%v)", core, memory, instanceType)
}

func (s *ServiceSvc) GetServiceList(ctx context.Context, page, pageSize int, serviceName, lang string) (map[string]interface{}, error) {
	var err error
	list, total, err := repository.GetServiceRepoInst().GetServiceList(ctx, page, pageSize, serviceName, lang)
	if err != nil {
		log.Logger.Errorf("repository.GetServiceRepoInst().GetServiceList error:", err)
	}
	ret := map[string]interface{}{
		"service_list": list,
		"pager": struct {
			PageNumber int   `json:"page_number"`
			PageSize   int   `json:"page_size"`
			Total      int64 `json:"total"`
		}{
			PageSize:   pageSize,
			PageNumber: page,
			Total:      total,
		},
	}
	return ret, nil
}

func (s *ServiceSvc) GetExpandHistory(ctx context.Context, page, pageSize, serviceClusterId int) (map[string]interface{}, error) {
	var err error
	tempList, total, err := repository.GetScheduleTemplateRepoInst().GetScheduleTempList(ctx, page, pageSize, serviceClusterId)
	if err != nil {
		log.Logger.Errorf("serviceClusterId:%v page:%v pageSize:%v error:%v", serviceClusterId, page, pageSize, err)
		return nil, err
	}
	ret := map[string]interface{}{
		"schedule_task_list": tempList,
		"pager": struct {
			PageNumber int   `json:"page_number"`
			PageSize   int   `json:"page_size"`
			Total      int64 `json:"total"`
		}{
			PageSize:   pageSize,
			PageNumber: page,
			Total:      total,
		},
	}
	return ret, nil
}

func (s *ServiceSvc) Detail(ctx context.Context, serviceName string) (map[string]interface{}, error) {
	var err error
	serviceDetail, err := repository.GetServiceRepoInst().GetServiceDetail(ctx, serviceName)
	if err != nil {
		return nil, err
	}
	ret := map[string]interface{}{
		"service_info": serviceDetail,
	}
	return ret, nil
}

func (s *ServiceSvc) UpdateDesc(ctx context.Context, serviceName, description string) (map[string]interface{}, error) {
	var err error
	ret, err := repository.GetServiceRepoInst().UpdateDesc(ctx, serviceName, description)
	if err != nil {
		log.Logger.Errorf("update service_name:%v error:%v", serviceName, err)
		return nil, err
	}
	return map[string]interface{}{
		"records": ret,
	}, nil
}

func (s *ServiceSvc) CreateService(ctx context.Context, svcReq *ServiceCreateSvcRequest) (*ServiceCreateSvcResponse, error) {
	var err error
	svcResp := &ServiceCreateSvcResponse{}
	svcObj := &db.Service{
		ServiceName: svcReq.ServiceInfo.ServiceName,
		Description: svcReq.ServiceInfo.Description,
		Language:    svcReq.ServiceInfo.Language,
	}
	dbo := client.WriteDBCli.Begin().WithContext(ctx)
	defer func() {
		if err != nil {
			dbo.Rollback()
			return
		}
		dbo.Commit()
	}()
	err = db.Create(svcObj, dbo)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	// 创建一个服务默认集群
	svcClusterObj := &db.ServiceCluster{
		ServiceName: svcReq.ServiceInfo.ServiceName,
		ClusterName: "default",
		//CreateAt: time.Now(),
	}
	err = db.Create(svcClusterObj, dbo)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	svcResp.ServiceClusterId = svcClusterObj.Id
	return svcResp, nil
}
