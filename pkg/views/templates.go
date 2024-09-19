package views

import (
	"fmt"
	"io"
	textTemplate "text/template"

	"vilmasoftware.com/colablists/pkg/list"
	"vilmasoftware.com/colablists/pkg/user"
)

type templates struct {
	Index *textTemplate.Template
	List  *textTemplate.Template
	Lists *textTemplate.Template
	Login *textTemplate.Template
}

type IndexArgs struct {
	Title       string
	Description string
}

type ListsArgs struct {
	Lists []list.List
}
type ListArgs struct {
	List     list.List
	Editing  bool
	AllUsers []user.User
	IsDirty  bool
}

func (t *templates) RenderIndex(w io.Writer, args *IndexArgs) {
	t.Index.Execute(w, args)
}

func (t *templates) RenderList(w io.Writer, args *ListArgs) {
	t.List.Execute(w, args)
}

func (t *templates) RenderLists(w io.Writer, args *ListsArgs) {
	t.Lists.Execute(w, args)
}

type LoginArgs struct{}

func (t *templates) RenderLogin(w io.Writer, args *LoginArgs) {
	t.Login.Execute(w, args)
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
	t.List.ExecuteTemplate(w, "item", args)
}

type GroupArgs = IndexedGroup

func (t *templates) RenderGroup(w io.Writer, args GroupArgs) {
	t.List.ExecuteTemplate(w, "group", args)
}

func (t *templates) RenderSaveList(w io.Writer, args *ListArgs) {
	t.List.ExecuteTemplate(w, "save", args)
}

type IndexedItem struct {
	ActionType int       `json:"actionType"`
	GroupIndex int       `json:"groupIndex"`
	ItemIndex  int       `json:"itemIndex"`
	Item       list.Item `json:"item"`
	Color      string    `json:"color"`
	HxSwapOob  string
}
type IndexedGroup struct {
	GroupIndex int        `json:"groupIndex"`
	Group      list.Group `json:"group"`
	Id         string
	HxSwapOob  string
}

func NewGroupIndex(groupIndex int, group *list.Group, hxSwapOob string) *IndexedGroup {
	return &IndexedGroup{GroupIndex: groupIndex, Group: *group, Id: fmt.Sprintf("group-%d", groupIndex), HxSwapOob: hxSwapOob}
}

func NewIndexedItem(groupIndex int, itemIndex int, item *list.Item, color string, hxSwapOob string) *IndexedItem {
	return &IndexedItem{GroupIndex: groupIndex, ItemIndex: itemIndex, Item: *item, Color: color, HxSwapOob: hxSwapOob}
}

func newTemplates() *templates {
	templates := &templates{}
	templates.Index = textTemplate.Must(textTemplate.ParseFiles("./templates/pages/index.html"))
	templates.Login = textTemplate.Must(textTemplate.ParseFiles("./templates/pages/login.html"))
	templates.Lists = textTemplate.Must(textTemplate.ParseFiles("./templates/pages/lists.html"))
	templates.List = textTemplate.Must(textTemplate.New("list.html").Funcs(textTemplate.FuncMap{
		"indexeditem": func(groupIndex int, itemIndex int, item *list.Item, color string) *IndexedItem {
			return NewIndexedItem(groupIndex, itemIndex, item, color, "")
		},
		"indexedgroup": func(groupIndex int, group *list.Group) *IndexedGroup {
			return NewGroupIndex(groupIndex, group, "")
		},
	}).ParseFiles("./templates/pages/list.html", "./templates/pages/lists.html"))
	return templates
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

var Templates *templates = newTemplates()
