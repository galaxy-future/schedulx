package handler

import (
	"net/http"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/repository"
	"github.com/galaxy-future/schedulx/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type Instance struct{}

type InstanceReq struct {
	PageNum        int    `form:"page_num" json:"page_num"`
	PageSize       int    `form:"page_size" json:"page_size"`
	TaskId         int    `form:"task_id" json:"task_id"`
	InstanceStatus string `form:"instance_status" json:"instance_status"`
}

type InstanceCountReq struct {
	ServiceClusterId   int64  `json:"service_cluster_id" form:"service_cluster_id"`
	ServiceName        string `json:"service_name" form:"service_name"`
	ServiceClusterName string `json:"service_cluster_name" form:"service_cluster_name"`
}

func (i *Instance) Count(ctx *gin.Context) {
	req := &InstanceCountReq{}
	err := ctx.BindQuery(req)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	if req.ServiceName == "" && req.ServiceClusterId == 0 {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	resp, err := service.GetInstanceService().InstanceCountByCluster(ctx, req.ServiceName, req.ServiceClusterName, req.ServiceClusterId)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	MkResponse(ctx, http.StatusOK, errOK, resp)
	return
}

func (t *Task) InstanceList(ctx *gin.Context) {
	req := &InstanceReq{}
	err := ctx.BindQuery(req)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	if req.TaskId <= 0 {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, "task_id empty")
		return
	}

	taskId, err := repository.GetTaskRepoInst().GetBridgXTaskId(ctx, cast.ToInt64(req.TaskId))
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	total, instances, err := service.GetTaskSvcInst().InstanceList(ctx, req.PageNum, req.PageSize, taskId, types.InstanceStatus(req.InstanceStatus))
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	ret := struct {
		Pager        types.Pager          `json:"pager"`
		InstanceList []types.InstInfoResp `json:"instance_list"`
		TaskId       int                  `json:"task_id"`
	}{
		Pager: types.Pager{
			PagerNum:  req.PageNum,
			PagerSize: req.PageSize,
			Total:     int(total),
		},
		InstanceList: instances,
		TaskId:       req.TaskId,
	}
	MkResponse(ctx, http.StatusOK, errOK, ret)
	return
}
