package base

import (
	"BaseGoUni/core/pojo"
	"gopkg.in/yaml.v3"
	"os"
)

type CsConfig struct {
	DefaultHost   pojo.HostInfo  `yaml:"defaultHost"`
	DefaultUser   pojo.SysUser   `yaml:"defaultUser"`
	DefaultRoles  []pojo.SysRole `yaml:"defaultRoles"`
	DefaultMenus  []pojo.SysMenu `yaml:"defaultMenus"`
	RunScheduler  bool           `yaml:"runScheduler"`
	TopInviteCode []InviteCode   `yaml:"topInviteCode"`
	LoginConfig   LoginConfig    `yaml:"loginConfig"`
	NewMenus      []pojo.SysMenu `yaml:"newMenus"`
	AwardIps      []string       `yaml:"awardIps"`
	AwardUrl      string         `yaml:"awardUrl"`
}

type LoginConfig struct {
	SingleLogin bool `yaml:"singleLogin"` // 是否会员单点登录
}

type InviteCode struct {
	Code string `yaml:"code"`
	Id   int64  `yaml:"id"`
}

func LoadCsConfig(file string, result *CsConfig) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	//log.Print("load config file.data=$data", string(data))
	return yaml.Unmarshal(data, &result)
}
