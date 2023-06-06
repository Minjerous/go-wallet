package model

import "time"

const TableNameTransaction = "transaction"

type Transaction struct {
	Id          int64     `json:"id" form:"id" db:"id"`                            // 交易记录 id
	UserId      int64     `json:"user_id" form:"user_id" db:"user_id"`             // 用户 id
	Amount      float64   `json:"amount" form:"amount" db:"amount"`                // 交易金额数量
	Description string    `json:"description" form:"description" db:"description"` // 交易详细
	CreateTime  time.Time `gorm:"autoCreateTime" json:"create_time" form:"create_time" db:"create_time"`
}

// TableName Counselor's table name
func (*Transaction) TableName() string {
	return TableNameTransaction
}
