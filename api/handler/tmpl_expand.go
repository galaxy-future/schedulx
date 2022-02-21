package handler

import (
	"net/http"
	"strconv"

	"github.com/galaxy-future/schedulx/pkg/tool"

	"github.com/spf13/cast"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/service"
	"github.com/gin-gonic/gin"
)

type TmplExpand struct {
}

type TmplExpandRequest struct {
	EndStep    string             `json:"end_step"`
	TmplInfo   *types.TmpInfo     `json:"tmpl_info"`
	BaseEnv    *types.BaseEnv     `json:"base_env"`
	ServiceEnv *types.ServiceEnv  `json:"service_env"`
	Mount      *types.ParamsMount `json:"mount"`
	DeployInfo *types.DeployInfo  `json:"deploy_info"`
}

type TmplUpdateRequest struct {
	TmplExpandId int64              `json:"tmpl_expand_id"`
	EndStep      string             `json:"end_step"`
	TmplInfo     *types.TmpInfo     `json:"tmpl_info"`
	BaseEnv      *types.BaseEnv     `json:"base_env"`
	ServiceEnv   *types.ServiceEnv  `json:"service_env"`
	Mount        *types.ParamsMount `json:"mount"`
	DeployInfo   *types.DeployInfo  `json:"deploy_info"`
}

type TmplInfoResponse struct {
	TmplInfo   *types.TmpInfo     `json:"tmpl_info"`
	BaseEnv    *types.BaseEnv     `json:"base_env"`
	ServiceEnv *types.ServiceEnv  `json:"service_env"`
	Mount      *types.ParamsMount `json:"mount"`
	DeployInfo *types.DeployInfo  `json:"deploy_info"`
}

type TmplExpandResponse struct {
	TmplId       string `json:"tmpl_id"`
	DeployTmplId string `json:"deploy_tmpl_id"`
}

type TmplUpdateResponse struct {
	TmplId string `json:"tmpl_id"`
}

// Create 模板
func (h *TmplExpand) Create(ctx *gin.Context) {
	var err error
	tmplExpandReq := &TmplExpandRequest{}
	err = ctx.BindJSON(tmplExpandReq)
	tmplSvc := service.GetTemplateSvcInst()
	tmplSvcReq := &service.TemplateSvcReq{
		TmplExpandSvcReq: &service.TmplExpandSvcReq{
			EndStep:    tmplExpandReq.EndStep,
			TmplInfo:   tmplExpandReq.TmplInfo,
			BaseEnv:    tmplExpandReq.BaseEnv,
			ServiceEnv: tmplExpandReq.ServiceEnv,
			Mount:      tmplExpandReq.Mount,
			DeployInfo: tmplExpandReq.DeployInfo,
		},
	}
	svcResp, err := tmplSvc.ExecAct(ctx, tmplSvcReq, tmplSvc.Create)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}
	resp := svcResp.(*service.TemplateSvcResp)
	log.Logger.Infof("resp info :%v", tool.ToJson(resp))
	tmplExpandResp := &TmplExpandResponse{
		TmplId:       resp.TmplExpandSvcResp.TmplId,
		DeployTmplId: resp.TmplExpandSvcResp.DeployTmplId,
	}
	MkResponse(ctx, http.StatusOK, "success", tmplExpandResp)
	return
}

func (h *TmplExpand) List(ctx *gin.Context) {
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
	list, err := service.GetTemplateSvcInst().List(ctx, serviceName, pageNumInt, pageSizeInt, scIdInt)
	if err != nil {
		MkResponse(ctx, http.StatusOK, err.Error(), list)
		return
	}
	MkResponse(ctx, http.StatusOK, "success", list)
	return
}

func (h *TmplExpand) Info(ctx *gin.Context) {
	TmplExpandId := ctx.Query("tmpl_expand_id")
	if TmplExpandId == "" {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	tmplSvc := service.GetTemplateSvcInst()
	svcReq := &service.TemplateSvcReq{
		TmplInfoSvcReq: &service.TmplInfoSvcReq{
			TmplExpandId: cast.ToInt64(TmplExpandId),
		},
	}
	svcResp, err := tmplSvc.ExecAct(ctx, svcReq, tmplSvc.Info)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp := svcResp.(*service.TemplateSvcResp)
	data := &TmplInfoResponse{
		TmplInfo:   resp.TmplInfoSvcResp.TmplInfo,
		BaseEnv:    resp.TmplInfoSvcResp.BaseEnv,
		ServiceEnv: resp.TmplInfoSvcResp.ServiceEnv,
		Mount:      resp.TmplInfoSvcResp.Mount,
		DeployInfo: resp.TmplInfoSvcResp.DeployInfo,
	}
	MkResponse(ctx, http.StatusOK, "success", data)
	return
}

func (h *TmplExpand) Update(ctx *gin.Context) {
	var err error
	tmplUpdateReq := &TmplUpdateRequest{}
	err = ctx.BindJSON(tmplUpdateReq)
	tmplSvc := service.GetTemplateSvcInst()
	tmplSvcReq := &service.TemplateSvcReq{
		TmplUpdateSvcReq: &service.TmplUpdateSvcReq{
			TmplExpandId: tmplUpdateReq.TmplExpandId,
			EndStep:      tmplUpdateReq.EndStep,
			TmplInfo:     tmplUpdateReq.TmplInfo,
			BaseEnv:      tmplUpdateReq.BaseEnv,
			ServiceEnv:   tmplUpdateReq.ServiceEnv,
			Mount:        tmplUpdateReq.Mount,
			DeployInfo:   tmplUpdateReq.DeployInfo,
		},
	}
	svcResp, err := tmplSvc.ExecAct(ctx, tmplSvcReq, tmplSvc.Update)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp := svcResp.(*service.TemplateSvcResp)
	log.Logger.Infof("resp:", tool.ToJson(resp))
	tmplExpandResp := &TmplExpandResponse{
		TmplId: tool.Interface2String(tmplSvcReq.TmplUpdateSvcReq.TmplExpandId),
	}
	MkResponse(ctx, http.StatusOK, "success", tmplExpandResp)
	return
}

// Delete 删除模版信息
func (h *TmplExpand) Delete(ctx *gin.Context) {
	var err error
	var params = struct {
		TmplExpandIds []int64 `json:"tmpl_expand_id"`
	}{}
	err = ctx.BindJSON(&params)
	if err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, "参数为整数数组")
		return
	}
	ret, err := service.GetTemplateSvcInst().Delete(ctx, params.TmplExpandIds)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	MkResponse(ctx, http.StatusOK, "success", ret)
	return
}
