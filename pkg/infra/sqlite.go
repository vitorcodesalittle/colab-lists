package infra

import (
	sql "database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func CreateConnection() (*sql.DB, error) {
	return sql.Open("sqlite3", "colablist.db")
}

func UseConnection[T *any](do func(*sql.DB) (T, error)) (T, error) {
	db, err := CreateConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return do(db)
}
