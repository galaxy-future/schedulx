package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/galaxy-future/schedulx/repository"
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
