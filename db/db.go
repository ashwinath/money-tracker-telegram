package db

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	DB *gorm.DB
}

// New initialises a new database object.
func New(host, user, password, dbName string, port uint) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Singapore",
		host,
		user,
		password,
		dbName,
		port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			ti, _ := time.LoadLocation("Asia/Singapore")
			return time.Now().In(ti)
		},
	})
	if err != nil {
		return nil, err
	}

	// Migrate database
	db.AutoMigrate(&Transaction{})

	return &DB{DB: db}, nil
}

func (d *DB) Close() error {
	db, err := d.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (d *DB) InsertTransaction(tx *Transaction) (*Transaction, error) {
	result := d.DB.Create(tx)
	if result.Error != nil {
		return nil, result.Error
	}

	return d.queryTransaction(tx.ID)
}

func (d *DB) queryTransaction(id uint) (*Transaction, error) {
	tx := &Transaction{}
	result := d.DB.First(tx, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return tx, nil
}

// returns the old copy of the deleted transaction
func (d *DB) DeleteTransaction(id uint) (*Transaction, error) {
	deletedTx, err := d.queryTransaction(id)
	if err != nil {
		return nil, err
	}

	result := d.DB.Delete(&Transaction{ID: id})
	if result.Error != nil {
		return nil, result.Error
	}

	return deletedTx, nil
}

type FindTransactionOptions struct {
	StartDate time.Time
	EndDate   time.Time
	Types     []TransactionType
}

type findTransactionResult struct {
	Total float64
}

func (d *DB) AggregateTransactions(o *FindTransactionOptions) (*float64, error) {
	result := findTransactionResult{}
	res := d.DB.Model(&Transaction{}).
		Select("sum(amount) as total").
		Where("date >= ? and date < ? and type in ?", o.StartDate, o.EndDate, o.Types).
		Scan(&result)

	if res.Error != nil {
		return nil, res.Error
	}

	return &result.Total, nil
}

func (d *DB) QueryTransactionByOptions(o *FindTransactionOptions) ([]Transaction, error) {
	var transactions []Transaction
	result := d.DB.
		Where("date >= ? and date < ? and type in ?", o.StartDate, o.EndDate, o.Types).
		Order("date asc").
		Find(&transactions)

	if result.Error != nil {
		return nil, result.Error
	}

	return transactions, nil
}
