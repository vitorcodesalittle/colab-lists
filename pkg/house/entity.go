package house

import "os/user"

type House struct {
	HouseId   int64
	HouseName string
	CreatedBy user.User
	Members   []HouseMember
}

type HouseMember struct {
	HouseId int64
	user.User
}
