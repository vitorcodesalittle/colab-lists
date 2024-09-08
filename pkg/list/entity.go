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
	Name  string
	Items []Item
}

func (l *List) String() string {
	return "List " + string(l.Id) + ": " + l.Title
}

type Item struct {
	Id          int
	Order       int
	Description string
	Quantity    int
}

func (i *Item) String() string {
	return "Item " + string(i.Id) + ": " + i.Description
}
