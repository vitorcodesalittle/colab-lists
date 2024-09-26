package community

import (
	"database/sql"
	"fmt"

	"vilmasoftware.com/colablists/pkg/infra"
)

type HouseRepository struct{}

func (h *HouseRepository) Get(id int64) (*Community, error) {
	db, err := infra.CreateConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rs := db.QueryRow("SELECT * FROM community WHERE communityId = ?", id)
	community := &Community{}
	community.CommunityId = id
	err = Scan(rs, community)
	if err != nil {
		return nil, err
	}
	community.Members, err = h.GetMembers(db, community)
	if err != nil {
		return nil, err
	}
	return community, nil
}

func (h *HouseRepository) GetMembers(tx infra.Queryable, community *Community) ([]*Member, error) {
	rows, err := tx.Query(`SELECT m.communityId, m.memberId, u.username, u.passwordHash, u.passwordSalt, u.email, u.avatarUrl
    FROM community_members m
    JOIN luser u ON m.memberId = u.luserId
    WHERE m.communityId = ?`, community.CommunityId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		member := &Member{CommunityId: community.CommunityId}
		err = ScanMember(rows, member)
		if err != nil {
			return nil, err
		}
		community.Members = append(community.Members, member)
	}
	return community.Members, nil
}

func ScanMember(rows *sql.Rows, m *Member) error {
	return rows.Scan(&m.CommunityId, &m.Id, &m.Username, &m.PasswordHash, &m.PasswordSalt, &m.Email, &m.AvatarUrl)
}

func (h *HouseRepository) Save(house *Community) (*Community, error) {
	db, err := infra.CreateConnection()
	if err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	fmt.Printf("house: %v\n", house)
	result, err := tx.Exec(`INSERT INTO community (communityName, createdByLuserId) VALUES (?, ?)`, house.CommunityName, house.CreatedBy.Id)
	if err != nil {
		return nil, err
	}
	house.CommunityId, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	for _, member := range house.Members {
		_, err = tx.Exec(`INSERT INTO community_members (communityId, memberId)
        VALUES (?, ?)`, house.CommunityId, member.Id)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return h.Get(house.CommunityId)
}

func (h *HouseRepository) FindMyHouses(userId int64) ([]*Community, error) {
	result := make([]*Community, 0)
	db, err := infra.CreateConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT c.*
    FROM community c
    LEFT JOIN community_members m ON c.communityId = m.memberId
    WHERE m.memberId = ?
    OR c.createdByLuserId = ?`, userId, userId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		community := &Community{}
		err = Scan(rows, community)
		if err != nil {
			return nil, err
		}
		result = append(result, community)
	}
	return result, nil
}
