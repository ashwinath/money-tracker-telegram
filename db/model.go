package db

import "time"

type TransactionType string

const TypeReimburse TransactionType = "REIM"
const TypeSharedReimburse TransactionType = "SHARED_REIM"
const TypeSpecialSharedReimburse TransactionType = "SPECIAL_SHARED_REIM"
const TypeOwn TransactionType = "OWN"

type Transaction struct {
	ID              uint            `gorm:"primaryKey"`
	Date            time.Time       `gorm:"index"`
	TransactionType TransactionType `gorm:"column:type"`
	Classification  string
	Amount          float64
}
