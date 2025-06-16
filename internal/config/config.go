package config

import (
	"fmt"
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/spf13/viper"
	"jaystar/internal/utils/osUtil"
	"log"
	"path"
	"runtime"
)

const (
	configPath = "/configs/%s"
)

type IConfigEnv interface {
	GetAppEnv() string
	GetLogConfig() logger.LogConfig
	GetDbConfig() DbConfig
	GetHttpConfig() httpConfig
	GetKintoneConfig() kintoneConfig
}

func ProviderIConfigEnv() IConfigEnv {
	var cfg configEnv
	appEnv := osUtil.GetOsEnv()
	cfg.AppEnv = appEnv

	wd := GetExactRoot(2)
	log.Printf("WORKING_DIRECTORY: %s, APP_ENV : %s, PATH: %s", wd, appEnv, fmt.Sprintf(configPath, appEnv))

	viper.AddConfigPath(wd + fmt.Sprintf(configPath, appEnv))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("🔔🔔🔔 fatal error viper.ReadInConfig: %v 🔔🔔🔔", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("🔔🔔🔔 fatal error viper.Unmarshal: %v 🔔🔔🔔", err)
	}

	return &cfg
}

// fields need to be public for viper to set values in
type configEnv struct {
	AppEnv        string
	HttpConfig    httpConfig    `mapstructure:"http"`
	LogConfig     logConfig     `mapstructure:"log"`
	DbConfig      DbConfig      `mapstructure:"postgres"`
	KintoneConfig kintoneConfig `mapstructure:"kintone"`
}

type httpConfig struct {
	BaseUrl string `mapstructure:"baseUrl"`
	Port    string `mapstructure:"port"`
}

type logConfig struct {
	Name     string `mapstructure:"name"`
	Encoding string `mapstructure:"encoding"`
	Level    string `mapstructure:"level"`
}

type DbConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	DBName   string `mapstructure:"dbName"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	LogLevel string `mapstructure:"log_level"`
}

type kintoneConfig struct {
	Url                     string    `mapstructure:"url"`
	CommonUserAuthorization string    `mapstructure:"common_user_authorization"`
	AdminUserAuthorization  string    `mapstructure:"admin_user_authorization"`
	AppId                   AppIdInfo `mapstructure:"app_id"`
}

type AppIdInfo struct {
	StudentInfo          string `mapstructure:"student_info"`
	PointCard            string `mapstructure:"point_card"`
	ScheduleRecord       string `mapstructure:"schedule_record"`
	DepositRecord        string `mapstructure:"deposit_record"`
	ReduceRecord         string `mapstructure:"reduce_record"`
	SemesterSettleRecord string `mapstructure:"semester_settle_record"`
}

func (c *configEnv) GetAppEnv() string {
	return c.AppEnv
}

func (c *configEnv) GetLogConfig() logger.LogConfig {
	return logger.LogConfig(c.LogConfig)
}

func (c *configEnv) GetDbConfig() DbConfig {
	return c.DbConfig
}

func (c *configEnv) GetHttpConfig() httpConfig {
	return c.HttpConfig
}

func (c *configEnv) GetKintoneConfig() kintoneConfig {
	return c.KintoneConfig
}

func GetExactRoot(pathDepRelFromRoot int) string {
	_, filename, _, _ := runtime.Caller(0)
	curFileDir := path.Dir(filename)

	for i := 0; i < pathDepRelFromRoot; i++ {
		curFileDir = path.Dir(curFileDir)
	}

	return curFileDir
}
