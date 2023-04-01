package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	err := RunTest(func(db *DB) {
		tx := &Transaction{
			Date:           time.Now(),
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

func TestDelete(t *testing.T) {
	err := RunTest(func(db *DB) {
		tx := &Transaction{
			Date:           time.Now(),
			Type:           TypeReimburse,
			Classification: "negative",
			Amount:         254.23,
		}
		_, err := db.InsertTransaction(tx)
		assert.Nil(t, err)

		deletedTx, err := db.DeleteTransaction(tx.ID)
		assert.Nil(t, err)

		assert.Equal(t, tx.Type, deletedTx.Type)
		assert.Equal(t, tx.Classification, deletedTx.Classification)
		assert.Equal(t, tx.Amount, deletedTx.Amount)
	})

	assert.Nil(t, err)
}
