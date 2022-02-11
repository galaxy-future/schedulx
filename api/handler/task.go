package handler

import (
	"errors"
	"net/http"

	"gorm.io/gorm"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type Task struct{}

type TaskInfoRequest struct {
	TaskId int64 `json:"task_id"`
}

type TaskInfoResponse struct {
	TaskInfo *types.TaskInfo `json:"task_info"`
}

// Info 查询任务进度信息
func (h *Task) Info(ctx *gin.Context) {
	var err error
	taskId := ctx.Query("task_id")
	if cast.ToInt64(taskId) == 0 {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	svcReq := &service.TaskInfoSvcReq{
		TaskId: cast.ToInt64(taskId),
	}
	svc := service.GetTaskSvcInst()
	svcResp, err := svc.Info(ctx, svcReq)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		MkResponse(ctx, http.StatusOK, "record not found", nil)
		return
	}
	data := &TaskInfoResponse{
		TaskInfo: svcResp.TaskInfo,
	}
	MkResponse(ctx, http.StatusOK, errOK, data)
	return
}

// GetDeployDetail 查询部署任务详情
func (h *Task) GetDeployDetail(ctx *gin.Context) {
	var err error
	serviceClusterId := ctx.Query("service_cluster_id")
	taskId := ctx.Query("task_id")
	if cast.ToInt64(serviceClusterId) == 0 {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	task, err := service.GetTaskSvcInst().GetRunningTask(ctx, cast.ToInt64(serviceClusterId), cast.ToInt64(taskId))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		MkResponse(ctx, http.StatusOK, "record not found", nil)
		return
	}

	MkResponse(ctx, http.StatusOK, errOK, task)
	return
}
