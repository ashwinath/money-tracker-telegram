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

	return &DB{DB: db}, nil
}

// Migrate handles all migrations of database tables.
func (d *DB) Migrate() {
	d.DB.AutoMigrate(&Transaction{})
}
