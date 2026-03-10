package storage

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type Database struct {
	*sql.DB
}

func NewSqlite(url string) (*Database, error) {
	db, err := sql.Open("sqlite", url)
	if err != nil {
		return nil, err
	}

	return &Database{db}, nil
}

const (
	CREATE_TABLE_FORMAT string = "CREATE TABLE IF NOT EXISTS %s (%s);"
	INSERT_INTO_FORMAT  string = "INSERT INTO %s (%s) VALUES (%s)"
)

type Query struct {
	Table   string
	Columns []string
	Values  []any
}

func (db *Database) Close() {
	db.DB.Close()
}

func (db *Database) CreateTable(q *Query) error {

	stmt := fmt.Sprintf(CREATE_TABLE_FORMAT, q.Table, strings.Join(q.Columns, ", "))
	_, err := db.DB.Exec(stmt)

	if err != nil {
		return err
	}

	return nil
}

func (db *Database) CreateId() (string, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("could not create Id: %w", err)
	}

	return uuid.String(), nil
}
