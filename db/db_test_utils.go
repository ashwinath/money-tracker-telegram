package db

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func createTestDatabase() (*string, error) {
	db, err := New(
		"localhost",
		"postgres",
		"password",
		"postgres",
		5432,
	)
	if err != nil {
		return nil, err
	}

	databaseName := fmt.Sprintf("%s_%s", "test", randStringRunes(10))

	// Drop database if exists
	stmt := fmt.Sprintf("DROP DATABASE IF EXISTS %s;", databaseName)
	if rs := db.DB.Exec(stmt); rs.Error != nil {
		return nil, rs.Error
	}

	// if not create it
	stmt = fmt.Sprintf("CREATE DATABASE %s;", databaseName)
	if rs := db.DB.Exec(stmt); rs.Error != nil {
		return nil, rs.Error
	}

	// close db connection
	sql, err := db.DB.DB()
	defer func() {
		_ = sql.Close()
	}()
	if err != nil {
		return nil, err
	}

	return &databaseName, nil
}

func dropDatabase(name string) error {
	db, err := New(
		"localhost",
		"postgres",
		"password",
		"postgres",
		5432,
	)
	if err != nil {
		return err
	}

	// Drop database if exists
	stmt := fmt.Sprintf("DROP DATABASE IF EXISTS %s;", name)
	if rs := db.DB.Exec(stmt); rs.Error != nil {
		return rs.Error
	}

	return nil
}

func RunTest(function func(db *DB)) error {
	dbName, err := createTestDatabase()
	if err != nil {
		return err
	}

	db, err := New(
		"localhost",
		"postgres",
		"password",
		*dbName,
		5432,
	)

	defer dropDatabase(*dbName)
	defer db.Close()

	function(db)

	return nil
}
