package infra

import (
	sql "database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func CreateConnection() (*sql.DB, error) {
	if os.Getenv("DATABASE_URL") == "" {
		panic("DATABASE_URL environment variable is not set")
	}
	return sql.Open("sqlite3", os.Getenv("DATABASE_URL"))
}

func GetDatabaseUrl() string {
	return os.Getenv("DATABASE_URL")
}


func UseConnection[T *any](do func(*sql.DB) (T, error)) (T, error) {
	db, err := CreateConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return do(db)
}
