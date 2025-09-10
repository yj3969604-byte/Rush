package pojo

import (
	"time"
)

type HostInfo struct {
	BaseModel
	HostName     string `yaml:"hostName" json:"hostName" gorm:"type:varchar(128);uniqueIndex"` // 域名
	TablePrefix  string `yaml:"tablePrefix" json:"tablePrefix" gorm:"type:varchar(8)"`         // 表名前缀
	HostMark     string `yaml:"hostMark" json:"hostMark" gorm:"type:varchar(64)"`              // 域名备注
	HostDesc     string `yaml:"hostDesc" json:"hostDesc" gorm:"type:text"`                     // 域名更多
	AccessSecret string `yaml:"accessSecret" json:"accessSecret" gorm:"type:text"`             // 密钥
	Salt         string `yaml:"salt" json:"salt" gorm:"type:varchar(32)"`                      // 密钥盐
	PriKey       string `yaml:"priKey" json:"priKey" gorm:"type:text"`                         // 私钥
	AccessExpire int64  `yaml:"accessExpire" json:"accessExpire" gorm:"type:bigint"`           // token到期(秒)
	Enabled      bool   `yaml:"enabled" json:"enabled"`                                        // 是否启用
}

type HostInfoSearch struct {
	PageInfo
	HostName string `yaml:"hostName" json:"hostName"` // 域名
	HostMark string `yaml:"hostMark" json:"hostMark"` // 域名备注
	Enabled  bool   `yaml:"enabled" json:"enabled"`   // 是否可用
}

type HostInfoSet struct {
	ID           int64  `json:"id"`                                    // 自动生成的ID
	HostName     string `json:"hostName" binding:"required,max=128"`   // 域名，必填，最长128个字符
	TablePrefix  string `json:"tablePrefix" binding:"required,max=8"`  // 表前缀，必填，最长8个字符
	HostMark     string `json:"hostMark" binding:"omitempty,max=64"`   // 域名备注，可选，最长64个字符
	HostDesc     string `json:"hostDesc" binding:"omitempty,max=1024"` // 域名描述，可选，最长1024个字符
	Enabled      bool   `json:"enabled" binding:"required"`            // 是否启用，必填，布尔值
	AccessExpire int64  `json:"accessExpire" binding:"required,min=1"` // Token过期时间，必填，必须大于0
}

type HostInfoBack struct {
	ID           int64     `json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	HostName     string    `yaml:"hostName" json:"hostName"`         // 域名
	TablePrefix  string    `yaml:"tablePrefix" json:"tablePrefix"`   // 表名前缀
	HostMark     string    `yaml:"hostMark" json:"hostMark"`         // 域名备注
	HostDesc     string    `yaml:"hostDesc" json:"hostDesc"`         // 域名更多
	AccessExpire int64     `yaml:"accessExpire" json:"accessExpire"` // token到期(秒)
	Enabled      bool      `yaml:"enabled" json:"enabled"`           // 是否可用
}

type HostInfoResp struct {
	BasePageResponse[HostInfoBack]
}

var HostInfoTableName = "host_info"

func (HostInfo) TableName() string {
	return HostInfoTableName
}
