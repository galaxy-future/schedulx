package service

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/galaxy-future/schedulx/repository"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/pkg/tool"
	"github.com/galaxy-future/schedulx/register/config"
	"github.com/galaxy-future/schedulx/register/config/client"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/register/constant"
	"github.com/galaxy-future/schedulx/repository/model/db"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

const (
	ExpandStepTmplInfo   = "tmpl_info"
	ExpandStepBaseEnv    = "base_env"
	ExpandStepServiceEnv = "service_env"
	ExpandStepMount      = "mount"
)

var (
	expandStep = []string{ExpandStepTmplInfo, ExpandStepBaseEnv, ExpandStepServiceEnv, ExpandStepMount} // 标准扩容执行步骤
)

type TemplateSvc struct {
	Create types.Action
	Info   types.Action
	Update types.Action
}

var templateSvc *TemplateSvc
var templateOnce sync.Once

func GetTemplateSvcInst() *TemplateSvc {
	templateOnce.Do(func() {
		templateSvc = &TemplateSvc{
			"Create",
			"Info",
			"update",
		}
	})
	return templateSvc
}

type TemplateSvcReq struct {
	TmplExpandSvcReq *TmplExpandSvcReq
	TmplInfoSvcReq   *TmplInfoSvcReq
	TmplUpdateSvcReq *TmplUpdateSvcReq
}

type TemplateSvcResp struct {
	TmplExpandSvcResp *TmplExpandSvcResp
	TmplInfoSvcResp   *TmplInfoSvcResp
	TmplUpdateSvcResp *TmplUpdateSvcResp
}

type TmplExpandSvcReq struct {
	EndStep    string             `json:"end_step"`
	TmplInfo   *types.TmpInfo     `json:"tmpl_info"`
	BaseEnv    *types.BaseEnv     `json:"base_env"`
	ServiceEnv *types.ServiceEnv  `json:"service_env"`
	Mount      *types.ParamsMount `json:"mount"`
}

type TmplUpdateSvcReq struct {
	TmplExpandId int64              `json:"tmpl_expand_id"`
	EndStep      string             `json:"end_step"`
	TmplInfo     *types.TmpInfo     `json:"tmpl_info"`
	BaseEnv      *types.BaseEnv     `json:"base_env"`
	ServiceEnv   *types.ServiceEnv  `json:"service_env"`
	Mount        *types.ParamsMount `json:"mount"`
}

type TmplInfoSvcReq struct {
	TmplExpandId int64 `json:"tmpl_expand_id"`
}
type TmplInfoSvcResp struct {
	TmplInfo   *types.TmpInfo     `json:"tmpl_info"`
	BaseEnv    *types.BaseEnv     `json:"base_env"`
	ServiceEnv *types.ServiceEnv  `json:"service_env"`
	Mount      *types.ParamsMount `json:"mount"`
}

type TmplExpandSvcResp struct {
	TmplId string `json:"tmpl_id"`
}

type TmplUpdateSvcResp struct {
	TmplExpandId int64 `json:"tmpl_expand_id"`
}

func (s *TemplateSvc) entryLog(ctx context.Context, act string, req interface{}) {
	log.Logger.Infof("entry log | act[%s] | req:%s", act, tool.ToJson(req))
}

func (s *TemplateSvc) exitLog(ctx context.Context, act string, req, resp interface{}, err error) {
	log.Logger.Infof("exit log | act[%s] | req:%s | resp:%s | err:%v", act, tool.ToJson(req), tool.ToJson(resp), err)
}

func (s *TemplateSvc) ExecAct(ctx context.Context, args interface{}, act types.Action) (svcResp interface{}, err error) {
	svcReq, ok := args.(*TemplateSvcReq)
	if !ok {
		return nil, errors.New("init service request assertion err")
	}
	s.entryLog(ctx, string(act), svcReq)
	defer func() {
		s.exitLog(ctx, string(act), svcReq, svcResp, err)
	}()
	switch act {
	case s.Create:
		svcResp, err = s.createAction(ctx, svcReq.TmplExpandSvcReq)
	case s.Info:
		svcResp, err = s.InfoAction(ctx, svcReq.TmplInfoSvcReq)
	case s.Update:
		svcResp, err = s.UpdateAction(ctx, svcReq.TmplUpdateSvcReq)
	default:
		err = errors.New("no act matched")
		return nil, err
	}

	return svcResp, err
}

