package handler

import (
	"net/http"
	"strconv"

	"github.com/galaxy-future/schedulx/service"
	"github.com/gin-gonic/gin"
)

type TmplDeploy struct {
}

func (h *TmplDeploy) List(ctx *gin.Context) {
	var err error
	serviceName := ctx.Query("service_name")
	scId := ctx.Query("service_cluster_id")
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
	scIdInt, err := strconv.Atoi(scId)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	list, err := service.GetTemplateSvcInst().ListDeployTemplates(ctx, serviceName, pageNumInt, pageSizeInt, scIdInt)
	if err != nil {
		MkResponse(ctx, http.StatusOK, err.Error(), list)
		return
	}
	MkResponse(ctx, http.StatusOK, "success", list)
	return
}
