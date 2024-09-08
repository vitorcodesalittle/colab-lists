package list

import (
	"errors"
)

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

func AddItemToGroup(groups []Group, item Item, groupName string) {
	for i, group := range groups {
		if group.Name == groupName {
			groups[i].Items = append(group.Items, item)
			return
		}
	}
	group := Group{
		Name:  groupName,
		Items: []Item{item},
	}
	groups = append(groups, group)
}

func UpdateItemToGroup(groups []Group, item Item, groupName string) {
	for i, group := range groups {
		if group.Name == groupName {
			for j, itemGroup := range group.Items {
				if itemGroup.Id == item.Id {
					groups[i].Items[j] = item
					return
				}
			}
		}
	}
}

func DeleteItemFromGroups(groups []Group, itemId int) {
	for i, group := range groups {
		for j, itemGroup := range group.Items {
			if itemGroup.Id == itemId {
				groups[i].Items = append(group.Items[:j], group.Items[j+1:]...)
				return
			}
		}
	}
}

func (r *ListsInMemoryRepository) Update(params *ListUpdateParams) (List, error) {
	for i, list := range r.lists {
		if list.Id == params.Id {
			r.lists[i].Title = params.Title
			r.lists[i].Description = params.Description

			for _, itemParams := range params.Items {
				item := Item{
					Id:          itemParams.Id,
					Order:       itemParams.Order,
					Description: itemParams.Description,
					Quantity:    itemParams.Quantity,
				}
				if itemParams.Operation == Add {
					AddItemToGroup(r.lists[i].Groups, item, itemParams.GroupName)
				} else if itemParams.Operation == Update {
					UpdateItemToGroup(r.lists[i].Groups, item, itemParams.GroupName)
				} else if itemParams.Operation == Remove {
					DeleteItemFromGroups(r.lists[i].Groups, itemParams.Id)
				} else {
					return List{}, errors.New("Invalid operation")
				}
			}

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
