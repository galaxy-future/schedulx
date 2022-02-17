package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/pkg/tool"
	"github.com/galaxy-future/schedulx/register/config"
	"github.com/galaxy-future/schedulx/register/config/client"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/repository/model/db"
	jsoniter "github.com/json-iterator/go"
)

type SubTaskRepo struct {
}

var subTaskRepoInst *SubTaskRepo
var subTaskRepoOnce sync.Once

func GetSubTaskRepoInst() *SubTaskRepo {
	subTaskRepoOnce.Do(func() {
		subTaskRepoInst = &SubTaskRepo{}
	})
	return subTaskRepoInst
}

func (r *SubTaskRepo) GetLastExpandSuccTask(ctx context.Context, tmplId int64) (*db.Task, error) {
	var err error
	where := map[string]interface{}{
		"sched_tmpl_id": tmplId,
		"task_status":   types.TaskStatusSuccess,
	}
	log.Logger.Infof("GetLastExpandSuccTask | %+v", where)
	obj := &db.Task{}
	err = db.QueryLast(where, obj)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	return obj, nil
}

func (r *SubTaskRepo) CountByCond(ctx context.Context, schedTmplIds []int64, status []string) (int64, error) {
	where := map[string]interface{}{
		"sched_tmpl_id": schedTmplIds,
		"task_status":   status,
	}
	var cnt int64
	if err := client.ReadDBCli.Where(where).Model(&db.Task{}).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}

func (r *SubTaskRepo) CreateSubTask(ipList []*types.InstanceInfo, stepLen int, schedTaskId int64, taskInfo string, isRollback bool) ([]*db.SubTask, error) {
	var err error
	taskStatus := types.TaskStatusRunning
	if isRollback {
		taskStatus = types.TaskStatusRollingBack
	}

	total := len(ipList)
	if stepLen < 1 {
		stepLen = 1
	}

	batchNum := total / stepLen
	if total%stepLen != 0 {
		batchNum++
	}

	for i := 0; i < batchNum; i++ {
		start, end := i*stepLen, i*stepLen+stepLen
		if end > total {
			end = total
		}

		part := ipList[start:end]
		ids := make([]int, len(part))
		for i, ns := range part {
			ids[i] = ns.Node.Id
		}

		idsBytes, _ := json.Marshal(ids)
		idsStr := string(idsBytes)

		batch := &models.FlowBatch{
			Flow:        instance,
			Status:      models.STATUS_INIT,
			Step:        -1, // not started
			Nodes:       idsStr,
			CreatedTime: time.Now(),
			UpdatedTime: time.Now(),
		}

		err := flowService.InsertBase(batch)
		if err != nil {
			return err
		}

		// update node states
		corrId := fmt.Sprintf("%d-%d", instance.Id, batch.Id)
		for _, ns := range part {
			ns.CorrId = corrId
			ns.Batch = batch

			flowService.UpdateBase(ns)
		}
	}

	newTask := &db.SubTask{
		//SuperTaskId: schedTmplId,
		//Operator:    operator,
		//ExecType:    execType,
		//InstCnt:     instCnt,
		TaskStatus: taskStatus,
		TaskStep:   types.TaskStepInit,
		BeginAt:    time.Now(),
	}
	if taskInfo != "" {
		newTask.TaskInfo = taskInfo
	}
	if err = db.Create(newTask, nil); err != nil {
		log.Logger.Error(err)
		return 0, err
	}
	return newTask.Id, err
}

func (r *SubTaskRepo) UpdateTaskRelationTaskId(ctx context.Context, taskId int64, field string, relationTaskId int64) error {
	var err error
	obj := &db.Task{}
	err = db.Get(taskId, obj)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	rtd := &types.RelationTaskId{}
	if obj.RelationTaskId != "" {
		err = jsoniter.Unmarshal([]byte(obj.RelationTaskId), &rtd)
		if err != nil {
			log.Logger.Error(err)
			return err
		}
	}
	switch field {
	case types.BridgXTaskId:
		rtd.BridgXTaskId = relationTaskId
	case types.NodeactTaskId:
		rtd.NodeActTaskId = relationTaskId
	}
	value, _ := jsoniter.Marshal(rtd)
	data := map[string]interface{}{
		"relation_task_id": value,
	}
	where := map[string]interface{}{
		"id": taskId,
	}
	rowEffected, err := db.Updates(&db.Task{}, where, data, nil)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	if rowEffected != 1 {
		err = config.ErrRowsAffectedInvalid
		log.Logger.Error(err)
		return err
	}
	return err
}

func (r *SubTaskRepo) UpdateTaskStatus(ctx context.Context, taskId int64, taskStatus, msg string) error {
	var err error
	data := map[string]interface{}{
		"task_status": taskStatus,
		"finish_at":   time.Now(),
		"msg":         tool.SubStr(msg, 100),
	}
	where := map[string]interface{}{
		"id": taskId,
	}
	rowEffected, err := db.Updates(&db.Task{}, where, data, nil)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	if rowEffected != 1 {
		err = config.ErrRowsAffectedInvalid
		log.Logger.Error(err)
		return err
	}
	return err
}

func (r *SubTaskRepo) UpdateTaskStep(ctx context.Context, taskId int64, taskStep, msg string) error {
	var err error
	data := map[string]interface{}{
		"task_step": taskStep,
		"finish_at": time.Now(),
		"msg":       tool.SubStr(msg, 100),
	}
	where := map[string]interface{}{
		"id": taskId,
	}
	rowEffected, err := db.Updates(&db.Task{}, where, data, nil)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	if rowEffected != 1 {
		err = config.ErrRowsAffectedInvalid
		log.Logger.Error(err)
		return err
	}
	return err
}

func (r *SubTaskRepo) GetTask(ctx context.Context, taskId int64) (*db.Task, error) {
	var err error
	obj := &db.Task{}
	err = db.Get(taskId, obj)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}

	return obj, err
}

func (r *SubTaskRepo) GetBridgXTaskId(ctx context.Context, taskId int64) (int64, error) {
	var err error
	task, err := r.GetTask(ctx, taskId)
	if err != nil {
		log.Logger.Error(err)
		return 0, err
	}
	relationIds := &types.RelationTaskId{}
	err = jsoniter.Unmarshal([]byte(task.RelationTaskId), relationIds)
	if err != nil {
		log.Logger.Error(err)
		return 0, err
	}
	return relationIds.BridgXTaskId, nil
}
