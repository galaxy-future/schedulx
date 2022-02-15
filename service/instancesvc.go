package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/repository"
	"gorm.io/gorm"
)

type InstanceService struct {
}

var instanceServiceInstance *InstanceService
var instanceOnce sync.Once

func GetInstanceService() *InstanceService {
	instanceOnce.Do(func() {
		instanceServiceInstance = &InstanceService{}
	})
	return instanceServiceInstance
}

type InstanceCountResp struct {
	ServiceClusterList []ClusterInstanceCount `json:"service_cluster_list"`
}

type ClusterInstanceCount struct {
	ServiceClusterName string `json:"service_cluster_name"`
	ServiceClusterId   int64  `json:"service_cluster_id"`
	InstanceCount      int    `json:"instance_count"`
}

func (s *InstanceService) InstanceCountByCluster(ctx context.Context, serviceName, serviceClusterName string, clusterId int64) (*InstanceCountResp, error) {
	var clusterIds []int64
	if clusterId != 0 {
		clusterIds = append(clusterIds, clusterId)
	}
	names := make(map[int64]string, 0)
	if serviceName != "" {
		clusters, err := repository.GetServiceRepoInst().GetServiceClusters(ctx, serviceName, serviceClusterName)
		if err != nil {
			return nil, err
		}
		if len(clusters) != 0 {
			for _, cluster := range clusters {
				names[cluster.Id] = cluster.ClusterName
				clusterIds = append(clusterIds, cluster.Id)
			}
		}
	}
	if len(clusterIds) == 0 {
		return nil, fmt.Errorf("cluster ids empty")
	}
	clusters, err := repository.GetInstanceRepoIns().GetInstanceCountByClusterIds(ctx, clusterIds)
	if err != nil {
		return nil, err
	}
	instanceCount := make([]ClusterInstanceCount, 0, len(clusters))
	for _, cluster := range clusters {
		instanceCount = append(instanceCount, ClusterInstanceCount{
			ServiceClusterName: names[cluster.ServiceClusterId],
			ServiceClusterId:   cluster.ServiceClusterId,
			InstanceCount:      cluster.InstanceCount,
		})
	}

	return &InstanceCountResp{instanceCount}, nil
}

type GetServiceByIpResponse struct {
	ServiceName string `json:"service_name"`
	ClusterName string `json:"cluster_name"`
}

func (s *InstanceService) GetServiceByIp(ctx context.Context, ip string) (*GetServiceByIpResponse, error) {
	ins, err := repository.GetInstanceRepoIns().GetInstanceByIpInner(ctx, ip)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("there isn't any instance using this ip")
		}
		log.Logger.Errorf("GetInstanceByIpInner:%v error:%v", ip, err)
		return nil, err
	}
	if ins.Id == 0 {
		return nil, errors.New("there isn't any instance using this ip")
	}
	svcRepo := repository.GetServiceRepoInst()
	cluster, err := svcRepo.GetServiceCluster(ctx, ins.ServiceClusterId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("there isn't any cluster using this ip")
		}
		return nil, err
	}
	if cluster == nil {
		return nil, nil
	}
	return &GetServiceByIpResponse{
		ServiceName: cluster.ServiceName,
		ClusterName: cluster.ClusterName,
	}, nil
}
