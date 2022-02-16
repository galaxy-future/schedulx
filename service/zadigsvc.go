package service

import (
	"context"
	"github.com/galaxy-future/schedulx/client/zadigcli"
	"sync"
)

type ZadigSvc struct {
}

var (
	zadigSvc     *ZadigSvc
	zadigSvcOnce sync.Once
)

func GetZadigSvcInst() *ZadigSvc {
	zadigSvcOnce.Do(func() {
		zadigSvc = &ZadigSvc{}
	})
	return zadigSvc
}

func (s *ZadigSvc) GetWorkflows(ctx context.Context, serviceName string) ([]*zadigcli.WorkflowResp, error) {
	zadig := GetIntegrationService()
	req := zadigcli.WorkflowsReq{
		ZadigHost:   zadig.GetZadigHost(ctx),
		ZadigToken:  zadig.GetZadigToken(),
		ProjectName: serviceName,
	}
	return zadigcli.GetZadigXCli(ctx).GetWorkflows(ctx, req)
}

func (s *ZadigSvc) GetWorkflowTasks(ctx context.Context, serviceName, workflowName, fileType string, pageNum, pageSize int) (int, []*zadigcli.WorkflowTaskResp, error) {
	zadig := GetIntegrationService()
	req := zadigcli.WorkflowTasksReq{
		ZadigHost:    zadig.GetZadigHost(ctx),
		ZadigToken:   zadig.GetZadigToken(),
		ProjectName:  serviceName,
		WorkflowName: workflowName,
		FileType:     fileType,
		Skip:         (pageNum - 1) * pageSize,
		Limit:        pageSize,
	}
	return zadigcli.GetZadigXCli(ctx).GetWorkflowTasks(ctx, req)
}
