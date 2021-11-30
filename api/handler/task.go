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

// Info 查询任务详情
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
