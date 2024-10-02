package list

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"vilmasoftware.com/colablists/pkg/community"
	infra "vilmasoftware.com/colablists/pkg/infra"
	"vilmasoftware.com/colablists/pkg/user"
)

type SqlListRepository struct{}

// Create implements ListsRepository.
func (s *SqlListRepository) Create(list *ListCreationParams) (List, error) {
	db, err := infra.CreateConnection()
	if err != nil {
		db.Close()
		return List{}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return List{}, err
	}
	stmt, err := tx.Prepare(`
    INSERT INTO list (title, description, creatorLuserId, communityId)
    VALUES (?, ?, ?, ?)
  `)
	if err != nil {
		return List{}, err
	}
	result, err := stmt.Exec(list.Title, list.Description, list.CreatorId, list.CommunityId)
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
	if err != nil {
		return List{}, err
	}
	_, err = tx.Exec("INSERT INTO list_group_items (groupId, description, quantity, order_) VALUES (?, ?, ?, ?)", groupId, "default", 1, 1)
	if err != nil {
		return List{}, err
	}

	if err = tx.Commit(); err != nil {
		return List{}, err
	}

	db.Close()
	return s.Get(listId)
}

// Delete implements ListsRepository.
func (s *SqlListRepository) Delete(listId, userId int64) error {
	db, err := infra.CreateConnection()
	if err != nil {
		return err
	}
	if _, err := db.Exec(`DELETE FROM list where listId = ? and creatorLuserId = ?`, listId, userId); err != nil {
		return err
	}
	return nil
}

type Scanner interface {
	Scan(dest ...interface{}) error
}

func scanList(row Scanner) (List, error) {
	l := &List{
		Creator: user.User{},
	}
	var communityId *int64
	err := row.Scan(&l.Id, &l.Title, &l.Description, &l.Creator.Id, &l.UpdatedAt, &communityId)
	if err != nil {
		return List{}, err
	}
	if communityId != nil {
		l.Community = &community.Community{CommunityId: *communityId}
	}
	return *l, nil
}

func GetCommunity(tx infra.Queryable, comm *community.Community) error {
	row := tx.QueryRow(`SELECT c.communityName, c.createdByLuserId, c.default_, c.createdAt, c.updatedAt from community c where communityId = ?`, comm.CommunityId)
	if row.Err() != nil {
		return row.Err()
	}
	comm.CreatedBy = &user.User{}
	row.Scan(&comm.CommunityName, &comm.CreatedBy.Id, comm.Default, &comm.CreatedAt, &comm.UpdatedAt)
	return nil
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
		log.Fatal(err)
	}
	rs := stmt.QueryRow(id)
	resultlis, err := scanList(rs)
	if err != nil {
		log.Println("error at scan list")
		log.Fatalln(err)
		return List{}, err
	}
	if resultlis.Community != nil {
		if err = GetCommunity(tx, resultlis.Community); err != nil {
			return List{}, err
		}
	}

	stmt, err = tx.Prepare(`
    SELECT lu.*
    FROM luser lu
    INNER JOIN list_colaborators lc ON lu.luserId = lc.luserId
    WHERE lc.listId = ?
    `)
	if err != nil {
		return List{}, err
	}
	rscolaborators, err := stmt.Query(id)
	if err != nil {
		return List{}, err
	}
	defer rscolaborators.Close()
	colaborators := make([]user.User, 0)
	for rscolaborators.Next() {
		u, err := user.UnsafeScanUser(rscolaborators)
		if err != nil {
			return List{}, err
		}
		colaborators = append(colaborators, u)
	}
	resultlis.Colaborators = colaborators

	stmt, err = tx.Prepare(`
    SELECT *
    FROM list_groups
    WHERE listId = ?
    `)
	if err != nil {
		return List{}, err
	}
	rs2, err := stmt.Query(id)
	if err != nil {
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
			return List{}, err
		}
		rsg, err := stmt.Query(g.GroupId)
		if err != nil {
			return List{}, err
		}
		defer rsg.Close()
		for rsg.Next() {
			i := Item{}
			err := rsg.Scan(&i.Id, &i.GroupId, &i.Description, &i.Quantity, &i.Order, &i.Checked)
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
	db, err := infra.CreateConnection()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	rs, err := db.Query(`
  SELECT l.*
  FROM list l
  WHERE l.creatorLuserId = ?
  OR l.listId IN (SELECT listId FROM list_colaborators WHERE luserId = ?)
  OR l.listId IN (SELECT listId FROM community_members WHERE memberId = ?)
  ORDER BY l.updatedAt DESC
  `, userId, userId, userId)
	if err == sql.ErrNoRows {
		return make([]List, 0), nil
	} else if err != nil {
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
	comms := make(map[int64]*community.Community)
	for _, list := range ls {
		log.Printf("Getting comm for list %v\n", list)
		if list.Community != nil {
			comm, ok := comms[list.Community.CommunityId]
			if !ok {
				err := GetCommunity(db, list.Community)
				if err != nil {
					return nil, err
				}
				comms[list.Community.CommunityId] = list.Community
			} else {
				list.Community = comm
			}
			log.Printf("Got comm %v\n", list.Community)
		} else {
			log.Printf("No comm")
		}
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
	if err != nil {
		return nil, err
	}
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
                INSERT INTO list_group_items (groupId, description, quantity, order_, checked)
                VALUES (?, ?, ?, ?, ?)
            `, groupId, item.Description, item.Quantity, itemindex, item.Checked)
			if err != nil {
				return nil, err
			}
		}
	}
	println("Inserting colaborators")
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
