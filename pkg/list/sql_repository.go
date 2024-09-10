package list

import (
	"log"

	infra "vilmasoftware.com/colablists/pkg/infra"
)

type SqlListRepository struct{}

type scannable interface {
	Scan(dest ...interface{}) error
}

// Create implements ListsRepository.
func (s *SqlListRepository) Create(list *ListCreationParams) (List, error) {
	sql, err := infra.CreateConnection()
	if err != nil {
		return List{}, err
	}
	stmt, err := sql.Prepare(`
    INSERT INTO list (title, description, creatorLuserId)
    VALUES (?, ?, ?)
    RETURNING listId
  `)
	if err != nil {
		log.Fatal("error preparing statement")
	}
	result, err := stmt.Exec(list.Title, list.Description, list.CreatorId)
	if err != nil {
		log.Fatal("error executing statement")
	}
	listId, err := result.LastInsertId()
	if err != nil {
		log.Fatal("error executing statement")
	}
	return s.Get(listId)
}

// Delete implements ListsRepository.
func (s *SqlListRepository) Delete(id int) error {
	panic("unimplemented")
}

type Scanner interface {
	Scan(dest ...interface{}) error
}

func scanList(row Scanner) (List, error) {
	l := &List{}
	err := row.Scan(&l.Id, &l.Title, &l.Description)
	if err != nil {
		return List{}, err
	}
	return *l, nil
}

// Get implements ListsRepository.
func (s *SqlListRepository) Get(id int64) (List, error) {
	sql, err := infra.CreateConnection()
	if err != nil {
		return List{}, err
	}
	stmt, err := sql.Prepare(`
  SELECT *
  FROM list
  Where listId = ?
  `)
	if err != nil {
		panic(err)
	}
	rs := stmt.QueryRow(id)
	return scanList(rs)
}

// GetAll implements ListsRepository.
func (s *SqlListRepository) GetAll() ([]List, error) {
	sql, err := infra.CreateConnection()
	if err != nil {
		panic(err)
	}
	rs, err := sql.Query(`
  SELECT *
  FROM list
  `)
	if err != nil {
		panic(err)
	}
	ls := make([]List, 0)
	for rs.Next() {
		l, err := scanList(rs)
		if err != nil {
			panic(err)
		}
		ls = append(ls, l)
	}
	return ls, nil
}

// Update implements ListsRepository.
func (s *SqlListRepository) Update(list *ListUpdateParams) (List, error) {
	panic("unimplemented")
}
