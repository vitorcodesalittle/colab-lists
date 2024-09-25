package community

import (
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
	err = Scan(rs, community)
	if err != nil {
		return nil, err
	}
	return community, nil
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
	result, err := tx.Exec(`INSERT INTO community (communityName, createdByLuserId)`, house.CommunityName, house.CreatedBy.Id)
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
	return result, nil
}
