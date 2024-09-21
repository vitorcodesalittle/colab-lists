package realtime

import (
	"log"
	"strconv"

	"vilmasoftware.com/colablists/pkg/list"
	"vilmasoftware.com/colablists/pkg/user"
	"vilmasoftware.com/colablists/pkg/views"
)

type ListState struct {
	Ui               *views.ListUi
	connections      []*connection
	Dirty            bool
	groupIdGenerator *Generator
	itemIdGenerator  *Generator
}

func NewListState(list *list.List, user *user.User, conn *connection) *ListState {
    groupIds := make([]int64, 0)
    itemIds := make([]int64, 0)
    for _, group := range list.Groups {
        groupIds = append(groupIds, group.GroupId)
        for _, item := range group.Items {
            itemIds = append(itemIds, item.Id)
        }
    }
	return &ListState{
		Ui: &views.ListUi{
			List:               list,
			ColaboratorsOnline: []*views.UserUi{{User: user, Color: "#18d825"}},
		},
		connections:      []*connection{conn},
		groupIdGenerator: NewGenerator(maxSlice(groupIds)+1),
		itemIdGenerator:  NewGenerator(maxSlice(itemIds)+1),
	}
}

func (ls *ListState)FindGroupById(groupId int64) *list.Group {
    for _, group := range ls.Ui.List.Groups {
        if group.GroupId == groupId {
            return group
        }
    }
    return nil
}

func (ls *ListState)FindItemById(groupId, itemId int64) *list.Item {
    group := ls.FindGroupById(groupId)
    if group == nil {
        return nil
    }
    for _, item := range group.Items {
        if item.Id == itemId {
            return item
        }
    }
    return nil
}

func maxSlice(arr []int64) int64 {
    result := int64(1)
    for _, v := range arr {
        if v > result {
            result = v
        }
    }
    return result
}

func (ls *ListState) AddGroup(groupText string) *list.Group {
    groupId := ls.groupIdGenerator.Next()
    group :=&list.Group{GroupId: groupId, Name: groupText, Items: []*list.Item{{
		Order:       groupId,
		GroupId:     groupId,
		Description: "New Item",
		Id:          ls.itemIdGenerator.Next(),
		Quantity:    1,
	}}}
	ls.Ui.List.Groups = append(ls.Ui.List.Groups, group)
    ls.Dirty = true
    return group
}

func (ls *ListState) AddItem(groupId int64, itemText string) *list.Item {
    group := ls.FindGroupById(groupId)
    if group == nil {
        return nil
    }
    itemId := ls.itemIdGenerator.Next()
    item := &list.Item{
		Order:       itemId,
		GroupId:     groupId,
		Description: "New Item",
		Id:          itemId,
		Quantity:    1,
	}
    group.Items = append(group.Items, item)
    ls.Dirty = true
    return item 
}

func (ls *ListState) DeleteGroup(groupId int64) {
    for i, group := range ls.Ui.List.Groups {
        if group.GroupId == groupId {
            ls.Ui.List.Groups = append(ls.Ui.List.Groups[:i], ls.Ui.List.Groups[i+1:]...)
            ls.Dirty = true
            return
        }
    }
}


func (ls *ListState)EditGroup (args *EditGroupAction) *list.Group {
    group := ls.FindGroupById(args.GroupIndex)
    if group == nil {
        return nil
    }
    group.Name = args.Text
    ls.Dirty = true
    return group
}

func (ls *ListState) EditItem(args *EditItemArgs) *list.Item {
    item := ls.FindItemById(args.GroupIndex, args.ItemIndex)
    if item == nil {
        return nil
    }

	qtd, err := strconv.Atoi(args.Quantity)
	if err != nil {
        log.Printf("Error converting quantity to int: %v\n", err)
        return nil
	}
	item.Description = args.Description
	item.Quantity = qtd
	ls.Dirty = true
    return item
}

func (ls *ListState) DeleteItem(groupId, itemId int64) {
    group := ls.FindGroupById(groupId)
    if group == nil {
        return
    }
    for i, item := range group.Items {
        if item.Id == itemId {
            group.Items = append(group.Items[:i], group.Items[i+1:]...)
            ls.Dirty = true
            return
        }
    }
    ls.Dirty = true
}
