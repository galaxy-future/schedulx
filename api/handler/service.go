package handler

import (
	"net/http"
	"strconv"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/register/constant"
	"github.com/galaxy-future/schedulx/repository"
	"github.com/galaxy-future/schedulx/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type Service struct{}

type ServiceExpandHttpRequest struct {
	ServiceClusterId int64  `form:"service_cluster_id" json:"service_cluster_id"`
	ServiceName      string `form:"service_name" json:"service_name" `
	ServiceCluster   string `form:"service_cluster" json:"service_cluster"`
	Count            int64  `form:"count" json:"count"`
	ExecType         string `form:"exec_type" json:"exec_type"`
}

type ServiceExpandHttpResponse struct {
	TaskId int64 `json:"task_id"`
}

type ServiceShrinkHttpRequest struct {
	ServiceClusterId int64  `form:"service_cluster_id" json:"service_cluster_id"`
	ServiceName      string `json:"service_name" form:"service_name"`
	ServiceCluster   string `json:"service_cluster" form:"service_cluster"`
	Count            int64  `form:"count" json:"count"`
	ExecType         string `form:"exec_type" json:"exec_type"`
}

type ServiceShrinkHttpResponse struct {
	TaskId int64 `json:"task_id"`
}

type ServiceCreateHttpRequest struct {
	ServiceInfo *types.ServiceInfo `json:"service_info"`
}

type ServiceCreateHttpResponse struct {
	ServiceClusterId int64 `json:"service_cluster_id"`
}

// Expand 服务扩容入口
func (h *Service) Expand(ctx *gin.Context) {
	var err error
	httpReq := &ServiceExpandHttpRequest{}
	err = ctx.BindQuery(httpReq)
	log.Logger.Infof("httpReq:%+v", httpReq)
	if err != nil {
		log.Logger.Error(err)
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	clusterId := cast.ToInt64(httpReq.ServiceClusterId)
	if clusterId == 0 {
		detail, err := service.GetServiceIns().Detail(ctx, httpReq.ServiceName)
		if err == nil && detail != nil {
			clusterId = detail["service_info"].(*repository.ServiceDetailLogic).ServiceClusterId
		}
	}

	if clusterId == 0 || cast.ToInt64(httpReq.Count) == 0 {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	if httpReq.ExecType == "" {
		httpReq.ExecType = constant.TaskExecTypeManual
	}
	scheduleSvc := service.GetScheduleSvcInst()
	tmplSvcReq := &service.ScheduleSvcReq{
		ServiceExpandSvcReq: &service.ServiceExpandSvcReq{
			ServiceClusterId: clusterId,
			Count:            httpReq.Count,
			ExecType:         httpReq.ExecType,
		},
	}
	resp, err := scheduleSvc.ExecAct(ctx, tmplSvcReq, scheduleSvc.Expand)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	data := &ServiceExpandHttpResponse{
		TaskId: resp.(*service.ScheduleSvcResp).ServiceExpandSvcResp.TaskId,
	}
	MkResponse(ctx, http.StatusOK, "success", data)
	return
}

// Shrink 服务缩容入口
func (h *Service) Shrink(ctx *gin.Context) {
	var err error
	httpReq := &ServiceShrinkHttpRequest{}
	err = ctx.BindQuery(httpReq)
	log.Logger.Infof("httpReq:%+v", httpReq)

	clusterId := cast.ToInt64(httpReq.ServiceClusterId)
	if clusterId == 0 {
		detail, err := service.GetServiceIns().Detail(ctx, httpReq.ServiceName)
		if err == nil && detail != nil {
			clusterId = detail["service_info"].(*repository.ServiceDetailLogic).ServiceClusterId
		}
	}

	if clusterId == 0 || cast.ToInt64(httpReq.Count) == 0 {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	if httpReq.ExecType == "" {
		httpReq.ExecType = constant.TaskExecTypeManual
	}
	scheduleSvc := service.GetScheduleSvcInst()
	tmplSvcReq := &service.ScheduleSvcReq{
		ServiceShrinkSvcReq: &service.ServiceShrinkSvcReq{
			ServiceClusterId: clusterId,
			Count:            httpReq.Count,
			ExecType:         httpReq.ExecType,
		},
	}
	resp, err := scheduleSvc.ExecAct(ctx, tmplSvcReq, scheduleSvc.Shrink)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	data := &ServiceShrinkHttpResponse{
		TaskId: resp.(*service.ScheduleSvcResp).ServiceShrinkSvcResp.TaskId,
	}
	MkResponse(ctx, http.StatusOK, "success", data)
	return
}

// Detail 查询服务详情
func (h *Service) Detail(ctx *gin.Context) {
	var err error
	serviceName := ctx.Query("service_name")
	if serviceName == "" {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	detail, err := service.GetServiceIns().Detail(ctx, serviceName)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	MkResponse(ctx, http.StatusOK, errOK, detail)
	return
}

func (h *Service) Scheduling(ctx *gin.Context) {
	serviceClusterName := ctx.Query("service_cluster_name")
	serviceName := ctx.Query("service_name")
	if serviceName == "" || serviceClusterName == "" {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	scheduling, err := service.GetTaskSvcInst().HasRunningTask(ctx, serviceName, serviceClusterName)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	MkResponse(ctx, http.StatusOK, errOK, gin.H{
		"service_name":         serviceName,
		"service_cluster_name": serviceClusterName,
		"scheduling":           scheduling,
	})
	return
}

// List 查询服务列表
func (h *Service) List(ctx *gin.Context) {
	var err error
	serviceName := ctx.Query("service_name")
	language := ctx.Query("language")
	pageNum := ctx.Query("page_num")
	pageSize := ctx.Query("page_size")

	pageNumInt, err := strconv.Atoi(pageNum)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	list, err := service.GetServiceIns().GetServiceList(ctx, pageNumInt, pageSizeInt, serviceName, language)
	if err != nil {
		MkResponse(ctx, http.StatusOK, err.Error(), list)
		return
	}
	MkResponse(ctx, http.StatusOK, errOK, list)
	return
}

// ClusterList 查询服务关联集群列表
func (h *Service) ClusterList(ctx *gin.Context) {
	var err error
	serviceName := ctx.Query("service_name")
	list, err := service.GetServiceIns().GetServiceClusterList(ctx, serviceName)
	if err != nil {
		MkResponse(ctx, http.StatusOK, err.Error(), list)
		return
	}
	MkResponse(ctx, http.StatusOK, errOK, list)
	return
}

// BreathRecord 查询单个服务扩容历史
func (h *Service) BreathRecord(ctx *gin.Context) {
	var err error
	serviceClusterId := ctx.Query("service_cluster_id")
	scIdInt64, err := strconv.Atoi(serviceClusterId)
	if err != nil || scIdInt64 == 0 {
		MkResponse(ctx, http.StatusOK, errParamInvalid, nil)
		return
	}
	pageNum := ctx.Query("page_num")
	pageSize := ctx.Query("page_size")

	pageNumInt, err := strconv.Atoi(pageNum)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	historyList, err := service.GetServiceIns().GetExpandHistory(ctx, pageNumInt, pageSizeInt, scIdInt64)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	MkResponse(ctx, http.StatusOK, errOK, historyList)
}

// Create 创建服务
func (h *Service) Create(ctx *gin.Context) {
	var err error
	httpReq := &ServiceCreateHttpRequest{}
	err = ctx.BindJSON(httpReq)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if httpReq.ServiceInfo.ServiceName == "" {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}

	serviceSvc := service.GetServiceIns()
	serviceSvcReq := &service.ServiceCreateSvcRequest{
		ServiceInfo: httpReq.ServiceInfo,
	}
	serviceSvcResp, err := serviceSvc.CreateService(ctx, serviceSvcReq)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	data := &ServiceCreateHttpResponse{
		ServiceClusterId: serviceSvcResp.ServiceClusterId,
	}
	MkResponse(ctx, http.StatusOK, "success", data)
	return
}

// Update 更新数据表记录
func (h *Service) Update(ctx *gin.Context) {
	var err error
	var params = struct {
		ServiceInfo struct {
			ServiceName string `json:"service_name"`
			Description string `json:"description"`
		} `json:"service_info"`
	}{}
	err = ctx.BindJSON(&params)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	ret, err := service.GetServiceIns().UpdateDesc(ctx, params.ServiceInfo.ServiceName, params.ServiceInfo.Description)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	MkResponse(ctx, http.StatusOK, errOK, ret)
	return
}
