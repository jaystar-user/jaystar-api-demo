package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"jaystar/internal/config"
	"log"
	"os"
	"time"
)

const (
	BuildDSNFormat     = "host=%s port=%s user=%s password=%s dbname=%s"
	setMaxIdleConns    = 25
	setMaxOpenConns    = 50
	setConnMaxIdleTime = 30 * time.Second
	setConnMaxLifeTime = 5 * time.Minute
)

type IPostgresDB interface {
	Session() *gorm.DB
}

func ProvidePostgresDB(config config.IConfigEnv) IPostgresDB {
	newDB := &postgresDB{
		config: config,
	}
	newDB.DB = newDB.postgresDBConnect()

	return newDB
}

type postgresDB struct {
	DB     *gorm.DB
	config config.IConfigEnv
}

func (p *postgresDB) Session() *gorm.DB {
	return p.DB.Session(&gorm.Session{SkipDefaultTransaction: true})
}

func (p *postgresDB) postgresDBConnect() *gorm.DB {
	cfgLogLevel := p.config.GetDbConfig().LogLevel
	var logLevel logger.LogLevel
	switch cfgLogLevel {
	case "debug":
		logLevel = logger.Info
	case "warn":
		logLevel = logger.Warn
	default:
		logLevel = logger.Error
	}

	customLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logLevel,    // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
		},
	)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DriverName: "pgx",
		DSN: fmt.Sprintf(BuildDSNFormat,
			p.config.GetDbConfig().Host,
			p.config.GetDbConfig().Port,
			p.config.GetDbConfig().User,
			p.config.GetDbConfig().Password,
			p.config.GetDbConfig().DBName,
		),
	}), &gorm.Config{
		Logger: customLogger,
	})

	db.Use(
		dbresolver.Register(dbresolver.Config{}).
			// 空閒連線 timeout 時間
			SetConnMaxIdleTime(setConnMaxIdleTime).
			// 空閒連線可重複使用的時間長度
			SetConnMaxLifetime(setConnMaxLifeTime).
			// 限制最多保留閒置連線數
			SetMaxIdleConns(setMaxIdleConns).
			// 限制最大開啟的連線數
			SetMaxOpenConns(setMaxOpenConns),
	)

	if err != nil {
		panic(fmt.Sprintf("open postgresql error and error message is '%v'", err))
	}

	return db
}