func (s *TemplateSvc) createAction(ctx context.Context, svcReq *TmplExpandSvcReq) (*TemplateSvcResp, error) {
	svcResp := &TemplateSvcResp{}
	var err error
	//var tmplId int64
	var tmplInfo, revTmplInfo *db.ScheduleTemplate
	var instrGroup, instrReverseGroup []int64
	dbo := client.WriteDBCli.Debug().Begin()
	defer func() { // 事务保证
		if err != nil {
			dbo.Rollback()
			return
		}
		dbo.Commit()
	}()
	serviceCluster := &db.ServiceCluster{}
	serviceClusterId := cast.ToInt64(svcReq.TmplInfo.ServiceClusterId)
	if err = db.Get(serviceClusterId, serviceCluster); err != nil {
		log.Logger.Errorf("db tabel:%v error:%v", serviceCluster.TableName(), err)
		return nil, err
	}

	if err = db.UpdatesByIds(serviceCluster, []int64{serviceClusterId}, map[string]interface{}{
		"bridgx_cluster": svcReq.TmplInfo.BridgxClusname,
	}, dbo); err != nil {
		log.Logger.Errorf("db tabel:%v error:%v", serviceCluster.TableName(), err)
		return nil, err
	}

	for _, step := range expandStep {
		var instrId int64
		var reverseInstrId int64
		switch step {
		case ExpandStepTmplInfo:
			tmplInfo, revTmplInfo, err = s.createExpandTmpl(ctx, svcReq.TmplInfo, serviceCluster.ServiceName, true, dbo)
			if err != nil {
				return nil, err
			}
			svc := GetInstrSvcInst()
			instrId, reverseInstrId, err = svc.CreateBridgxExpandInstr(ctx, tmplInfo.Id, revTmplInfo.Id, true, dbo)
			if err != nil {
				return nil, err
			}
			instrGroup = append(instrGroup, instrId)
			if reverseInstrId != 0 {
				instrReverseGroup = append(instrReverseGroup, reverseInstrId)
			}
		case ExpandStepBaseEnv:
			svc := GetInstrSvcInst()
			instrId, reverseInstrId, err = svc.CreateBaseEnvInstr(ctx, svcReq.BaseEnv, tmplInfo.Id, false, dbo)
			if err != nil {
				return nil, err
			}
			instrGroup = append(instrGroup, instrId)
			if reverseInstrId != 0 {
				instrReverseGroup = append(instrReverseGroup, reverseInstrId)
			}
		case ExpandStepServiceEnv:
			svc := GetInstrSvcInst()
			svcReq.ServiceEnv.ServiceName = serviceCluster.ServiceName
			instrId, reverseInstrId, err = svc.CreateServiceEnvInstr(ctx, svcReq.ServiceEnv, tmplInfo.Id, false, dbo)
			if err != nil {
				return nil, err
			}
			instrGroup = append(instrGroup, instrId)
			if reverseInstrId != 0 {
				instrReverseGroup = append(instrReverseGroup, reverseInstrId)
			}
		case ExpandStepMount:
			svc := GetInstrSvcInst()
			instrId, reverseInstrId, err = svc.CreateMountSlbInstr(ctx, svcReq.Mount, tmplInfo.Id, revTmplInfo.Id, true, dbo)
			if err != nil {
				return nil, err
			}
			instrGroup = append(instrGroup, instrId)
			if reverseInstrId != 0 {
				instrReverseGroup = append(instrReverseGroup, reverseInstrId)
			}
		}
		if step == svcReq.EndStep {
			break
		}
	}
	// 生成逆向 instrGroup , 更新逆向 tmpl
	instrReverseGroup = tool.ReverseIntSlice(instrReverseGroup)
	err = s.updateTmplInstrGroup(ctx, revTmplInfo.Id, instrReverseGroup, tmplInfo.Id, dbo)
	if err != nil {
		return nil, err
	}
	err = s.updateTmplInstrGroup(ctx, tmplInfo.Id, instrGroup, revTmplInfo.Id, dbo)
	if err != nil {
		return nil, err
	}
	svcResp.TmplExpandSvcResp = &TmplExpandSvcResp{
		TmplId: cast.ToString(tmplInfo.Id),
	}

	return svcResp, nil
}

