package main

import (
	"fmt"
	"math/rand"

	"vilmasoftware.com/colablists/pkg/list"
	"vilmasoftware.com/colablists/pkg/user"
)

var (
	IdListCurrent int64 = 1
	IdItemCurrent int64 = 1
	IdUserCurrent int64 = 1
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func MockUser() user.User {
	name := RandStringRunes(5)
	u := user.User{Id: 1, Username: fmt.Sprintf("%s@domain.com", name), Online: bool(rand.Intn(2) == 1)}
	IdUserCurrent += 1
	return u
}

func MockItem() *list.Item {
	item := &list.Item{
		Id:          IdItemCurrent + 1,
		Order:       IdItemCurrent + 1,
		Description: fmt.Sprintf("Item %d", IdItemCurrent),
		Quantity:    rand.Intn(5),
	}
	IdItemCurrent += 1
	return item
}

func MockList() list.List {
	numItems := rand.Intn(20)
	group := list.Group{Name: "Group test", Items: make([]*list.Item, numItems)}
	numColaborators := 7
	l := list.List{
		Id:           IdListCurrent,
		Title:        fmt.Sprintf("List %d", IdListCurrent),
		Description:  fmt.Sprintf("Description of list %d", IdListCurrent),
		Groups:       []*list.Group{&group},
		Colaborators: make([]user.User, numColaborators),
		Creator:      MockUser(),
	}
	IdListCurrent += 1
	for i := 0; i < numItems; i++ {
		group.Items[i] = MockItem()
	}
	for i := 0; i < numColaborators; i++ {
		l.Colaborators[i] = MockUser()
	}
	return l
}
