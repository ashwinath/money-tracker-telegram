package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	err := RunTest(func(db *DB) {
		loc, _ := time.LoadLocation("Asia/Singapore")
		tx := &Transaction{
			Date:           time.Now().In(loc),
			Type:           TypeReimburse,
			Classification: "negative",
			Amount:         254.23,
		}
		res, err := db.InsertTransaction(tx)
		assert.Nil(t, err)
		assert.Equal(t, tx.Type, res.Type)
		assert.Equal(t, tx.Classification, res.Classification)
		assert.Equal(t, tx.Amount, res.Amount)
	})

	assert.Nil(t, err)
}
