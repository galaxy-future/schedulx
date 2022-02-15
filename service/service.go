package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/client/bridgxcli"
	"github.com/galaxy-future/schedulx/pkg/bridgx"
	"github.com/galaxy-future/schedulx/register/config/client"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/repository"
	"github.com/galaxy-future/schedulx/repository/model/db"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cast"
	"gorm.io/gorm/clause"
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
		if cluster.BridgxCluster == "" {
			//not binding bridgx cluster yet
			continue
		}
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

		info := bridgxCluster2ServiceCluster(resp.Data, count)
		info.ServiceCluster = cluster.ClusterName
		res = append(res, *info)
	}
	return &ClusterListResp{ClusterList: res}, nil
}

func (s *ServiceSvc) GetCloudClusterInfo(ctx context.Context, clusterId int64) (*ServiceClusterInfo, error) {
	resp, err := bridgxcli.GetBridgXCli(ctx).GetClusterById(ctx, &bridgxcli.GetClusterByIdReq{ClusterId: clusterId})
	if err != nil {
		return nil, err
	}
	serviceClusters, err := repository.GetInstanceRepoIns().GetInstanceCountByClusterIds(ctx, []int64{clusterId})
	if err != nil {
		return nil, err
	}
	var count int64
	if len(serviceClusters) > 0 {
		count = int64(serviceClusters[0].InstanceCount)
	}

	return bridgxCluster2ServiceCluster(resp.Data, count), nil
}

func (s *ServiceSvc) GetK8sClusterInfo(ctx context.Context, clusterId int64) (*ServiceClusterInfo, error) {
	return bridgxCluster2ServiceCluster(nil, 0), nil
}

