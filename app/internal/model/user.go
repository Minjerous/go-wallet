package model

import "time"

const TableNameUserSubject = "user_subject"

type UserSubject struct {
	Id          int64     `json:"id" form:"id" db:"id"`
	Username    string    `json:"username" form:"username" db:"username"`
	Password    string    `json:"password" form:"password" db:"password"`
	PaymentCode string    `json:"payment_code" form:"payment_code" db:"payment_code"`
	Phone       string    `json:"phone" form:"phone" db:"phone"`
	Email       string    `json:"email" form:"email" db:"email"`
	LastIp      string    `json:"last_ip" form:"last_ip" db:"last_ip"`
	Avatar      string    `json:"avatar" form:"avatar" db:"avatar"`
	Balance     float64   `json:"balance" form:"balance" db:"balance"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"create_time" form:"create_time" db:"create_time"`
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"update_time" form:"update_time" db:"update_time"`
}

// TableName Counselor's table name
func (*UserSubject) TableName() string {
	return TableNameUserSubject
}
