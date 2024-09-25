package community

import "vilmasoftware.com/colablists/pkg/user"

type Scannable interface {
	Scan(dest ...interface{}) error
}

func Scan(scannable Scannable, c *Community) error {
	if c.CreatedBy == nil {
		c.CreatedBy = &user.User{}
	}
	err := scannable.Scan(
		&c.CommunityId,
		&c.CommunityName,
        &c.CreatedBy.Id,
        &c.CreatedAt,
        &c.UpdatedAt,
        &c.Default,
	)
	if err != nil {
		return err
	}
	return nil
}
