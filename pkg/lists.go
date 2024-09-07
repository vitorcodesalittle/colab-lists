package lists

import (
	"errors"
	"fmt"
	"math/rand"
)

type User struct {
  Id          int
  Name        string
  Email       string
  Online      bool
}

type List struct {
	Id          int
	Title       string
	Description string
	Items       []Item
  Colaborators []User
  Creator User
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

type ListCreationParams struct {
	Title       string
	Description string
}
type ListUpdateParams struct {
	Id int
	ListCreationParams
	Items []Item
}

type ListsRepository interface {
	GetAll() ([]List, error)
	Get(id int) (List, error)
	Create(list *ListCreationParams) (List, error)
	Update(list *ListUpdateParams) (List, error)
	Delete(id int) error
}

type ListsInMemoryRepository struct {
	lists []List
}

func NewListsInMemoryRepository(lists []List) *ListsInMemoryRepository {
	return &ListsInMemoryRepository{
		lists,
	}
}

func (r *ListsInMemoryRepository) GetAll() ([]List, error) {
	return r.lists, nil
}

func (r *ListsInMemoryRepository) Get(id int) (List, error) {
	for _, list := range r.lists {
		if list.Id == id {
			return list, nil
		}
	}
	return List{}, nil
}

func (r *ListsInMemoryRepository) Create(params *ListCreationParams) (List, error) {
	list := List{
		Id:          len(r.lists) + 1,
		Title:       params.Title,
		Description: params.Description,
	}
	r.lists = append(r.lists, list)
	return list, nil
}

func (r *ListsInMemoryRepository) Update(params *ListUpdateParams) (List, error) {
	for i, list := range r.lists {
		if list.Id == params.Id {
			r.lists[i].Title = params.Title
			r.lists[i].Description = params.Description
			r.lists[i].Items = params.Items
			return r.lists[i], nil
		}
	}
	return List{}, errors.New("List not found")
}

func (r *ListsInMemoryRepository) Delete(id int) error {
	for i, list := range r.lists {
		if list.Id == id {
			r.lists = append(r.lists[:i], r.lists[i+1:]...)
			return nil
		}
	}
	return errors.New("List not found")
}

var IdListCurrent = 1
var IdItemCurrent = 1
var IdUserCurrent = 1

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}
func MockUser() User {
  name := RandStringRunes(5)
  u := User{Id: 1, Name: name, Email: fmt.Sprintf("%s@domain.com", name), Online: bool(rand.Intn(2) == 1)}
  IdUserCurrent += 1
  return u
}

func MockItem() Item {
  item := Item{
		Id:          IdItemCurrent + 1,
		Order:       IdItemCurrent + 1,
		Description: fmt.Sprintf("Item %d", IdItemCurrent),
		Quantity:    rand.Intn(5),
	}
  IdItemCurrent += 1
  return item
}

func MockList() List {
	numItems := rand.Intn(20)
  numColaborators := 7
	l := List{
		Id:          IdListCurrent,
		Title:       fmt.Sprintf("List %d", IdListCurrent),
		Description: fmt.Sprintf("Description of list %d", IdListCurrent),
		Items:       make([]Item, numItems),
    Colaborators: make([]User, numColaborators),
    Creator: MockUser(),
	}
	IdListCurrent += 1
	for i := 0; i < numItems; i++ {
		l.Items[i] = MockItem()
	}
	for i := 0; i < numColaborators; i++ {
		l.Colaborators[i] = MockUser()
	}
	return l
}
