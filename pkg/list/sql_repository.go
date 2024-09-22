package list

import (
	"errors"
	"log"
	"time"

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
	err := row.Scan(&l.Id, &l.Title, &l.Description, &l.Creator.Id, &l.UpdatedAt)
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
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
    SELECT *
    FROM list
    Where listId = ?
    `)
	if err != nil {
		println("error at scan")
		panic(err)
	}
	rs := stmt.QueryRow(id)
	resultlis, err := scanList(rs)
	if err != nil {
		println("error at scan list")
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
		return List{}, err
	}
	rscolaborators, err := stmt.Query(resultlis.Creator.Id)
	if err != nil {
		println("error at colab query")
		return List{}, err
	}
	defer rscolaborators.Close()
	colaborators := make([]user.User, 0)
	for rscolaborators.Next() {
		u, err := user.ScanUser(rscolaborators)
		if err != nil {
			return List{}, err
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
		println("error at group query")
		return List{}, err
	}
	defer rs2.Close()
	groups := make([]*Group, 0)
	for rs2.Next() {
		g := &Group{Items: make([]*Item, 0)}
		err := rs2.Scan(&g.GroupId, &g.ListId, &g.CreatedAt, &g.Name)
		if err != nil {
			return List{}, err
		}
		stmt, err = tx.Prepare(`
        SELECT *
        FROM list_group_items
        WHERE groupId = ?
        `)
		if err != nil {
			println("error at item query")
			return List{}, err
		}
		rsg, err := stmt.Query(g.GroupId)
		if err != nil {
			return List{}, err
		}
		defer rsg.Close()
		for rsg.Next() {
			i := Item{}
			err := rsg.Scan(&i.Id, &i.GroupId, &i.Description, &i.Quantity, &i.Order)
			if err != nil {
				return List{}, err
			}
			g.Items = append(g.Items, &i)
		}
		groups = append(groups, g)
	}
	err = tx.Commit()
	if err != nil {
		return List{}, err
	}
	resultlis.Groups = groups
	return resultlis, nil
}

// GetAll implements ListsRepository.
func (s *SqlListRepository) GetAll(userId int64) ([]List, error) {
	sql, err := infra.CreateConnection()
	if err != nil {
		panic(err)
	}
    defer sql.Close()
	rs, err := sql.Query(`
  SELECT l.*
  FROM list l
  LEFT JOIN list_colaborators lc ON l.listId = lc.listId
  LEFT JOIN luser lu ON lc.luserId = lu.luserId
  WHERE lu.luserId = ?
  OR l.creatorLuserId = ?
  ORDER BY l.updatedAt DESC
  `, userId, userId)
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
func (s *SqlListRepository) Update(list *List) (*List, error) {
	if list == nil || list.Id <= 0 {
		return nil, errors.New("list.Id must be a positive integer")
	}
	sql, err := infra.CreateConnection()
	if err != nil {
		return nil, err
	}
	defer sql.Close()

	tx, err := sql.Begin()
	defer tx.Rollback()

	list.UpdatedAt = time.Now()
	_, err = tx.Exec(`
        UPDATE list
        SET title = ?,
        description = ?,
        updatedAt = ?
        WHERE listId = ?
    `, list.Title, list.Description, list.UpdatedAt, list.Id)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(`
        DELETE FROM list_groups
        WHERE listId = ?
    `, list.Id)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(`
        DELETE FROM list_colaborators
        WHERE listId = ?
    `, list.Id)
	if err != nil {
		return nil, err
	}
	for _, group := range list.Groups {
		result, err := tx.Exec(`
            INSERT INTO list_groups (listId, name)
            VALUES (?, ?)
        `, list.Id, group.Name)
		if err != nil {
			return nil, err
		}
		groupId, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		for itemindex, item := range group.Items {
			_, err = tx.Exec(`
                INSERT INTO list_group_items (groupId, description, quantity, order_)
                VALUES (?, ?, ?, ?)
            `, groupId, item.Description, item.Quantity, itemindex)
			if err != nil {
				return nil, err
			}
		}
	}
	for _, user := range list.Colaborators {
		_, err = tx.Exec(`
                INSERT INTO list_colaborators (listId, luserId)
                VALUES (?, ?)
            `, list.Id, user.Id)
		if err != nil {
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return list, nil
}
