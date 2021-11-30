package handler

import (
	"net/http"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/service"
	"github.com/gin-gonic/gin"
)

type InstanceReq struct {
	PageNum        int    `form:"page_num" json:"page_num"`
	PageSize       int    `form:"page_size" json:"page_size"`
	TaskId         int    `form:"task_id" json:"task_id"`
	InstanceStatus string `form:"instance_status" json:"instance_status"`
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
	total, instances, err := service.GetTaskSvcInst().InstanceList(ctx, req.PageNum, req.PageSize, req.TaskId, types.InstanceStatus(req.InstanceStatus))
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
