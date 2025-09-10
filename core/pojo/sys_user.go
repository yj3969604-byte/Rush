package pojo

import (
	"time"
)

type SysUser struct {
	BaseModel
	Username    string   `yaml:"username" json:"username" gorm:"type:varchar(64);uniqueIndex;"` // 账号(手机号)
	UniKey      string   `yaml:"uniKey" json:"uniKey" gorm:"type:varchar(4);uniqueIndex;"`      // 用户唯一码
	SecurityKey string   `yaml:"securityKey" json:"securityKey" gorm:"type:varchar(32);"`       // 密钥
	UserType    int      `yaml:"userType" json:"userType" gorm:"type:int;index;"`               // 用户类型 0 无效用户/1 管理员/2 后台管理/3 商户
	Password    string   `yaml:"password" json:"password" gorm:"type:varchar(64);"`             // 密码
	Enabled     bool     `yaml:"enabled" json:"enabled"`                                        // 是否启用
	Ip          string   `yaml:"ip" json:"ip" gorm:"type:varchar(64);"`                         // 最近登录ip
	Amount      float64  `yaml:"amount" json:"amount" gorm:"type:numeric(20,3);"`               // 可提现额度
	TopAmount   float64  `yaml:"topAmount" json:"topAmount" gorm:"type:numeric(20,3);"`         // 累计可提现额度
	GoogleCode  string   `yaml:"google_code" json:"google_code" gorm:"type:varchar(64);"`       // 验证器
	BindCode    bool     `yaml:"bindCode" json:"bindCode"`                                      // 是否绑定验证器
	Mark        string   `yaml:"mark" json:"mark" gorm:"type:text;"`                            // 备注
	RoleStr     string   `yaml:"roleStr" json:"roleStr" gorm:"type:text;"`                      // 角色 (JSON 字符串)
	Roles       []string `yaml:"roles" json:"roles" gorm:"-"`                                   // 角色 (不映射到数据库)
}

type AdminAwardInfo struct {
	UserId   int64   `json:"userId"`   // 操作的用户id
	Amount   float64 `json:"amount"`   // 操作金额（加钱为正数/扣钱为负数）
	CashMark string  `json:"cashMark"` // 加钱/扣钱 备注
}

type AwardInfo struct {
	CheckKey   string     `json:"checkKey"`
	CheckValue string     `json:"checkValue"`
	AwardUnis  []AwardUni `json:"awardUnis"`
}

type AwardUni struct {
	UserId     int64   `json:"userId"`
	Amount     float64 `json:"amount"`
	AwardUni   string  `json:"awardUni"`
	CashMark   string  `json:"cashMark"`
	CashDesc   string  `json:"cashDesc"`
	RefuseCash bool    `json:"refuseCash"`
	FromUserId int64   `json:"fromUserId"`
}

type UserSearch struct {
	PageInfo
	Username string `json:"username"` // 用户名
	Enabled  *bool  `json:"enabled"`  // 用户是否可用
}

type UserAdd struct {
	ID       int64    `json:"id" binding:"required,gt=0"`                     // ID，必填，大于0
	Username string   `json:"username" binding:"required,min=10,max=64"`      // 用户名，必填，长度10-64字符
	NickName string   `json:"nickName" binding:"omitempty,max=64"`            // 昵称，选填，最大64字符
	Password string   `json:"password" binding:"required,min=8,max=64"`       // 密码，必填，长度8-64字符
	Gender   int      `json:"gender" binding:"required,oneof=0 1"`            // 性别，必填，0或1
	Enabled  bool     `json:"enabled" binding:"required"`                     // 是否可用，必填
	UserType int      `json:"userType" binding:"required,oneof=0 1 2 3"`      // 用户类型，必填，值只能是0, 1, 2, 3
	Roles    []string `json:"roles" binding:"omitempty,dive,required,max=32"` // 权限，选填，数组中的每个元素最大32字符
	Mark     string   `json:"mark" binding:"omitempty,max=1024"`              // 备注，选填，最大1024字符
}

type UserLogin struct {
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
	Code     string `json:"code"`     // 谷歌验证码
}

type LoginBack struct {
	Username    string   `json:"username"`    // 用户名
	AccessToken string   `json:"accessToken"` // token
	QrCode      string   `json:"qrCode"`      // 需要绑定二维码时返回的二维码
	UserType    int      `json:"userType"`    // 用户类型 0 无效用户/1 管理员/2 后台管理/3 商户
	Roles       []string `json:"roles" `      // 角色
}

type OnlineUser struct {
	UserId    int64     `json:"userId"`
	Username  string    `json:"username"`
	Browser   string    `json:"browser"`
	Ip        string    `json:"ip"`
	Address   string    `json:"address"`
	Key       string    `json:"key"`
	LoginTime time.Time `json:"loginTime"`
}

type UserBack struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Username    string    `json:"username"`    // 账号(手机号)
	UniKey      string    `json:"uniKey"`      // 用户唯一码
	SecurityKey string    `json:"securityKey"` // 密钥
	UserType    int       `json:"userType"`    // 用户类型 0 无效用户/1 管理员/2 后台管理/3 商户
	Password    string    `json:"password"`    // 密码
	Enabled     bool      `json:"enabled"`     // 是否启用
	Ip          string    `json:"ip"`          // 最近登录ip
	Amount      float64   `json:"amount"`      // 可提现额度
	TopAmount   float64   `json:"topAmount"`   // 累计可提现额度
	GoogleCode  string    `json:"google_code"` // 验证器
	BindCode    bool      `json:"bindCode"`    // 是否绑定验证器
	Mark        string    `json:"mark"`        // 备注
	RoleStr     string    `json:"roleStr"`     // 角色 (JSON 字符串)
	Roles       []string  `json:"roles"`       // 角色 (不映射到数据库)
}

type UserResetPwd struct {
	ID          int64  `json:"id"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type UserResp struct {
	BasePageResponse[UserBack]
}

var UserTableName = "sys_user"

func (SysUser) TableName() string {
	return UserTableName
}

func (p *PageInfo) SetPageDefaults() {
	if p.CurrentPage == 0 {
		p.CurrentPage = 0
	}
	if p.PageSize == 0 {
		p.PageSize = 10
	}
}
