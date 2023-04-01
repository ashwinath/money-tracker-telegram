package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate database
	db.AutoMigrate(&Transaction{})

	return &DB{DB: db}, nil
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

func (d *DB) Close() error {
	db, err := d.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
