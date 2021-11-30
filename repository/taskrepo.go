package repository

import (
	"context"
	"sync"
	"time"

	"github.com/galaxy-future/schedulx/pkg/tool"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/register/config"
	"github.com/galaxy-future/schedulx/register/config/log"
	"github.com/galaxy-future/schedulx/repository/model/db"
	jsoniter "github.com/json-iterator/go"
)

type TaskRepo struct {
}

var taskRepoInst *TaskRepo
var taskRepoOnce sync.Once

func GetTaskRepoInst() *TaskRepo {
	taskRepoOnce.Do(func() {
		taskRepoInst = &TaskRepo{}
	})
	return taskRepoInst
}

func (r *TaskRepo) GetLastExpandSuccTask(ctx context.Context, tmplId int64) (*db.Task, error) {
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

func (r *TaskRepo) CreateTask(ctx context.Context, schedTmplId, instCnt int64, operator, execType string) (int64, error) {
	var err error
	newTask := &db.Task{
		SchedTmplId: schedTmplId,
		Operator:    operator,
		ExecType:    execType,
		InstCnt:     instCnt,
		TaskStatus:  types.TaskStatusRunning,
		TaskStep:    types.TaskStepInit,
		BeginAt:     time.Now(),
	}
	if err = db.Create(newTask, nil); err != nil {
		log.Logger.Error(err)
		return 0, err
	}
	return newTask.Id, err
}

func (r *TaskRepo) UpdateTaskRelationTaskId(ctx context.Context, taskId int64, field string, relationTaskId int64) error {
	var err error
	obj := &db.Task{}
	err = db.Get(taskId, obj)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	rTaskIdMap := make(map[string]interface{})
	if obj.RelationTaskId != "" {
		err = jsoniter.Unmarshal([]byte(obj.RelationTaskId), &rTaskIdMap)
		if err != nil {
			log.Logger.Error(err)
			return err
		}
	}
	rTaskIdMap[field] = relationTaskId
	value, _ := jsoniter.Marshal(rTaskIdMap)
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

func (r *TaskRepo) UpdateTaskStatus(ctx context.Context, taskId int64, taskStatus, msg string) error {
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

func (r *TaskRepo) UpdateTaskStep(ctx context.Context, taskId int64, taskStep, msg string) error {
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

func (r *TaskRepo) GetTask(ctx context.Context, taskId int64) (*db.Task, error) {
	var err error
	obj := &db.Task{}
	err = db.Get(taskId, obj)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}

	return obj, err
}