func (s *TemplateSvc) createExpandTmpl(ctx context.Context, args *types.TmpInfo, serviceName string, needReverse bool, dbo *gorm.DB) (*db.ScheduleTemplate, *db.ScheduleTemplate, error) {
	var err error
	obj := &db.ScheduleTemplate{
		TmplName:         args.TmplName,
		ServiceName:      serviceName,
		ServiceClusterId: cast.ToInt64(args.ServiceClusterId),
		BridgxClusname:   args.BridgxClusname,
		Description:      args.Describe,
		ScheduleType:     constant.ScheduleTypeExpand,
	}
	err = db.Create(obj, dbo)
	if err != nil {
		log.Logger.Error(err)
		return nil, nil, err
	}
	var reverseObj *db.ScheduleTemplate
	if needReverse {
		reverseObj, err = s.createShrinkTmpl(ctx, obj, dbo)
		if err != nil {
			return nil, nil, err
		}
		return obj, reverseObj, nil
	}
	return obj, nil, nil
}

func (s *TemplateSvc) updateTmplInfo(ctx context.Context, tmplExpandId int64, args *types.TmpInfo, dbo *gorm.DB) error {
	var err error
	where := map[string]interface{}{
		"id": tmplExpandId,
	}
	updates := map[string]interface{}{
		"tmpl_name":       args.TmplName,
		"bridgx_clusname": args.BridgxClusname,
		"description":     args.Describe,
	}
	rowsAffected, err := db.Updates(&db.ScheduleTemplate{}, where, updates, dbo)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	if rowsAffected != 1 {
		err = config.ErrRowsAffectedInvalid
		log.Logger.Error(err)
		return err
	}

	return nil
}

func (s *TemplateSvc) updateReverseTmplInfo(ctx context.Context, tmplExpandId int64, args *types.TmpInfo, dbo *gorm.DB) error {
	var err error
	where := map[string]interface{}{
		"id": tmplExpandId,
	}
	updates := map[string]interface{}{
		"tmpl_name":       "[逆向]" + args.TmplName,
		"bridgx_clusname": args.BridgxClusname,
	}
	rowsAffected, err := db.Updates(&db.ScheduleTemplate{}, where, updates, dbo)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	if rowsAffected != 1 {
		err = config.ErrRowsAffectedInvalid
		log.Logger.Error(err)
		return err
	}

	return nil
}

func (s *TemplateSvc) createShrinkTmpl(ctx context.Context, tmpl *db.ScheduleTemplate, dbo *gorm.DB) (*db.ScheduleTemplate, error) {
	var err error
	obj := &db.ScheduleTemplate{
		TmplName:           "[逆向]" + tmpl.TmplName,
		ServiceName:        tmpl.ServiceName,
		ServiceClusterId:   tmpl.ServiceClusterId,
		BridgxClusname:     tmpl.BridgxClusname,
		ScheduleType:       constant.ScheduleTypeShrink,
		ReverseSchedTmplId: tmpl.Id,
	}
	err = db.Create(obj, dbo)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	return obj, nil
}

func (s *TemplateSvc) updateTmplInstrGroup(ctx context.Context, tmplId int64, instrGroup []int64, reverseTmplId int64, dbo *gorm.DB) error {
	var err error
	ig, _ := jsoniter.MarshalToString(instrGroup)
	data := map[string]interface{}{
		"instr_group":           ig,
		"reverse_sched_tmpl_id": reverseTmplId,
	}
	where := map[string]interface{}{
		"id": tmplId,
	}
	rowAffected, err := db.Updates(&db.ScheduleTemplate{}, where, data, dbo)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	if rowAffected != 1 {
		err = config.ErrRowsAffectedInvalid
		log.Logger.Error(err)
		return err
	}

	return nil
}

func (s *TemplateSvc) List(ctx context.Context, serviceName string, page, pageSize, serviceClusterId int) (map[string]interface{}, error) {
	var err error
	list, total, err := repository.GetScheduleTemplateRepoInst().GetExpandList(ctx, serviceName, page, pageSize, serviceClusterId)
	if err != nil {
		log.Logger.Errorf(" func list error:%v", err)
	}
	ret := map[string]interface{}{
		"tmpl_expand_list": list,
		"pager": struct {
			PageNumber int   `json:"page_number"`
			PageSize   int   `json:"page_size"`
			Total      int64 `json:"total"`
		}{
			PageSize:   pageSize,
			PageNumber: page,
			Total:      total,
		},
	}
	return ret, nil
}

