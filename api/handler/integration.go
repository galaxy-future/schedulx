package handler

import (
	"net/http"

	"github.com/galaxy-future/schedulx/repository"
	"github.com/gin-gonic/gin"
)

type Integration struct{}

type CreateIntegrationInfoReq struct {
	Host     string `json:"host"`
	Account  string `json:"account"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

type DeleteIntegrationReq struct {
	Id int `json:"id"`
}

type ListIntegrationReq struct {
	Type     string `json:"type" form:"type"`
	PageNum  int    `json:"page_num" form:"page_num"`
	PageSize int    `json:"page_size" form:"page_size"`
}

func (i *Integration) Create(ctx *gin.Context) {
	req := &CreateIntegrationInfoReq{}
	if err := ctx.ShouldBind(req); err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	err := repository.GetIntegrationInstance().Create(ctx, req.Host, req.Account, req.Password, req.Type)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	MkResponse(ctx, http.StatusOK, errOK, nil)
	return
}

func (i *Integration) Delete(ctx *gin.Context) {
	req := &DeleteIntegrationReq{}
	if err := ctx.ShouldBind(req); err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	err := repository.GetIntegrationInstance().Delete(ctx, req.Id)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	MkResponse(ctx, http.StatusOK, errOK, nil)
	return
}

func (i *Integration) List(ctx *gin.Context) {
	req := &ListIntegrationReq{}
	if err := ctx.ShouldBind(req); err != nil {
		MkResponse(ctx, http.StatusBadRequest, errParamInvalid, nil)
		return
	}
	pageNum, pageSize := parsePage(req.PageNum, req.PageSize)
	res, total, err := repository.GetIntegrationInstance().List(ctx, req.Type, pageNum, pageSize)
	if err != nil {
		MkResponse(ctx, http.StatusInternalServerError, err.Error(), gin.H{
			"integration_list": res,
			"pager": Pager{
				PageNumber: pageNum,
				PageSize:   pageSize,
				Total:      total,
			},
		})
		return
	}
	MkResponse(ctx, http.StatusOK, errOK, nil)
	return
}

func parsePage(pageNum, pageSize int) (int, int) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return pageNum, pageSize
}
