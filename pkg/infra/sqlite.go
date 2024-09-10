package infra

import (
	sql "database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func CreateConnection() (*sql.DB, error) {
  log.Printf("Drivers are %v\n", sql.Drivers)
  return sql.Open("sqlite3", "colablist.db")
}

func test() {
  db, err := CreateConnection()
  log.Fatal(err)
  tx, err := db.Begin();
  log.Fatal(err)
  tx.Prepare(`INSERT INTO lists (title, description) VALUES (?, ?)`)
}