func bridgxCluster2ServiceCluster(clusterInfo *bridgx.ClusterInfo, count int64) *ServiceClusterInfo {
	if clusterInfo == nil {
		return nil
	}
	var chargeType, cpuType string
	var core, memory int
	if clusterInfo.ChargeConfig != nil {
		chargeType = clusterInfo.ChargeConfig.ChargeType
	}
	if clusterInfo.ExtendConfig != nil {
		core = clusterInfo.ExtendConfig.InstanceCore
		memory = clusterInfo.ExtendConfig.InstanceMemory
		cpuType = clusterInfo.ExtendConfig.CpuType
	}
	info := ServiceClusterInfo{
		ServiceClusterId:   clusterInfo.Id,
		BridgxCluster:      clusterInfo.Name,
		InstanceCount:      count,
		InstanceTypeDesc:   genDesc(clusterInfo.InstanceType, core, memory),
		Provider:           clusterInfo.Provider,
		ComputingPowerType: cpuType,
		ChargeType:         chargeType,
	}
	return &info
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

func (s *ServiceSvc) Update(ctx context.Context, serviceName, description, domain, port, gitRepo string) (map[string]interface{}, error) {
	var err error
	ret, err := repository.GetServiceRepoInst().Update(ctx, serviceName, description, domain, port, gitRepo)
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
		Domain:      svcReq.ServiceInfo.Domain,
		Port:        svcReq.ServiceInfo.Port,
		GitRepo:     svcReq.ServiceInfo.GitRepo,
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

func (s *ServiceSvc) Delete(ctx context.Context, ids []int64) error {
	return repository.GetServiceRepoInst().DeleteServices(ctx, ids)
}

func (s *ServiceSvc) CreateRunningEnv(ctx context.Context, envInfo *types.RunningEnv) (err error) {
	dbo := client.WriteDBCli.WithContext(ctx)
	var service db.Service
	err = db.Get(envInfo.ServiceId, &service)
	if err != nil {
		log.Logger.Errorf("get service %v failed %v", envInfo.ServiceId, err)
		return err
	}

	env, computingRes := types2runningEnv(envInfo)
	env.Id = 0
	err = db.Create(&env, dbo)
	if err != nil {
		log.Logger.Errorf("create RunningEnv failed %v", err)
		return err
	}
	envId := env.GetId()
	if envId == 0 {
		err = errors.New("envId is 0")
		log.Logger.Errorf("create RunningEnv failed %v", err)
		return err
	}

	for _, resource := range computingRes {
		resource.Id = 0
		resource.EnvId = envId
		err = db.CreateIgnoreDuplicate(&resource, dbo)
		if err != nil {
			log.Logger.Errorf("create computing resource failed %v", err)
			continue
		}
	}
	return nil
}

func (s *ServiceSvc) DeleteRunningEnv(ctx context.Context, ids []int64) (err error) {
	dbo := client.WriteDBCli.WithContext(ctx)
	err = dbo.Delete(&db.RunningEnv{}, ids).Error
	if err != nil {
		log.Logger.Errorf("delete running env failed %v", err)
		return err
	}

	err = dbo.Where("env_id in ?", ids).Delete(&db.ComputingResource{}).Error
	if err != nil {
		log.Logger.Errorf("delete ComputingResource failed %v", err)
		return err
	}
	return nil
}

func (s *ServiceSvc) UpdateRunningEnv(ctx context.Context, envInfo *types.RunningEnv) (err error) {
	dbo := client.WriteDBCli.WithContext(ctx)
	var oldEnv db.RunningEnv
	err = db.Get(envInfo.Id, &oldEnv)
	if err != nil {
		log.Logger.Errorf("get running env %v failed %v", envInfo.Id, err)
		return err
	}

	env, computingRes := types2runningEnv(envInfo)
	env.ServiceId = oldEnv.ServiceId
	env.CreateAt = oldEnv.CreateAt
	err = db.Save(env, dbo)
	if err != nil {
		log.Logger.Errorf("save running env %v failed %v", envInfo.Id, err)
		return err
	}

	err = updateComputingRes(ctx, envInfo.Id, computingRes)
	if err != nil {
		log.Logger.Errorf("env %v, updateComputingRes failed %v", envInfo.Id, err)
		return err
	}
	return nil
}

func (s *ServiceSvc) GetRunningEnvById(ctx context.Context, envId int64) (*types.RunningEnv, error) {
	var env db.RunningEnv
	err := db.Get(envId, &env)
	if err != nil {
		log.Logger.Errorf("get running env %v failed %v", envId, err)
		return nil, err
	}

	computingRes := env.GetComputingRes()
	resourceInfo := make(map[int64]types.ResourceInfo, len(computingRes))
	for _, resource := range computingRes {
		switch resource.ComputingType {
		case types.CloudCluster:
			info, err := s.GetCloudClusterInfo(ctx, resource.ClusterId)
			if err != nil {
				log.Logger.Errorf("GetCloudClusterInfo %v failed, %v", resource.ClusterId, err)
				continue
			}
			resourceInfo[resource.Id] = types.ResourceInfo{
				ClusterId:          resource.ClusterId,
				BridgxCluster:      info.BridgxCluster,
				InstanceCount:      info.InstanceCount,
				InstanceTypeDesc:   info.InstanceTypeDesc,
				Provider:           info.Provider,
				ComputingPowerType: info.ComputingPowerType,
				ChargeType:         info.ChargeType,
			}
		case types.K8sCluster:
			_, err := s.GetK8sClusterInfo(ctx, resource.ClusterId)
			if err != nil {
				log.Logger.Errorf("GetK8sClusterInfo %v failed %v", resource.ClusterId, err)
				continue
			}
			resourceInfo[resource.Id] = types.ResourceInfo{
				ClusterId: resource.ClusterId,
			}
		}
	}

	return runningEnv2Types(&env, computingRes, resourceInfo), nil
}

func (s *ServiceSvc) GetRunningEnvByService(serviceId int64) ([]*types.RunningEnv, error) {
	var service db.Service
	err := db.Get(serviceId, &service)
	if err != nil {
		log.Logger.Errorf("get service %v failed %v", serviceId, err)
		return nil, err
	}

	envs := service.GetRunningEnv()
	envList := make([]*types.RunningEnv, 0, len(envs))
	for _, env := range envs {
		envList = append(envList, runningEnv2Types(&env, nil, nil))
	}
	return envList, nil
}

func getRunningEnvByRes(resId int64) (*types.RunningEnv, error) {
	var computingRes db.ComputingResource
	if err := client.ReadDBCli.Where("id=?", resId).First(&computingRes).Error; err != nil {
		log.Logger.Errorf("get computing resource %v failed, %v", resId, err)
		return nil, err
	}
	var env db.RunningEnv
	err := db.Get(computingRes.EnvId, &env)
	if err != nil {
		log.Logger.Errorf("get running env %v failed, %v", computingRes.EnvId, err)
		return nil, err
	}

	return runningEnv2Types(&env, []db.ComputingResource{computingRes}, nil), nil
}

func runningEnv2Types(env *db.RunningEnv, computingRes []db.ComputingResource, resourceInfo map[int64]types.ResourceInfo) *types.RunningEnv {
	healthCheck := types.HealthCheck{}
	err := jsoniter.UnmarshalFromString(env.HealthCheck, &healthCheck)
	if err != nil {
		log.Logger.Warnf("HealthCheck unmarshal failed, %v", err)
	}
	info := types.RunningEnv{
		ServiceId:   env.ServiceId,
		Id:          env.Id,
		Name:        env.Name,
		Type:        env.Type,
		Domain:      env.Domain,
		Port:        env.Port,
		HealthCheck: healthCheck,
	}
	for _, res := range computingRes {
		info.Resources = append(info.Resources, types.ComputingResource{
			Id:            res.Id,
			ComputingType: res.ComputingType,
			ResourceInfo:  resourceInfo[res.Id],
		})
	}
	return &info
}

func types2runningEnv(info *types.RunningEnv) (*db.RunningEnv, []db.ComputingResource) {
	healthCheck, err := jsoniter.MarshalToString(info.HealthCheck)
	if err != nil {
		log.Logger.Warnf("HealthCheck marshal failed, %v", err)
	}
	env := db.RunningEnv{
		Base: db.Base{
			Id: info.Id,
		},
		ServiceId:   info.ServiceId,
		Name:        info.Name,
		Type:        info.Type,
		Domain:      info.Domain,
		Port:        info.Port,
		HealthCheck: healthCheck,
	}

	resources := make([]db.ComputingResource, 0, len(info.Resources))
	for _, resource := range info.Resources {
		resources = append(resources, db.ComputingResource{
			Base: db.Base{
				Id: resource.Id,
			},
			EnvId:         info.Id,
			ComputingType: resource.ComputingType,
			ClusterId:     resource.ClusterId,
			ClusterName:   resource.BridgxCluster,
		})
	}

	return &env, resources
}

func updateComputingRes(ctx context.Context, envId int64, resources []db.ComputingResource) (err error) {
	oldIds := make([]string, 0)
	if err = client.ReadDBCli.WithContext(ctx).Model(&db.ComputingResource{}).Select("id").
		Where("env_id=?", envId).Scan(&oldIds).Error; err != nil {
		return err
	}
	newIds := make([]string, 0, len(resources))
	for _, v := range resources {
		newIds = append(newIds, cast.ToString(v.Id))
	}
	IdDiff := StringSliceDiff(oldIds, newIds)
	if len(IdDiff) > 0 {
		if err = client.WriteDBCli.WithContext(ctx).Where("id in ?", IdDiff).Delete(&db.ComputingResource{}).Error; err != nil {
			return err
		}
	}

	return client.WriteDBCli.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"computing_type", "cluster_id"}),
	}).Create(&resources).Error
}
