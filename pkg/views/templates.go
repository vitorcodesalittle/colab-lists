package views

import (
	"fmt"
	"io"
	textTemplate "text/template"

	"vilmasoftware.com/colablists/pkg/list"
	"vilmasoftware.com/colablists/pkg/user"
)
var Templates *templates = newTemplates()

type templates struct {
	Index  *textTemplate.Template
	List   *textTemplate.Template
	Lists  *textTemplate.Template
	Auth   *textTemplate.Template
	Signup *textTemplate.Template
	Base   *textTemplate.Template
}

type ListUi struct {
	*list.List
	ColaboratorsOnline []*UserUi
	// Try not to use this
	// focusMap map[int64]map[int]int
}

type UserUi struct {
	*user.User
	Color string
	//*Action
}


type IndexArgs struct {
	Title       string
	Description string
}


type Colaborator struct {
	user.User
}
type ColaboratorsListArgs = []Colaborator

func (t *templates) RenderCollaboratorsList(w io.Writer, args []*UserUi) {
	t.List.ExecuteTemplate(w, "colaborators", args)
}

//type ItemArgs struct {
//	IndexedItem
//	IsAdding bool
//}

type ItemArgs = IndexedItem

func (t *templates) RenderItem(w io.Writer, args ItemArgs) {
    err := t.List.ExecuteTemplate(w, "item", args)
    if err != nil {
        panic (err)
    }
}

type GroupArgs = IndexedGroup

func (t *templates) RenderGroup(w io.Writer, args GroupArgs) {
    err := t.List.ExecuteTemplate(w, "group", args)
    if err != nil {
        panic(err)
    }
}

func (t *templates) RenderSaveList(w io.Writer, args *ListArgs) {
    err := t.List.ExecuteTemplate(w, "save", args)
    if err != nil {
        panic(err)
    }
}

type IndexedItem struct {
	ActionType int       `json:"actionType"`
	GroupIndex int64       `json:"groupIndex"`
	ItemIndex  int64       `json:"itemIndex"`
	Item       *list.Item `json:"item"`
	Color      string    `json:"color"`
	HxSwapOob  string
}
type IndexedGroup struct {
	GroupIndex int64        `json:"groupIndex"`
	Group      *list.Group `json:"group"`
	Id         string
	HxSwapOob  string
}

func NewGroupIndex(groupIndex int64, group *list.Group, hxSwapOob string) *IndexedGroup {
	return &IndexedGroup{GroupIndex: groupIndex, Group: group, Id: fmt.Sprintf("group-%d", groupIndex), HxSwapOob: hxSwapOob}
}

func NewIndexedItem(groupIndex int64, itemIndex int64, item *list.Item, color string, hxSwapOob string) *IndexedItem {
	return &IndexedItem{GroupIndex: groupIndex, ItemIndex: itemIndex, Item: item, Color: color, HxSwapOob: hxSwapOob}
}

func newTemplates() *templates {
	templates := &templates{}
	templates.Base = textTemplate.Must(textTemplate.ParseFiles("./templates/pages/_base.html"))
	templates.Index = textTemplate.Must(textTemplate.ParseFiles("./templates/pages/index.html"))
	templates.Auth = textTemplate.Must(textTemplate.ParseFiles("./templates/pages/auth.html", "./templates/pages/_base.html"))
	templates.Lists = textTemplate.Must(textTemplate.ParseFiles("./templates/pages/lists.html", "./templates/pages/_base.html"))
	templates.List = textTemplate.Must(textTemplate.New("list.html").Funcs(textTemplate.FuncMap{
		"indexeditem": func(groupIndex int64, itemIndex int64, item *list.Item, color string) *IndexedItem {
			return NewIndexedItem(groupIndex, itemIndex, item, color, "")
		},
		"indexedgroup": func(groupIndex int64, group *list.Group) *IndexedGroup {
			return NewGroupIndex(groupIndex, group, "")
		},
	}).ParseFiles("./templates/pages/list.html", "./templates/pages/_base.html"))
	return templates
}
