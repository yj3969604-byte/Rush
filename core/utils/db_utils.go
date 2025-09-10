package utils

import (
	"BaseGoUni/core/pojo"
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"log"
	"os"
	"strings"
	"time"
)

var Db *gorm.DB

func InitDb() (firstInit bool, err error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			//LogLevel:      logger.Info, // Log level
			LogLevel:                  logger.Error, // Log level
			IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,         // Don't include params in the SQL log
			Colorful:                  true,         // Disable color
		},
	)
	noSchemataMaster := fmt.Sprintf(GlobalConfig.Mysql.Master, "")
	Db, err = gorm.Open(mysql.Open(noSchemataMaster), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Printf("连接数据库错误 %s;noSchemataMaster=%s", err.Error(), noSchemataMaster)
		panic(err)
		return
	}
	err = Db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", CsConfig.DefaultHost.TablePrefix)).Error
	if err != nil {
		log.Printf("创建数据库错误 %s", err.Error())
		panic(err)
		return
	}
	masterStr := fmt.Sprintf(GlobalConfig.Mysql.Master, CsConfig.DefaultHost.TablePrefix)
	slaveStr := fmt.Sprintf(GlobalConfig.Mysql.Slave, CsConfig.DefaultHost.TablePrefix)
	err = Db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{mysql.Open(masterStr)}, // 主库，写操作
		Replicas: []gorm.Dialector{mysql.Open(slaveStr)},  // 从库，读操作
		Policy:   dbresolver.RandomPolicy{},               // 读库负载均衡策略
	}))
	if err != nil {
		panic(err)
		return
	}
	sqlDB, err := Db.DB()
	firstInit, err = InitTables(CsConfig.DefaultHost.TablePrefix)
	_ = Db.AutoMigrate(&pojo.HostInfo{})
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(15 * time.Minute)
	return firstInit, nil
}

func InitTables(prefix string) (firstInit bool, err error) {
	db := NewPrefixDb(prefix)
	if db.Exec("desc "+pojo.UserTableName).Error != nil {
		firstInit = true
		err = db.AutoMigrate(
			&pojo.SysUser{},
			&pojo.SysRole{},
			&pojo.SysMenu{},
		)
		if err != nil {
			panic(err)
		}
	}
	InitShardingHook(db)
	if db.Exec("desc "+pojo.AllCashHistoryShardingName).Error != nil {
		err = InitShardingDataBase(db, pojo.CashHistory{}, pojo.CashHistoryTableName, pojo.CashHistoryShards)
		if err != nil {
			panic(fmt.Sprintf("Failed to init table: %v", err))
		}
		CreateView(uint(pojo.CashHistoryShards), pojo.AllCashHistoryShardingName, pojo.CashHistoryTableName)
		log.Printf("Init cash history success...\n")
	}
	return firstInit, nil
}

var dbPool = make(map[string]*gorm.DB)

func NewPrefixDb(prefix string) (db *gorm.DB) {
	if existingDb, ok := dbPool[prefix]; ok {
		return existingDb
	}
	db = Db.Session(&gorm.Session{
		NewDB: true,
	})
	err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", prefix)).Error
	if err != nil {
		log.Printf("创建数据库错误 %s", err.Error())
		return nil
	}
	masterStr := fmt.Sprintf(GlobalConfig.Mysql.Master, prefix)
	slaveStr := fmt.Sprintf(GlobalConfig.Mysql.Slave, prefix)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			//LogLevel:      logger.Info, // Log level
			LogLevel:                  logger.Error, // Log level
			IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,         // Don't include params in the SQL log
			Colorful:                  true,         // Disable color
		},
	)
	newDb, err := gorm.Open(mysql.Open(masterStr), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Printf("连接数据库错误 %s", err.Error())
		return nil
	}
	err = newDb.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{mysql.Open(masterStr)}, // 主库，写操作
		Replicas: []gorm.Dialector{mysql.Open(slaveStr)},  // 从库，读操作
		Policy:   dbresolver.RandomPolicy{},               // 读库负载均衡策略
	}))
	if err != nil {
		panic(err)
		return
	}
	sqlDB, err := newDb.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(50)                  // 设置最大连接数
	sqlDB.SetMaxIdleConns(20)                  // 设置最大空闲连接数
	sqlDB.SetConnMaxLifetime(15 * time.Minute) // 设置连接最大生命周期
	ctx := context.WithValue(context.Background(), KeyDbPrefix, prefix)
	newDb = newDb.WithContext(ctx)
	dbPool[prefix] = newDb
	return newDb
}

func GetDbPrefix(db *gorm.DB) (prefix string) {
	prefixObj := db.Statement.Context.Value(KeyDbPrefix)
	if prefixObj != nil && !strings.HasPrefix(db.Statement.Table, prefixObj.(string)) {
		prefix = prefixObj.(string)
	}
	return prefix
}
