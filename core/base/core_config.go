package base

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Salt   string `yaml:"salt"`
	PriKey string `yaml:"priKey"`
	Mysql  struct {
		DataSource string `yaml:"dataSource"`
		Master     string `yaml:"master"`
		Slave      string `yaml:"slave"`
	} `yaml:"mysql"`
	Redis    Redis        `yaml:"redis"`
	RabbitMq RabbitMq     `yaml:"rabbitMq"`
	AliOss   AliOssConfig `yaml:"aliOss"`
}

type AliOssConfig struct {
	BaseDownloadFolder string `yaml:"baseDownloadFolder"`
	DownloadLink       string `yaml:"downloadLink"`
	Bucket             string `yaml:"bucket"`
	Language           string `yaml:"language"`
	Endpoint           string `yaml:"endpoint"`
	AccessKeyId        string `yaml:"accessKeyId"`
	AccessKeySecret    string `yaml:"accessKeySecret"`
}

type Redis struct {
	Host string `yaml:"host"`
	Pass string `yaml:"pass"`
	Db   int    `yaml:"db"`
}

type RabbitMq struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Exchange string `yaml:"exchange"`
	Routing  string `yaml:"routing"`
	Queue    string `yaml:"queue"`
}

func InitGlobalConfig(file string, GlobalConfig *Config) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	//log.Print("load config file.data=$data", string(data))
	return yaml.Unmarshal(data, &GlobalConfig)
}
