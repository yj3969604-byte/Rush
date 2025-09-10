package pojo

import (
	"time"
)

type SysConfig struct { // 任务记录
	BaseModel
	ConfigKey   string `yaml:"configKey" json:"configKey" gorm:"type:varchar(64);uniqueIndex"` // 配置键
	ConfigValue string `yaml:"configValue" json:"configValue" gorm:"type:text"`                // 配置值
	ConfigDesc  string `yaml:"configDesc" json:"configDesc" gorm:"type:text"`                  // 配置描述
}

type SysConfigBack struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ConfigKey   string    `json:"configKey"`
	ConfigValue string    `json:"configValue"`
	ConfigDesc  string    `json:"configDesc"`
}

type SysConfigResp struct {
	BasePageResponse[SysConfigBack]
}

var SysConfigTableName = "sys_config"

func (SysConfig) TableName() string {
	return SysConfigTableName
}
