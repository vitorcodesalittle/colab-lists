package list

import (
	"fmt"

	"vilmasoftware.com/colablists/pkg/user"
)

type List struct {
	Id           int64
	Title        string
	Description  string
	Colaborators []user.User
	Creator      user.User
	Groups       []*Group
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
        Title: %s,
        Description: %s,
        Creator: %s,
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
	return "Item " + string(i.Id) + ": " + i.Description
}
