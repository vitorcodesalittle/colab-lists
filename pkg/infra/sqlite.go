package infra

import (
	sql "database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"vilmasoftware.com/colablists/pkg/config"
)

type Queryable interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func CreateConnection() (*sql.DB, error) {
	databaseUrl := config.GetConfig().DatabaseUrl
	if _, err := os.Stat(databaseUrl); os.IsNotExist(err) {
		if err = os.MkdirAll("./data", os.ModePerm); err != nil {
			return nil, err
		}
		println("Creating database file")
		if _, err = os.Create(databaseUrl); err != nil {
			return nil, err
		}
	}
	return sql.Open("sqlite3", databaseUrl)
}

func GetDatabaseUrl() string {
	return config.GetConfig().DatabaseUrl
}

func UseConnection[T *any](do func(*sql.DB) (T, error)) (T, error) {
	db, err := CreateConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return do(db)
}
