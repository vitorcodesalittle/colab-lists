package community

import (
	"database/sql"
	"errors"

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

// TODO: validate if user saving is community creator if updating
func (h *HouseRepository) Save(house *Community, userId *int64) (*Community, error) {
	db, err := infra.CreateConnection()
	if err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if house.CommunityId == 0 {
		result, err := tx.Exec(`INSERT INTO community (communityName, createdByLuserId) VALUES (?, ?)`, house.CommunityName, house.CreatedBy.Id)
		if err != nil {
			return nil, err
		}
		house.CommunityId, err = result.LastInsertId()
		if err != nil {
			return nil, err
		}
	} else {
		_, err := tx.Exec(`UPDATE community SET communityName = ? WHERE communityId = ?`, house.CommunityName, house.CommunityId)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`DELETE FROM community_members WHERE communityId = ?`, house.CommunityId)
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
    WHERE c.createdByLuserId = ?
    OR c.communityId IN (SELECT m.communityId FROM community_members m WHERE m.memberId = ?)`, userId, userId)
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

func (h *HouseRepository) Delete(communityId, userId int64) error {
	db, err := infra.CreateConnection()
	if err != nil {
		return err
	}
	defer db.Close()
	rs := db.QueryRow(`SELECT createdByLuserId FROM community WHERE communityId = ?`, communityId)
	if rs.Err() != nil {
		return rs.Err()
	}
	var createdByLuserId int64
	err = rs.Scan(&createdByLuserId)
	if err != nil {
		return err
	}
	if createdByLuserId != userId {
		return errors.New("unauthorized: user deleting must be creator of community")
	}

	deleteResult, err := db.Exec(`DELETE FROM community WHERE communityId = ?`, communityId)
	if err != nil {
		return err
	}
	rowsAffected, err := deleteResult.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no rows deleted")
	}
	return nil
}

func GetDefault(comms []*Community) *Community {
	for _, comm := range comms {
		if comm != nil && comm.Default {
			return comm
		}
	}
	return nil
}
