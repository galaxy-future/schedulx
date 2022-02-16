package db

import (
	"github.com/galaxy-future/schedulx/register/config/client"
	"github.com/galaxy-future/schedulx/register/config/log"
	"gorm.io/gorm"
)

const (
	BatchSize = 5
)

// Create insert the value into database
func Create(model interface{}, db *gorm.DB) error {
	if db == nil {
		db = client.WriteDBCli
	}
	err := db.Debug().Create(model).Error
	if err != nil {
		emitDbErrorMetrics("create data to write db", err)
		return err
	}
	return nil
}

// Delete delete
func Delete(model interface{}, db *gorm.DB) error {
	if db == nil {
		db = client.WriteDBCli
	}
	err := db.Debug().Where(model).Delete(model).Error
	if err != nil {
		emitDbErrorMetrics("delete data", err)
		return err
	}
	return nil
}

func BatchCreate(values interface{}, db *gorm.DB) error {
	if db == nil {
		db = client.WriteDBCli
	}
	err := db.CreateInBatches(values, BatchSize).Error
	if err != nil {
		emitDbErrorMetrics("create data to write db", err)
		return err
	}
	return nil
}

// Save update value in database, if the value doesn't have primary key, will insert it
func Save(model interface{}, db *gorm.DB) error {
	if db == nil {
		db = client.WriteDBCli
	}
	err := db.Save(model).Error
	if err != nil {
		emitDbErrorMetrics("save data to write db", err)
		return err
	}
	return nil
}

// Query records
func Query(where map[string]interface{}, page int, pageSize int, models interface{}, order string, fields []string, withCount bool) (count int64, err error) {
	cli := client.ReadDBCli
	if len(fields) != 0 {
		cli = cli.Select(fields)
	}
	offset := (page - 1) * pageSize
	if err = cli.Debug().Where(where).Order(order).Offset(offset).Limit(pageSize).Find(models).Error; err != nil {
		emitDbErrorMetrics("query data from read db", err)
		return 0, err
	}
	if withCount {
		if err = cli.Where(where).Order(order).Offset(offset).Limit(pageSize).Find(models).Offset(-1).Limit(-1).Count(&count).Error; err != nil {
			emitDbErrorMetrics("query data from read db", err)
			return 0, err
		}
		return count, nil
	}
	return 0, nil
}

// QueryAll records
func QueryAll(where map[string]interface{}, models interface{}, order string, fields []string) (err error) {
	cli := client.ReadDBCli
	if len(fields) != 0 {
		cli = cli.Select(fields)
	}
	if err = cli.Debug().Where(where).Order(order).Find(models).Error; err != nil {
		emitDbErrorMetrics("query all from read db", err)
		return err
	}
	return nil
}

func QueryFirst(where map[string]interface{}, model interface{}) (err error) {
	if err = client.ReadDBCli.Debug().Where(where).First(model).Error; err != nil {
		emitDbErrorMetrics("query first from read db", err)
		return err
	}
	return nil
}

func QueryLimit(where map[string]interface{}, models interface{}, order string, fields []string, limitCnt int) (err error) {
	cli := client.ReadDBCli
	if len(fields) != 0 {
		cli = cli.Select(fields)
	}
	if err = cli.Where(where).Order(order).Limit(limitCnt).Find(models).Error; err != nil {
		emitDbErrorMetrics("query all from read db", err)
		return err
	}
	return nil
}

func QueryLast(where map[string]interface{}, model interface{}) (err error) {
	if err = client.ReadDBCli.Where(where).Last(model).Error; err != nil {
		emitDbErrorMetrics("query last from read db", err)
		return err
	}
	return nil
}

// UpdatesByIds update attributes with callbacks
func UpdatesByIds(model interface{}, ids []int64, updates map[string]interface{}, db *gorm.DB) error {
	if db == nil {
		db = client.WriteDBCli
	}
	if err := db.Model(model).Where("id IN (?)", ids).Updates(updates).Error; err != nil {
		emitDbErrorMetrics("update data list to write db", err)
		return err
	}
	return nil
}

func Updates(model interface{}, where map[string]interface{}, updates map[string]interface{}, db *gorm.DB) (int64, error) {
	if db == nil {
		db = client.WriteDBCli
	}
	r := db.Debug().Model(model).Where(where).Updates(updates)
	if err := r.Error; err != nil {
		emitDbErrorMetrics("update data list to write db", err)
		return 0, err
	}
	return r.RowsAffected, nil
}

// Get find first record that match given conditions, order by primary key
func Get(id int64, out interface{}) error {
	if err := client.ReadDBCli.Debug().Where("id = ?", id).First(out).Error; err != nil {
		emitDbErrorMetrics("get data from read db", err)
		return err
	}
	return nil
}

// Gets find records that match given conditions
func Gets(ids []int64, out interface{}) error {
	if err := client.ReadDBCli.Where(ids).Find(out).Error; err != nil {
		emitDbErrorMetrics("get data list from read db", err)
		return err
	}
	return nil
}

func emitDbErrorMetrics(errType string, err error) {
	log.Logger.Error(errType, err)
}
