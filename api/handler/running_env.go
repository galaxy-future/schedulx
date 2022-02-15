package handler

import (
	"net/http"
	"strings"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type CreateRunningEnvReq struct {
	Info *types.RunningEnv `json:"running_env_info"`
}

func CreateRunningEnv(ctx *gin.Context) {
	httpReq := &CreateRunningEnvReq{}
	err := ctx.BindJSON(httpReq)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if httpReq.Info == nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}

	err = service.GetServiceIns().CreateRunningEnv(ctx, httpReq.Info)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	MkResponse(ctx, http.StatusOK, errOK, nil)
	return
}

func DeleteRunningEnv(ctx *gin.Context) {
	idParam := ctx.Param("ids")
	input := strings.Split(idParam, ",")
	ids := make([]int64, 0, len(input))
	for _, v := range input {
		ids = append(ids, cast.ToInt64(v))
	}
	err := service.GetServiceIns().DeleteRunningEnv(ctx, ids)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	MkResponse(ctx, http.StatusOK, errOK, nil)
	return
}

type UpdateRunningEnvReq struct {
	Info *types.RunningEnv `json:"running_env_info"`
}

func UpdateRunningEnv(ctx *gin.Context) {
	httpReq := &UpdateRunningEnvReq{}
	err := ctx.BindJSON(httpReq)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	if httpReq.Info == nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}

	err = service.GetServiceIns().UpdateRunningEnv(ctx, httpReq.Info)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	MkResponse(ctx, http.StatusOK, errOK, nil)
	return
}

func GetRunningEnv(ctx *gin.Context) {
	id := ctx.Param("id")
	res, err := service.GetServiceIns().GetRunningEnvById(ctx, cast.ToInt64(id))
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	MkResponse(ctx, http.StatusOK, errOK, res)
	return
}
