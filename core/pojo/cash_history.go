package pojo

type CashHistory struct { // 积分变动记录
	BaseModel
	UserId      int64   `yaml:"userId" json:"userId" gorm:"type:bigint;uniqueIndex:ut"`          // 用户id
	AwardUni    string  `yaml:"awardUni" json:"awardUni" gorm:"type:varchar(32);uniqueIndex:ut"` // 奖励唯一key
	Amount      float64 `yaml:"amount" json:"amount" gorm:"type:numeric(20,3)"`                  // 变动积分
	StartAmount float64 `yaml:"startAmount" json:"startAmount" gorm:"type:numeric(20,3)"`        // 变动前积分
	EndAmount   float64 `yaml:"endAmount" json:"endAmount" gorm:"type:numeric(20,3)"`            // 变动后积分
	CashMark    string  `yaml:"cashMark" json:"cashMark" gorm:"type:varchar(32);index:ccf"`      // 积分备注
	CashDesc    string  `yaml:"cashDesc" json:"cashDesc" gorm:"type:varchar(32);index:ccf"`      // 积分描述
	FromUserId  int64   `yaml:"fromUserId" json:"fromUserId" gorm:"type:bigint;index:ccf"`       // 来源用户id
}

type CashHistoryResp struct {
	UserId      int64   `json:"userId"`
	Amount      float64 `json:"amount"`
	StartAmount float64 `json:"startAmount"`
	CashMark    string  `json:"cashMark"`
}

type CashHistorySearch struct {
	PageInfo
	UserId int64 `json:"userId"`
}

type CashHistoryPage struct {
	BasePageResponse[CashHistoryResp]
}

var CashHistoryTableName = "cash_history"
var CashHistoryShardingName = "user_id"
var CashHistoryShards = 8
var AllCashHistoryShardingName = "all_cash_history"

func (CashHistory) TableName() string {
	return CashHistoryTableName
}