func (s *TemplateSvc) InfoAction(ctx context.Context, svcReq *TmplInfoSvcReq) (*TemplateSvcResp, error) {
	var err error
	svcResp := &TmplInfoSvcResp{}
	tmplRepo := repository.GetScheduleTemplateRepoInst()
	tmpl, err := tmplRepo.GetSchedTmpl(svcReq.TmplExpandId)
	templInfo := &types.TmpInfo{
		TmplName:         tmpl.TmplName,
		ServiceClusterId: tmpl.ServiceClusterId,
		Describe:         tmpl.Description,
		BridgxClusname:   tmpl.BridgxClusname,
	}
	svcResp.TmplInfo = templInfo
	var instrGroup []int64
	if err = jsoniter.Unmarshal([]byte(tmpl.InstrGroup), &instrGroup); err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	instrRepo := repository.GetInstrRepoInst()
	instrSvc := GetInstrSvcInst()
	for _, instrId := range instrGroup {
		instrInfo, rRrr := instrRepo.GetInstr(ctx, instrId)
		if rRrr != nil && err != gorm.ErrRecordNotFound {
			log.Logger.Errorf("func [instrRepo.GetInstr] error:%v", rRrr)
			return nil, rRrr
		}
		switch instrInfo.InstrAction {
		case instrSvc.NodeActInitBase:
			params := &types.BaseEnv{}
			if err = jsoniter.Unmarshal([]byte(instrInfo.Params), params); err != nil {
				log.Logger.Error(err)
				return nil, err
			}
			svcResp.BaseEnv = params
		case instrSvc.NodeActInitSvc:
			params := &types.ParamsServiceEnv{}
			if err = jsoniter.Unmarshal([]byte(instrInfo.Params), params); err != nil {
				log.Logger.Error(err)
				return nil, err
			}
			var password []byte
			password, err = tool.AesDecrypt(params.Password, []byte(params.Account))
			if err != nil {
				log.Logger.Error(err)
				return nil, err
			}
			svcResp.ServiceEnv = &types.ServiceEnv{
				ImageStorageType: params.ImageStorageType,
				ImageUrl:         params.ImageUrl,
				Port:             params.Port,
				Account:          params.Account,
				Password:         string(password),
				Cmd:              instrInfo.Cmd,
				ServiceName:      tmpl.ServiceName,
			}
		case instrSvc.MountSLB:
			params := &types.ParamsMount{}
			if err = jsoniter.Unmarshal([]byte(instrInfo.Params), params); err != nil {
				log.Logger.Error(err)
				return nil, err
			}
			svcResp.Mount = params
		}
	}
	return &TemplateSvcResp{TmplInfoSvcResp: svcResp}, nil
}

