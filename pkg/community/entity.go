package community

import (
	"time"

	"vilmasoftware.com/colablists/pkg/user"
)

type Community struct {
	CommunityId   int64
	CommunityName string
	CreatedBy     *user.User
	Members       []*Member
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Default       bool
}

type Member struct {
	CommunityId int64
	user.User
}
