package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"
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

func (r *SubTaskRepo) GetLastSuccessSubTask(ctx context.Context, tmplId int64) (*db.Task, error) {
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

func (r *SubTaskRepo) CreateSubTask(instanceList []*types.InstanceInfo, maxSurge string, schedTaskId int64) ([]*db.SubTask, error) {
	var err error
	total := uint64(len(instanceList))
	split := strings.Split(maxSurge, ",")
	subTaskList := make([]*db.SubTask, len(split))
	stepLens := make([]int, len(split))
	stepLens[0] = 0
	for i, s := range split {
		surge, err := strconv.Atoi(s)
		//max surge check
		if surge <= 0 || surge > 100 {
			log.Logger.Error(err.Error())
			err = errors.New("max surge error")
			return nil, err
		}
		stepLen := int(float64(total) * float64(surge) / 100.0)
		stepLens = append(stepLens, stepLens[i]+stepLen)
		start, end := stepLens[i], stepLens[i]+stepLen
		if end > int(total) {
			end = int(total)
		}

		instanceList, _ := jsoniter.MarshalToString(instanceList[start:end])
		newSubTask := &db.SubTask{
			SuperTaskId:  schedTaskId,
			TaskStatus:   types.TaskStatusRunning,
			TaskStep:     types.TaskStepInit,
			MaxSurge:     surge,
			InstanceList: instanceList,
			TaskInfo:     strconv.Itoa(i),
			BeginAt:      time.Now(),
		}
		subTaskList = append(subTaskList, newSubTask)
	}

	if err = db.BatchCreate(subTaskList, nil); err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	return subTaskList, err
}

func (r *SubTaskRepo) UpdateSubTaskRelationTaskId(ctx context.Context, taskId int64, field string, relationTaskId int64) error {
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

func (r *SubTaskRepo) UpdateSubTaskStatus(ctx context.Context, taskId int64, taskStatus, msg string) error {
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

func (r *SubTaskRepo) UpdateSubTaskStep(ctx context.Context, taskId int64, taskStep, msg string) error {
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

func (r *SubTaskRepo) GetSubTask(ctx context.Context, taskId int64) (*db.Task, error) {
	var err error
	obj := &db.Task{}
	err = db.Get(taskId, obj)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}

	return obj, err
}

func (r *SubTaskRepo) GetBridgXSubTaskId(ctx context.Context, taskId int64) (int64, error) {
	var err error
	task, err := r.GetSubTask(ctx, taskId)
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
