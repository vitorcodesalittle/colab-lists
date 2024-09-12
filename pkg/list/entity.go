package list

import "vilmasoftware.com/colablists/pkg/user"

type List struct {
	Id           int
	Title        string
	Description  string
	Colaborators []user.User
	Creator      user.User
	Groups       []Group
}

type Group struct {
	GroupId   int
	ListId    int
	CreatedAt string
	Name      string
	Items     []Item
}

func (l *List) String() string {
	return "List " + string(l.Id) + ": " + l.Title
}

type Item struct {
	Id          int
    GroupId     int
	Description string
	Quantity    int
	Order       int
}

func (i *Item) String() string {
	return "Item " + string(i.Id) + ": " + i.Description
}