func (s *TemplateSvc) UpdateAction(ctx context.Context, svcReq *TmplUpdateSvcReq) (*TemplateSvcResp, error) {
	svcResp := &TmplUpdateSvcResp{}
	var err error
	var instrGroup = make([]int64, 0)
	var instrReverseGroup = make([]int64, 0)
	dbo := client.WriteDBCli.Debug().Begin()
	defer func() { // 事务保证
		if err != nil {
			dbo.Rollback()
			return
		}
		dbo.Commit()
	}()
	serviceCluster := &db.ServiceCluster{}
	serviceClusterId := cast.ToInt64(svcReq.TmplInfo.ServiceClusterId)
	if err = db.Get(serviceClusterId, serviceCluster); err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	if serviceCluster.BridgxCluster != svcReq.TmplInfo.BridgxClusname {
		if err = db.UpdatesByIds(serviceCluster, []int64{serviceClusterId}, map[string]interface{}{
			"bridgx_cluster": svcReq.TmplInfo.BridgxClusname,
		}, dbo); err != nil {
			log.Logger.Errorf("db tabel:%v error:%v", serviceCluster.TableName(), err)
			return nil, err
		}

	}

	tmplRepo := repository.GetScheduleTemplateRepoInst()
	oriTmpl, err := tmplRepo.GetSchedTmpl(svcReq.TmplExpandId)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}

	for _, step := range expandStep {
		var instrId int64
		var reverseInstrId int64
		switch step {
		case ExpandStepTmplInfo:
			err = s.updateTmplInfo(ctx, svcReq.TmplExpandId, svcReq.TmplInfo, dbo)
			if err != nil && !strings.Contains(err.Error(), "db update rows affected invalid") {
				log.Logger.Errorf("repository func : updateTmplInfo update error:%v", err)
				return nil, err
			}
			err = s.updateReverseTmplInfo(ctx, oriTmpl.ReverseSchedTmplId, svcReq.TmplInfo, dbo)
			if err != nil && !strings.Contains(err.Error(), "db update rows affected invalid") {
				log.Logger.Errorf("repository func : updateReverseTmplInfo error:%v", err)
				return nil, err
			}
			svc := GetInstrSvcInst()
			//设置原 instruction 为 is_deleted
			instrRepo := repository.GetInstrRepoInst()
			instrRepo.DeleteByTmplExpandId(ctx, oriTmpl.Id, dbo)
			instrRepo.DeleteByTmplExpandId(ctx, oriTmpl.ReverseSchedTmplId, dbo)
			instrId, reverseInstrId, err = svc.CreateBridgxExpandInstr(ctx, oriTmpl.Id, oriTmpl.ReverseSchedTmplId, true, dbo)
			if err != nil {
				return nil, err
			}
			instrGroup = append(instrGroup, instrId)
			if reverseInstrId != 0 {
				instrReverseGroup = append(instrReverseGroup, reverseInstrId)
			}
		case ExpandStepBaseEnv:
			svc := GetInstrSvcInst()
			instrId, reverseInstrId, err = svc.CreateBaseEnvInstr(ctx, svcReq.BaseEnv, oriTmpl.Id, false, dbo)
			if err != nil {
				return nil, err
			}
			instrGroup = append(instrGroup, instrId)
			if reverseInstrId != 0 {
				instrReverseGroup = append(instrReverseGroup, reverseInstrId)
			}
		case ExpandStepServiceEnv:
			svc := GetInstrSvcInst()
			svcReq.ServiceEnv.ServiceName = serviceCluster.ServiceName
			instrId, reverseInstrId, err = svc.CreateServiceEnvInstr(ctx, svcReq.ServiceEnv, oriTmpl.Id, false, dbo)
			if err != nil {
				return nil, err
			}
			instrGroup = append(instrGroup, instrId)
			if reverseInstrId != 0 {
				instrReverseGroup = append(instrReverseGroup, reverseInstrId)
			}
		case ExpandStepMount:
			svc := GetInstrSvcInst()
			instrId, reverseInstrId, err = svc.CreateMountSlbInstr(ctx, svcReq.Mount, oriTmpl.Id, oriTmpl.ReverseSchedTmplId, true, dbo)
			if err != nil {
				return nil, err
			}
			instrGroup = append(instrGroup, instrId)
			if reverseInstrId != 0 {
				instrReverseGroup = append(instrReverseGroup, reverseInstrId)
			}
		}
		if step == svcReq.EndStep {
			break
		}
	}
	// 生成逆向 instrGroup , 更新逆向 tmpl
	log.Logger.Info("instrGroup", instrGroup)
	log.Logger.Info("instrReverseGroup", instrReverseGroup)
	instrReverseGroup = tool.ReverseIntSlice(instrReverseGroup)
	log.Logger.Info("instrReverseGroup", instrReverseGroup)
	err = s.updateTmplInstrGroup(ctx, oriTmpl.ReverseSchedTmplId, instrReverseGroup, oriTmpl.Id, dbo)
	if err != nil {
		return nil, err
	}
	err = s.updateTmplInstrGroup(ctx, oriTmpl.Id, instrGroup, oriTmpl.ReverseSchedTmplId, dbo)
	if err != nil {
		return nil, err
	}
	svcResp.TmplExpandId = oriTmpl.Id

	return &TemplateSvcResp{TmplUpdateSvcResp: svcResp}, nil
}

func (s *TemplateSvc) Delete(ctx context.Context, tmpExpandIds []int64) (int64, error) {
	var err error
	records, err := repository.GetScheduleTemplateRepoInst().Delete(ctx, tmpExpandIds)
	if err != nil {
		log.Logger.Errorf("template deleted error:%v", err)
		return 0, err
	}
	return records, nil
}
