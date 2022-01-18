package repository

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/galaxy-future/schedulx/api/types"
	"gorm.io/gorm"

	"github.com/galaxy-future/schedulx/pkg/nodeact"
	jsoniter "github.com/json-iterator/go"

	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/repository/model/db"
)

type ServiceRepo struct {
}

// ServiceListLogic 服务列表数据item
type ServiceListLogic struct {
	ServiceId          int64  `json:"service_id"`
	ServiceName        string `json:"service_name"`
	ClusterNum         int    `json:"cluster_num"`
	Language           string `json:"language"`
	ImageUrl           string `json:"image_url"`
	ServiceClusterId   int64  `json:"service_cluster_id"`
	ServiceClusterName string `json:"service_cluster_name"`
	TmplExpandId       int64  `json:"tmpl_expand_id"`
	TmplExpandName     string `json:"tmpl_expand_name"`
	Description        string `json:"description"`
	AutoDecision       string `json:"auto_decision"`
	TaskTypeStatus     string `json:"task_type_status"`
}

// ServiceDetailLogic 服务详情数据
type ServiceDetailLogic struct {
	ServiceName        string `json:"service_name"`
	ServiceClusterId   int64  `json:"service_cluster_id"`
	ServiceClusterName string `json:"service_cluster_name"`
	Description        string `json:"description"`
	TmplExpandName     string `json:"tmpl_expand_name"`
	TmplExpandId       int64  `json:"tmpl_expand_id"`
}

var serviceRepoInst *ServiceRepo
var serviceRepoOnce sync.Once

func GetServiceRepoInst() *ServiceRepo {
	serviceRepoOnce.Do(func() {
		serviceRepoInst = &ServiceRepo{}
	})
	return serviceRepoInst
}

func (r *ServiceRepo) GetService(ctx context.Context, serviceName string) (*db.Service, error) {
	var err error
	obj := &db.Service{}
	where := map[string]interface{}{
		"service_name": serviceName,
	}
	err = db.QueryFirst(where, obj)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	return obj, nil
}

func (r *ServiceRepo) GetServiceCluster(ctx context.Context, id int64) (*db.ServiceCluster, error) {
	var err error
	obj := &db.ServiceCluster{}
	err = db.Get(id, obj)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	return obj, nil
}

func (r *ServiceRepo) GetServiceClusters(ctx context.Context, serviceName, serviceClusterName string) ([]*db.ServiceCluster, error) {
	var err error
	clusters := make([]*db.ServiceCluster, 0)
	where := map[string]interface{}{
		"service_name": serviceName,
	}
	if serviceClusterName != "" {
		where["cluster_name"] = serviceClusterName
	}
	err = db.QueryAll(where, &clusters, "", nil)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	return clusters, nil
}

