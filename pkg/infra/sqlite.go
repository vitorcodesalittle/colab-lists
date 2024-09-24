package infra

import (
	sql "database/sql"

	_ "github.com/mattn/go-sqlite3"
	"vilmasoftware.com/colablists/pkg/config"
)


func CreateConnection() (*sql.DB, error) {
	return sql.Open("sqlite3", config.GetConfig().DatabaseUrl)
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
