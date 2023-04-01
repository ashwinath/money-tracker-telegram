package db

import "time"

type TransactionType string

const TypeReimburse TransactionType = "REIM"
const TypeSharedReimburse TransactionType = "SHARED_REIM"
const TypeShared TransactionType = "SHARED"
const TypeSpecialShared TransactionType = "SPECIAL_SHARED"
const TypeSpecialSharedReimburse TransactionType = "SPECIAL_SHARED_REIM"
const TypeOwn TransactionType = "OWN"
const TypeSpecialOwn TransactionType = "SPECIAL_OWN"

type Transaction struct {
	ID             uint            `gorm:"primaryKey"`
	Date           time.Time       `gorm:"index"`
	Type           TransactionType `gorm:"column:type"`
	Classification string
	Amount         float64
}
