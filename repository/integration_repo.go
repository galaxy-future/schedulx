package repository

import (
	"context"
	"sync"
	"time"

	"github.com/galaxy-future/schedulx/register/constant"
	"github.com/spf13/cast"

	"github.com/galaxy-future/schedulx/register/config/client"

	"github.com/galaxy-future/schedulx/api/types"
	"github.com/galaxy-future/schedulx/pkg/tool"
	"github.com/galaxy-future/schedulx/repository/model/db"
	"gorm.io/gorm"
)

type IntegrationRepo struct {
}

func (r IntegrationRepo) GetOneByType(ctx context.Context, _type string) (*types.Integration, error) {
	ret := &db.Integration{}
	err := db.QueryLast(map[string]interface{}{
		"type": _type,
	}, ret)
	if err != nil {
		return nil, err
	}
	if ret == nil || ret.Id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	decryptPassword, err := tool.AesDecrypt(ret.Password, []byte(ret.Account))
	if err != nil {
		return nil, err
	}
	return &types.Integration{
		Host:     ret.Host,
		Account:  ret.Account,
		Password: string(decryptPassword),
		Type:     _type,
		CreateAt: ret.CreateAt,
		CreateBy: ret.CreateBy,
	}, nil
}

func (r *IntegrationRepo) Create(ctx context.Context, host, account, password, _type string) error {
	now := time.Now()
	operator := cast.ToString(ctx.Value(constant.CtxUserNameKey))
	encryptPassword, err := tool.AesEncrypt([]byte(password), []byte(account))
	if err != nil {
		return err
	}
	integration := &db.Integration{
		Host:     host,
		Account:  account,
		Password: encryptPassword,
		Type:     _type,
		CreateAt: &now,
		UpdateAt: &now,
		CreateBy: operator,
		UpdateBy: operator,
	}
	return db.Create(integration, nil)
}

func (r *IntegrationRepo) Delete(ctx context.Context, ids []int64) error {
	return client.WriteDBCli.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			m := &db.Integration{Id: id}
			err := tx.Delete(m).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *IntegrationRepo) List(ctx context.Context, _type string, page int, pageSize int) ([]db.Integration, int, error) {
	res := make([]db.Integration, 0)
	total, err := db.Query(map[string]interface{}{
		"type": _type,
	}, page, pageSize, &res, "", nil, true)
	if err != nil {
		return nil, 0, err
	}
	return res, int(total), nil
}

var integrationInstance *IntegrationRepo
var integrationRepoOnce sync.Once

func GetIntegrationInstance() *IntegrationRepo {
	integrationRepoOnce.Do(func() {
		integrationInstance = &IntegrationRepo{}
	})
	return integrationInstance
}
