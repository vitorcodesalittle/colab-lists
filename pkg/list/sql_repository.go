package list

import (
	"log"

	infra "vilmasoftware.com/colablists/pkg/infra"
	"vilmasoftware.com/colablists/pkg/user"
)

type SqlListRepository struct{}

type scannable interface {
	Scan(dest ...interface{}) error
}

// Create implements ListsRepository.
func (s *SqlListRepository) Create(list *ListCreationParams) (List, error) {
	sql, err := infra.CreateConnection()
	if err != nil {
		sql.Close()
		return List{}, err
	}
	tx, err := sql.Begin()
	if err != nil {
		return List{}, err
	}
	stmt, err := tx.Prepare(`
    INSERT INTO list (title, description, creatorLuserId)
    VALUES (?, ?, ?)
  `)
	if err != nil {
		return List{}, err
	}
	result, err := stmt.Exec(list.Title, list.Description, list.CreatorId)
	if err != nil {
		return List{}, err
	}

	listId, err := result.LastInsertId()
	if err != nil {
		return List{}, err
	}

	_, err = tx.Exec(`
    INSERT INTO list_colaborators (listId, luserId)
    VALUES (?, ?)
  `, listId, list.CreatorId)

	if err != nil {
		return List{}, err
	}

	result, err = tx.Exec("INSERT INTO list_groups (listId, name) VALUES (?, ?)", listId, "default")
	if err != nil {
		return List{}, err
	}

	groupId, err := result.LastInsertId()
	_, err = tx.Exec("INSERT INTO list_group_items (groupId, description, quantity, order_) VALUES (?, ?, ?, ?)", groupId, "default", 1, 1)
	if err != nil {
		return List{}, err
	}

	if err = tx.Commit(); err != nil {
		return List{}, err
	}

	sql.Close()
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
	l := &List{
		Creator: user.User{},
	}
	err := row.Scan(&l.Id, &l.Title, &l.Description, &l.Creator.Id)
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
	tx, err := sql.Begin()
	if err != nil {
		return List{}, err
	}
	stmt, err := tx.Prepare(`
    SELECT *
    FROM list
    Where listId = ?
    `)
	if err != nil {
		panic(err)
	}
	rs := stmt.QueryRow(id)
	resultlis, err := scanList(rs)
	if err != nil {
		log.Println("Failed to scan list")
		return List{}, err
	}

	stmt, err = tx.Prepare(`
    SELECT lu.*
    FROM luser lu
    INNER JOIN list_colaborators lc ON lu.luserId = lc.luserId
    WHERE lc.listId = ?
    `)
	if err != nil {
		log.Println("Failed to query luser")
		return List{}, infra.ErrorRollback(err, tx)
	}
	rscolaborators, err := stmt.Query(resultlis.Creator.Id)
	if err != nil {
		log.Println("Failed to query colaborators")
		return List{}, infra.ErrorRollback(err, tx)
	}
    defer rscolaborators.Close()
	colaborators := make([]user.User, 0)
	for rscolaborators.Next() {
		u, err := user.ScanUser(rscolaborators)
		if err != nil {
			log.Println("Failed to scan user")
			return List{}, infra.ErrorRollback(err, tx)
		}
		colaborators = append(colaborators, u)
	}

	stmt, err = tx.Prepare(`
    SELECT *
    FROM list_groups
    WHERE listId = ?
    `)
	rs2, err := stmt.Query(id)
    if err != nil {
        return List{}, err
    }
    defer rs2.Close()
	groups := make([]Group, 0)
	for rs2.Next() {
		g := Group{Items: make([]Item, 0)}
		err := rs2.Scan(&g.GroupId, &g.ListId, &g.CreatedAt, &g.Name)
		if err != nil {
			log.Println("Failed to scan group")
			return List{}, infra.ErrorRollback(err, tx)
		}
		stmt, err = tx.Prepare(`
        SELECT *
        FROM list_group_items
        WHERE groupId = ?
        `)
		if err != nil {
			log.Println("Failed to prepare statement for items")
			return List{}, infra.ErrorRollback(err, tx)
		}
		rsg, err := stmt.Query(g.GroupId)
		if err != nil {
			log.Println("Error querying items")
			return List{}, infra.ErrorRollback(err, tx)
		}
        defer rsg.Close()
		for rsg.Next() {
			i := Item{}
			err := rsg.Scan(&i.Id, &i.GroupId, &i.Description, &i.Quantity, &i.Order)
			if err != nil {
				log.Println("Failed to scan item")
				return List{}, infra.ErrorRollback(err, tx)
			}
			g.Items = append(g.Items, i)
		}
		groups = append(groups, g)
	}
	resultlis.Groups = groups
	return resultlis, nil
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
    defer rs.Close()
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
