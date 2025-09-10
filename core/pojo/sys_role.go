package pojo

import "time"

type SysRole struct {
	BaseModel
	Code        string   `yaml:"code" json:"code" gorm:"type:varchar(32);uniqueIndex"` // 角色编码
	Name        string   `yaml:"name" json:"name" gorm:"type:varchar(32)"`             // 角色名称
	Description string   `yaml:"description" json:"description" gorm:"type:text"`      // 描述
	MenuIdStr   string   `yaml:"menuIdStr" json:"menuIdStr" gorm:"type:text"`          // 菜单集id，JSON 字符串
	MenuNames   []string `yaml:"menuNames" json:"menuNames" gorm:"-"`                  // 菜单集，程序使用，不映射数据库
}

type RoleSet struct {
	ID          int64   `yaml:"id" json:"id"`
	Code        string  `yaml:"code" json:"code" binding:"required,max=32"`                  // 角色编码，必填，最大32字符
	Name        string  `yaml:"name" json:"name" binding:"required,max=32"`                  // 角色名称，必填，最大32字符
	Description string  `yaml:"description" json:"description" binding:"omitempty,max=1024"` // 描述，选填，最大1024字符
	MenuIdStr   string  `yaml:"menuIdStr" json:"menuIdStr" binding:"required"`               // 菜单集id，必填
	MenuIds     []int64 `yaml:"menuIds" json:"menuIds" binding:"required,dive,gt=0"`         // 菜单集，必填，数组中的每个元素都必须大于0
}

type RoleSearch struct {
	PageInfo
	Code string `yaml:"code" json:"code"` // 角色编码
	Name string `yaml:"name" json:"name"` // 角色名称
}

type RoleBack struct {
	ID          int64     `yaml:"id" json:"id"`
	Code        string    `yaml:"code" json:"code"`               // 角色编码
	Name        string    `yaml:"name" json:"name"`               // 角色名称
	Description string    `yaml:"description" json:"description"` // 描述
	MenuIds     []int64   `yaml:"menuIds" json:"menuIds"`         // 菜单集
	CreatedAt   time.Time `yaml:"createdAt" json:"createdAt"`
}

type RoleResp struct {
	BasePageResponse[RoleBack]
}

var RoleTableName = "sys_role"

func (SysRole) TableName() string {
	return RoleTableName
}
