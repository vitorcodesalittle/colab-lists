package list

import (
	"fmt"
	"strconv"
	"time"

	"vilmasoftware.com/colablists/pkg/user"
)

type List struct {
	Id           int64
	Title        string
	Description  string
	Colaborators []user.User
	Creator      user.User
	Groups       []*Group
	CreatedAt    time.Time
	UpdatedAt    time.Time
	HouseId      *int64
}

type Group struct {
	GroupId   int64
	ListId    int64
	CreatedAt string
	Name      string
	Items     []*Item
}

func (l *List) String() string {
	return fmt.Sprintf(`List {
        Id: %d,
        Title: %v,
        Description: %v,
        Creator: %v,
        Colaborators: %v,
        Groups: %v
    }`, l.Id, l.Title, l.Description, l.Creator, l.Colaborators, l.Groups)
}

type Item struct {
	Id          int64
	GroupId     int64
	Description string
	Quantity    int
	Order       int64
}

func (i *Item) String() string {
	return "Item " + strconv.FormatInt(i.Id, 10) + ": " + i.Description
}
