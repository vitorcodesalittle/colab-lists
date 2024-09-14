package infra

import (
	sql "database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

func CreateConnection() (*sql.DB, error) {
  return sql.Open("sqlite3", "colablist.db")
}

func UseConnection[T *any](do func (*sql.DB) (T, error)) (T, error) {
  db, err := CreateConnection()
  if err != nil {
    return nil, err
  }
  defer db.Close()
  return do(db)
}


func ErrorRollback(err error, tx *sql.Tx) error {
    if errr := tx.Rollback(); errr != nil {
        return errors.New("Rollback error: " + err.Error() + "\nCaused by" + errr.Error())
    }
    return err;
}