// GetServiceList 获取分页数据
func (r *ServiceRepo) GetServiceList(ctx context.Context, page, pageSize int, serviceName, lang string) ([]ServiceListLogic, int64, error) {
	var err error
	list := []ServiceListLogic{}
	model := []db.Service{}
	where := make(map[string]interface{})

	if serviceName != "" {
		where["service_name"] = serviceName
	}
	if lang != "" {
		where["language"] = lang
	}
	// 1.查询service 表
	count, err := db.Query(where, page, pageSize, &model, "id desc", []string{"id", "service_name", "description", "language"}, true)
	if err != nil {
		log.Logger.Errorf("db.Query error:%v", err)
		return nil, 0, err
	}
	if count == 0 {
		err = gorm.ErrRecordNotFound
		log.Logger.Error(err)
		return nil, 0, err
	}
	// 2.查询service_cluster表
	whereCluster := map[string]interface{}{
		"service_name": serviceName,
	}
	list = make([]ServiceListLogic, len(model))
	// 组装数据返回
	for index, item := range model {
		if serviceName == "" {
			whereCluster["service_name"] = item.ServiceName
		}
		modelCluster := db.ServiceCluster{}
		err = db.QueryFirst(whereCluster, &modelCluster)
		if err != nil {
			log.Logger.Errorf("db.queryAll table [service_cluster] error:%v", err)
			return nil, 0, err
		}
		whereTemp := map[string]interface{}{
			"service_name":       serviceName,
			"service_cluster_id": modelCluster.Id,
		}
		if serviceName == "" {
			whereTemp["service_name"] = item.ServiceName
		}
		// 3.查询 schedule_template表
		modelTempLate := db.ScheduleTemplate{}
		err = db.QueryFirst(whereTemp, &modelTempLate)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Logger.Errorf("db.QuereyAll table modelTemplate error:%v", err)
			continue
			//return nil, 0, err
		}
		instrIds := strings.Split(strings.Trim(modelTempLate.InstrGroup, "[]"), ",")
		// 查询指令
		whereInstr := map[string]interface{}{
			"id": instrIds,
		}
		instrModel := []db.Instruction{}
		// 4.查询指令步骤表
		err = db.QueryAll(whereInstr, &instrModel, "id asc", []string{"params", "instr_action"})
		if err != nil {
			log.Logger.Errorf("db.Query table [instruction] error:%v", err)
			continue
			//return nil, 0, err
		}
		imageUrl := ""
		for _, insItem := range instrModel {
			if insItem.InstrAction == types.SERVICEInt {
				params := &nodeact.ParamsServiceEnv{}
				err = jsoniter.Unmarshal([]byte(insItem.Params), params)
				if err != nil {
					log.Logger.Errorf("params exception %v:error:%v", insItem.Params, err)
					return nil, 0, err
				}
				imageUrl = params.ImageUrl
			}
			log.Logger.Infof("instr info id:%d params:%v,action:%v", insItem.Id, insItem.Params, insItem.InstrAction)
		}

		whereTask := map[string]interface{}{
			"sched_tmpl_id": modelTempLate.Id,
		}
		// 5.查询 任务表
		modelTask := &db.Task{}
		err = db.QueryLast(whereTask, modelTask)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Logger.Errorf("db.QueryAll table:schedule_tmeplate error:%v", err)
			return nil, 0, err
		}
		list[index] = ServiceListLogic{
			ServiceId:          item.Id,
			ServiceName:        item.ServiceName,
			ClusterNum:         1, // 目前一个服务只有一个集群
			Description:        item.Description,
			Language:           item.Language,
			ImageUrl:           imageUrl,
			ServiceClusterId:   modelCluster.Id,
			ServiceClusterName: modelCluster.ClusterName,
			AutoDecision:       modelCluster.AutoDecision,
			TmplExpandId:       modelTempLate.Id,
			TmplExpandName:     modelTempLate.TmplName,
			TaskTypeStatus:     modelTask.TaskStatus,
		}
	}
	return list, count, err
}

func (r *ServiceRepo) GetServiceDetail(ctx context.Context, serviceName string) (*ServiceDetailLogic, error) {
	var err error
	detailInfo := &ServiceDetailLogic{}
	serviceModel := db.Service{}
	serviceWhere := map[string]interface{}{
		"service_name": serviceName,
	}
	err = db.QueryFirst(serviceWhere, &serviceModel)
	if err != nil {
		log.Logger.Errorf("db.Querey  service_name:%v table [service] error:%v", serviceName, err)
		return nil, err
	}
	if serviceModel.Id > 0 {
		// 查询集群表
		serviceClusterModel := db.ServiceCluster{}
		serviceClusterWhere := map[string]interface{}{
			"service_name": serviceName,
		}
		err = db.QueryFirst(serviceClusterWhere, &serviceClusterModel)
		if err != nil {
			log.Logger.Errorf("service_name:%v table [service_cluster] error:%v", serviceName, err)
			return nil, err
		}
		if serviceClusterModel.Id > 0 {
			// 查询模版表
			templateModel := db.ScheduleTemplate{}
			templateWhere := map[string]interface{}{
				"service_cluster_id": serviceClusterModel.Id,
				"service_name":       serviceName,
			}
			err = db.QueryFirst(templateWhere, &templateModel)
			if err != nil {
				log.Logger.Errorf("service_name:%v table [schedule_template] error:%v", serviceName, err)
				return nil, err
			}
			detailInfo.TmplExpandId = templateModel.Id
			detailInfo.TmplExpandName = templateModel.TmplName
		}
		detailInfo.ServiceClusterName = serviceClusterModel.ClusterName
		detailInfo.ServiceClusterId = serviceClusterModel.Id
	}
	detailInfo.Description = serviceModel.Description
	detailInfo.ServiceName = serviceModel.ServiceName
	return detailInfo, nil
}

func (r *ServiceRepo) UpdateDesc(ctx context.Context, serviceName, description string) (int64, error) {
	var err error
	serviceModel := db.Service{}
	where := map[string]interface{}{
		"service_name": serviceName,
	}
	fields := map[string]interface{}{
		"description": description,
	}
	records, err := db.Updates(&serviceModel, where, fields, nil)
	if err != nil {
		log.Logger.Error(err)
		return 0, err
	}
	return records, nil
}
