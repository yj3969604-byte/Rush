package pojo

type SysMenu struct {
	BaseModel
	ParentID   int64     `yaml:"parentID" json:"parentID" gorm:"type:bigint"`          // 父菜单
	MenuType   int       `yaml:"menuType" json:"menuType" gorm:"type:int"`             // 菜单类型 0 菜单/1 iframe/2 链接/3 按钮
	Name       string    `yaml:"name" json:"name" gorm:"type:varchar(32);uniqueIndex"` // 路由名字（必须保持唯一）
	Path       string    `yaml:"path" json:"path" gorm:"type:varchar(64)"`             // 路由地址
	Component  string    `yaml:"component" json:"component" gorm:"type:varchar(64)"`   // 按需加载需要展示的页面
	MetaStr    string    `yaml:"metaStr" json:"metaStr" gorm:"type:text"`              // 路由元信息 JSON 字符串
	Meta       MenuMeta  `yaml:"meta" json:"meta" gorm:"-"`                            // 路由元信息，程序使用
	ParentName string    `yaml:"parentName" json:"parentName" gorm:"-"`                // 父路由名字，程序使用
	Children   []SysMenu `yaml:"children" json:"children,omitempty" gorm:"-"`          // 子路由，程序使用
}

type Transition struct {
	EnterTransition string `yaml:"enterTransition" json:"enterTransition"` // 进场动画
	LeaveTransition string `yaml:"leaveTransition" json:"leaveTransition"` // 离场动画
}

type MenuMeta struct {
	Title        string     `yaml:"title" json:"title"`               // 菜单名称
	Rank         int        `yaml:"rank" json:"rank"`                 // 菜单排序（平台规定只有`home`路由的`rank`才能为`0`，所以后端在返回`rank`的时候需要从非`0`开始
	Redirect     string     `yaml:"redirect" json:"redirect"`         // 路由重定向
	Icon         string     `yaml:"icon" json:"icon"`                 // 菜单图标
	ExtraIcon    string     `yaml:"extraIcon" json:"extraIcon"`       // 右侧图标
	ActivePath   string     `yaml:"activePath" json:"activePath"`     // 菜单激活（将某个菜单激活，主要用于通过`query`或`params`传参的路由，当它们通过配置`showLink: false`后不在菜单中显示，就不会有任何菜单高亮，而通过设置`activePath`指定激活菜单即可获得高亮，`activePath`为指定激活菜单的`path`）
	Auths        []string   `yaml:"auths" json:"auths"`               // 权限标识（按钮级别权限设置）
	FrameSrc     string     `yaml:"frameSrc" json:"frameSrc"`         // 需要内嵌的iframe链接地址
	FrameLoading bool       `yaml:"frameLoading" json:"frameLoading"` // 内嵌的iframe页面是否开启首次加载动画
	KeepAlive    bool       `yaml:"keepAlive" json:"keepAlive"`       // 是否缓存该路由页面（开启后，会保存该页面的整体状态，刷新后会清空状态）
	HiddenTag    bool       `yaml:"hiddenTag" json:"hiddenTag"`       // 当前菜单名称或自定义信息禁止添加到标签页
	FixedTag     bool       `yaml:"fixedTag" json:"fixedTag"`         // 固定标签页（当前菜单名称是否固定显示在标签页且不可关闭）
	ShowLink     bool       `yaml:"showLink" json:"showLink"`         // 是否在菜单中显示
	ShowParent   bool       `yaml:"showParent" json:"showParent"`     // 是否显示父级菜单
	Transition   Transition `yaml:"transition" json:"transition"`     // 动画
}

type BackMenu struct {
	ID        int64      `yaml:"id" json:"id"`
	ParentID  int64      `yaml:"parentId" json:"parentId"`           // 父菜单
	MenuType  int        `yaml:"menuType" json:"menuType"`           // 菜单类型 1 菜单/2 链接/3 跳转
	Path      string     `yaml:"path" json:"path"`                   // 路由地址
	Name      string     `yaml:"name" json:"name"`                   // 路由名字（必须保持唯一）
	Component string     `yaml:"component" json:"component"`         // 按需加载需要展示的页面
	Meta      MenuMeta   `yaml:"meta" json:"meta"`                   // 路由元信息
	Children  []BackMenu `yaml:"children" json:"children,omitempty"` // 子路由
	NameCode  string     `yaml:"nameCode" json:"nameCode"`           // 路由名字（必须保持唯一）
}

type MenuSet struct {
	ID              int64  `yaml:"id" json:"id"`
	MenuType        int    `yaml:"menuType" json:"menuType" binding:"required,oneof=0 1 2 3"`         // 菜单类型（0-菜单, 1-iframe, 2-链接, 3-按钮）
	ParentID        int64  `yaml:"parentId" json:"parentId" binding:"required"`                       // 父菜单
	Title           string `yaml:"title" json:"title" binding:"required,max=128"`                     // 菜单名称，必填，最大128字符
	Name            string `yaml:"name" json:"name" binding:"required,max=32"`                        // 路由名字，必填，最长32个字符
	Path            string `yaml:"path" json:"path" binding:"required,max=64"`                        // 路由地址，必填，最长64个字符
	Component       string `yaml:"component" json:"component" binding:"omitempty,max=64"`             // 按需加载需要展示的页面，可选，最长64个字符
	Rank            int    `yaml:"rank" json:"rank" binding:"required,min=1"`                         // 菜单排序，必须大于等于 1
	Redirect        string `yaml:"redirect" json:"redirect" binding:"omitempty,url"`                  // 路由重定向，可选，必须是有效的URL
	Icon            string `yaml:"icon" json:"icon" binding:"omitempty,max=64"`                       // 菜单图标，可选，最大64字符
	ExtraIcon       string `yaml:"extraIcon" json:"extraIcon" binding:"omitempty,max=64"`             // 右侧图标，可选，最大64字符
	EnterTransition string `yaml:"enterTransition" json:"enterTransition" binding:"omitempty,max=64"` // 进场动画，可选，最大64字符
	LeaveTransition string `yaml:"leaveTransition" json:"leaveTransition" binding:"omitempty,max=64"` // 离场动画，可选，最大64字符
	ActivePath      string `yaml:"activePath" json:"activePath" binding:"omitempty,max=64"`           // 菜单激活路径，可选，最大64字符
	Auths           string `yaml:"auths" json:"auths" binding:"omitempty,max=256"`                    // 权限标识，可选，最大256字符
	FrameSrc        string `yaml:"frameSrc" json:"frameSrc" binding:"omitempty,url"`                  // iframe 链接地址，可选，必须是URL
	FrameLoading    bool   `yaml:"frameLoading" json:"frameLoading" binding:"omitempty"`              // 是否开启首次加载动画
	KeepAlive       bool   `yaml:"keepAlive" json:"keepAlive" binding:"omitempty"`                    // 是否缓存页面
	HiddenTag       bool   `yaml:"hiddenTag" json:"hiddenTag" binding:"omitempty"`                    // 禁止添加到标签页
	FixedTag        bool   `yaml:"fixedTag" json:"fixedTag" binding:"omitempty"`                      // 固定标签页
	ShowLink        bool   `yaml:"showLink" json:"showLink" binding:"omitempty"`                      // 是否显示在菜单中
	ShowParent      bool   `yaml:"showParent" json:"showParent" binding:"omitempty"`                  // 是否显示父级菜单
}

var MenuTableName = "sys_menu"

func (SysMenu) TableName() string {
	return MenuTableName
}
