package client

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm/logger"

	"github.com/galaxy-future/schedulx/register/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var WriteDBCli *gorm.DB
var ReadDBCli *gorm.DB

func InitDBClients() {
	var err error
	WriteDBCli, err = GetSqlDriver(config.GlobalConfig.WriteDB)
	if err != nil {
		panic(err)
	}
	ReadDBCli, err = GetSqlDriver(config.GlobalConfig.ReadDB)
	if err != nil {
		panic(err)
	}
	if config.GlobalConfig.DebugMode {
		WriteDBCli = WriteDBCli.Debug()
		ReadDBCli = ReadDBCli.Debug()
	}
}

func GetSqlDriver(dbConf config.DBConfig) (*gorm.DB, error) {
	var err error
	var dbDialector = getDbDialector(dbConf)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
	)
	gormDb, err := gorm.Open(dbDialector, &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 newLogger,
	})
	if err != nil {
		return nil, err
	}

	rawDb, err := gormDb.DB()
	if err != nil {
		return nil, err
	}

	rawDb.SetConnMaxIdleTime(time.Second * 30)
	rawDb.SetMaxIdleConns(dbConf.MaxIdleConns)
	rawDb.SetMaxOpenConns(dbConf.MaxOpenConns)
	return gormDb, nil
}

func getDbDialector(conf config.DBConfig) gorm.Dialector {
	var dbDialector gorm.Dialector
	dsn := getDsn(conf)
	dbDialector = mysql.Open(dsn)
	return dbDialector
}

func getDsn(dbConf config.DBConfig) string {
	Host := dbConf.Host
	DataBase := dbConf.Name
	Port := dbConf.Port
	User := dbConf.User
	Pass := dbConf.Password
	Charset := "utf8mb4"
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local", User, Pass, Host, Port, DataBase, Charset)
}
