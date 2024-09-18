package realtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gorilla/websocket"
	"vilmasoftware.com/colablists/pkg/list"
	"vilmasoftware.com/colablists/pkg/user"
	"vilmasoftware.com/colablists/pkg/views"
)

// Must match actionType in html
const (
	ACTION_NOOP         = iota
	ACTION_FOCUS_ITEM   = iota
	ACTION_UNFOCUS_ITEM = iota
	ACTION_UPDATE_COLOR = iota
	ACTION_ADD_GROUP    = iota
	ACTION_EDIT_GROUP   = iota
	ACTION_ADD_ITEM     = iota
	ACTION_DELETE_GROUP = iota
	ACTION_DELETE_ITEM  = iota
	ACTION_EDIT_ITEM    = iota
)

type Connection struct {
	ListId int64
	UserId int64
	Conn   *websocket.Conn
}

func (c *Connection) String() string {
	return fmt.Sprintf("Connection{ListId: %d, UserId: %d, Conn: %v}", c.ListId, c.UserId, c.Conn)
}

type ListState struct {
	ui          *views.ListUi
	connections []*Connection
}

type LiveEditor struct {
	listsById      map[int64]*ListState
	listRepository list.ListsRepository
}

func (l *LiveEditor) Info() {
	println("Live editor info")
	for k, v := range l.listsById {
		println("List ", k, " has ", len(v.ui.ColaboratorsOnline), " colaborators")
		println("List ", k, " has ", len(v.connections), " connections")
	}
}

func NewLiveEditor(repository list.ListsRepository) *LiveEditor {
	return &LiveEditor{
		listRepository: repository,
		listsById:      make(map[int64]*ListState),
	}
}

func (l *LiveEditor) GetConnectionsOfList(listId int64) []*Connection {
	conns, ok := l.listsById[listId]
	if !ok {
		return []*Connection{}
	}
	return conns.connections
}

func (l *LiveEditor) removeConnection(conn *websocket.Conn) {
	for k, v := range l.listsById {
		connections := make([]*Connection, 0)
		for _, c := range v.connections {
			if c.Conn != conn {
				connections = append(connections, c)
			}
		}
		l.listsById[k] = &ListState{ui: v.ui, connections: connections}
	}
}

func (l *LiveEditor) HandleWebsocketConn(conn *Connection) {
	conn.Conn.WriteMessage(websocket.TextMessage, []byte("Hello"))
	for {
		messageType, p, err := conn.Conn.ReadMessage()
		if err != nil {

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				l.removeConnection(conn.Conn)
				log.Println("Unexcepted Close Error: ", err)
				return
			} else if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				l.removeConnection(conn.Conn)
				log.Println("Close Error: ", err)
				return
			}
			log.Println("Unexpected error reading websocket message ", err.Error())
			continue
		}
		switch messageType {
		case websocket.CloseMessage:
			log.Println("CloseMessage")
			l.removeConnection(conn.Conn)
			continue
		case websocket.PingMessage:
			log.Println("PingMessage")
			continue
		case websocket.PongMessage:
			log.Println("PongMessage")
			continue
		case websocket.BinaryMessage:
			log.Println("BinaryMessage")
			continue
		case websocket.TextMessage:
			log.Println("TextMessage")
			var msg json.RawMessage
			action := Action{Msg: &msg}
			if err := json.Unmarshal(p, &action); err != nil {
				log.Println("Error unmarshalling message ", err)
				continue
			}
			if action.Type == nil {
				log.Println("ActionType is nil")
				continue
			}
			switch *action.Type {
			case ACTION_FOCUS_ITEM:
				var focusItemAction FocusItemAction
				if err := json.Unmarshal(p, &focusItemAction); err != nil {
					log.Println("Error unmarshalling action", err)
					continue
				}
				l.HandleFocusItem(&focusItemAction, conn)
			case ACTION_UNFOCUS_ITEM:
				var unfocusItemAction UnfocusItemAction
				if err := json.Unmarshal(p, &unfocusItemAction); err != nil {
					log.Println("Error unmarshalling action", err)
					continue
				}
				l.HandleUnfocusItem(&unfocusItemAction, conn)
			case ACTION_UPDATE_COLOR:
				var updateColorAction UpdateColorAction
				if err := json.Unmarshal(p, &updateColorAction); err != nil {
					log.Println("Error unmarshalling action", err)
					continue
				}
				l.HandleUpdateColor(&updateColorAction, conn)
			case ACTION_ADD_GROUP:
				l.HandleAddGroup(conn.ListId, "New Group")
			case ACTION_ADD_ITEM:
				var addItemAction AddItemAction
				if err := json.Unmarshal(p, &addItemAction); err != nil {
					log.Println("Error unmarshalling action", err)
					continue
				}
				l.HandleAddItem(&addItemAction, conn)
			case ACTION_EDIT_GROUP:
				var editGroupAction EditGroupAction
				if err := json.Unmarshal(p, &editGroupAction); err != nil {
					log.Println("Error unmarshalling action", err)
					continue
				}
				l.HandleEditGroup(&editGroupAction, conn)
			case ACTION_DELETE_GROUP:
				var deleteGroupAction DeleteGroupArgs
				if err := json.Unmarshal(p, &deleteGroupAction); err != nil {
					log.Println("Error unmarshalling action", err)
					continue
				}
				l.HandleDeleteGroup(&deleteGroupAction, conn)
			case ACTION_DELETE_ITEM:
				var deleteItemArgs DeleteItemArgs
				if err := json.Unmarshal(p, &deleteItemArgs); err != nil {
					log.Println("Error unmarshalling action", err)
					continue
				}
				l.HandleDeleteItem(&deleteItemArgs, conn)
			case ACTION_EDIT_ITEM:
				var editItemArgs EditItemArgs
				if err := json.Unmarshal(p, &editItemArgs); err != nil {
					log.Println("Error unmarshalling action", err)
					continue
				}
                l.HandleEditItem(&editItemArgs, conn)
			}

		}
	}
}

func (l *LiveEditor) SetupList(listId int64, user *user.User, conn *websocket.Conn) {
	defer l.Info()
	listUi, ok := l.listsById[listId]
	conn2 := &Connection{ListId: listId, UserId: user.Id, Conn: conn}
	if !ok {
		list, err := l.listRepository.Get(listId)
		panicIfError(err)
		l.listsById[listId] = &ListState{ui: &views.ListUi{
			List:               list,
			ColaboratorsOnline: []*views.UserUi{{User: user, Color: "#18d825"}},
		}, connections: []*Connection{conn2}}
		listUi = l.listsById[listId]
	} else {
		listUi.connections = append(listUi.connections, conn2)
		found := false
		for _, userUi := range listUi.ui.ColaboratorsOnline {
			if userUi.Id == user.Id {
				found = true
				break
			}
		}
		if !found {
			listUi.ui.ColaboratorsOnline = append(listUi.ui.ColaboratorsOnline, &views.UserUi{User: user, Color: "#18d825"})
		}
		l.listsById[listId] = listUi
	}
	s := ""
	buf := bytes.NewBufferString(s)
	views.Templates.RenderCollaboratorsList(buf, listUi.ui.ColaboratorsOnline)
	conns := l.GetConnectionsOfList(listId)
	for _, conn2 := range conns {
		conn2.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
	go l.HandleWebsocketConn(conn2)
}

func (l *LiveEditor) HandleFocusItem(action *FocusItemAction, conn *Connection) {
	log.Println("Handling focus Item")
	list, ok := l.listsById[conn.ListId]
	if !ok {
		return
	}
	item := list.ui.List.Groups[action.GroupIndex].Items[action.ItemIndex]
	args := views.IndexedItem{
		GroupIndex: action.GroupIndex,
		ItemIndex:  action.ItemIndex,
		Item:       item,
		Color:      l.GetColaboratorOnline(conn.ListId, conn.UserId).Color,
		ActionType: ACTION_FOCUS_ITEM,
	}
	conns := l.GetConnectionsOfList(conn.ListId)
	for _, conn2 := range conns {
		conn2.Conn.WriteJSON(&args)
	}
}

func (l *LiveEditor) HandleUnfocusItem(action *UnfocusItemAction, conn *Connection) {
	log.Println("Handling unfocus Item")
	list, ok := l.listsById[conn.ListId]
	if !ok {
		return
	}
	item := list.ui.List.Groups[action.GroupIndex].Items[action.ItemIndex]
	args := views.IndexedItem{
		GroupIndex: action.GroupIndex,
		ItemIndex:  action.ItemIndex,
		Item:       item,
		Color:      "",
		ActionType: ACTION_UNFOCUS_ITEM,
	}
	conns := l.GetConnectionsOfList(conn.ListId)
	for _, conn2 := range conns {
		conn2.Conn.WriteJSON(&args)
	}
}

func (l *LiveEditor) GetColaboratorOnline(listId int64, userId int64) *views.UserUi {
	list, ok := l.listsById[listId]
	if !ok {
		return nil
	}
	for _, userUi := range list.ui.ColaboratorsOnline {
		if userUi.Id == userId {
			return userUi
		}
	}
	return nil
}

func (l *LiveEditor) HandleUpdateColor(action *UpdateColorAction, conn *Connection) {
	log.Println("Handling update color")
	listUi, ok := l.listsById[conn.ListId]
	if !ok {
		return
	}
	l.GetColaboratorOnline(conn.ListId, action.UserId).Color = action.Color

	s := ""
	buf := bytes.NewBufferString(s)
	views.Templates.RenderCollaboratorsList(buf, listUi.ui.ColaboratorsOnline)
	conns := l.GetConnectionsOfList(conn.ListId)
	for _, conn2 := range conns {
		conn2.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l *LiveEditor) HandleAddGroup(listId int64, groupText string) {
	editList := l.GetCurrentList(listId)
	if editList == nil {
		return
	}
	groupIndex := len(editList.Groups)
	g := list.Group{Name: groupText, Items: []list.Item{{
		Order:       0,
		GroupId:     groupIndex,
		Description: "New Item",
		Id:          0,
		Quantity:    1,
	}}}
	editList.Groups = append(editList.Groups, g)
	s := ""
	buf := bytes.NewBufferString(s)
	views.Templates.RenderGroup(buf, *views.NewGroupIndex(groupIndex, &g, "beforeend:#groups"))
	for _, conn := range l.GetConnectionsOfList(listId) {
		conn.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l *LiveEditor) HandleEditGroup(action *EditGroupAction, conn *Connection) {
	editList := l.GetCurrentList(conn.ListId)
	if editList == nil {
		return
	}
	editList.Groups[action.GroupIndex].Name = action.Text
	s := ""
	buf := bytes.NewBufferString(s)
	gi := *views.NewGroupIndex(action.GroupIndex, &editList.Groups[action.GroupIndex], "outerHTML:")
	gi.HxSwapOob = "outerHTML:#" + gi.Id
	views.Templates.RenderGroup(buf, gi)

	for _, conn := range l.GetConnectionsOfList(conn.ListId) {
		conn.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}


func (l *LiveEditor) HandleAddItem(args *AddItemAction, conn *Connection) {
	println("HandleAddItem")
	editList := l.GetCurrentList(conn.ListId)
	items := editList.Groups[args.GroupIndex].Items
	if args.GroupIndex < 0 || args.GroupIndex >= len(editList.Groups) {
		return
	}
	editList.Groups[args.GroupIndex].Items = append(items, list.Item{Id: 0, Description: "New item", Order: len(items)})

	s := ""
	buf := bytes.NewBufferString(s)
	color := l.GetColaboratorOnline(conn.ListId, conn.UserId).Color

	views.Templates.RenderItem(buf, *views.NewIndexedItem(args.GroupIndex, len(items), &editList.Groups[args.GroupIndex].Items[len(items)-1], color, "beforeend:#items-"+strconv.Itoa(args.GroupIndex)))
	for _, conn := range l.GetConnectionsOfList(conn.ListId) {
		conn.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l *LiveEditor) HandleDeleteGroup(args *DeleteGroupArgs, conn *Connection) {
	editList := l.GetCurrentList(conn.ListId)
	if args.GroupIndex < 0 || args.GroupIndex >= len(editList.Groups) {
		return
	}
	group := editList.Groups[args.GroupIndex]
	editList.Groups = append(editList.Groups[:args.GroupIndex], editList.Groups[args.GroupIndex+1:]...)
	s := ""
	buf := bytes.NewBufferString(s)
	g := *views.NewGroupIndex(args.GroupIndex, &group, "delete")
	g.HxSwapOob = "delete:#" + g.Id
	views.Templates.RenderGroup(buf, g)
	for _, conn := range l.GetConnectionsOfList(conn.ListId) {
		conn.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l *LiveEditor) HandleDeleteItem(args *DeleteItemArgs, conn *Connection) {
	editList := l.GetCurrentList(conn.ListId)
	if args.GroupIndex < 0 || args.GroupIndex >= len(editList.Groups) {
		return
	}
	group := editList.Groups[args.GroupIndex]
	if args.ItemIndex < 0 || args.ItemIndex >= len(group.Items) {
		return
	}
	item := group.Items[args.ItemIndex]
	group.Items = append(group.Items[:args.ItemIndex], group.Items[args.ItemIndex+1:]...)
	s := ""
	buf := bytes.NewBufferString(s)
	color := l.GetColaboratorOnline(conn.ListId, conn.UserId).Color
	i := *views.NewIndexedItem(args.GroupIndex, args.ItemIndex, &item, color, fmt.Sprintf("delete:#desc-%d-%d", args.GroupIndex, args.ItemIndex))
	views.Templates.RenderItem(buf, i)
	for _, conn := range l.GetConnectionsOfList(conn.ListId) {
		conn.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l* LiveEditor) HandleEditItem(args *EditItemArgs, conn *Connection) {
    println("Handling edit item")
    editList := l.GetCurrentList(conn.ListId)
    if args.GroupIndex < 0 || args.GroupIndex >= len(editList.Groups) {
        return
    }
    group := editList.Groups[args.GroupIndex]
    if args.ItemIndex < 0 || args.ItemIndex >= len(group.Items) {
        return
    }
    qtd, err := strconv.Atoi(args.Quantity)
    if err != nil {
        log.Println("Error converting quantity to int ", err)
        return
    }
    item := group.Items[args.ItemIndex]
    item.Description = args.Description
    item.Quantity = qtd
    println("Description : " + item.Description)
    println("Quantity : " + strconv.Itoa(item.Quantity))
    s := ""
    buf := bytes.NewBufferString(s)
    color := l.GetColaboratorOnline(conn.ListId, conn.UserId).Color
    i := *views.NewIndexedItem(args.GroupIndex, args.ItemIndex, &item, color, fmt.Sprintf("outerHTML:#desc-%d-%d", args.GroupIndex, args.ItemIndex))
    views.Templates.RenderItem(buf, i)
    for _, conn := range l.GetConnectionsOfList(conn.ListId) {
        conn.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
    }
}

type DeleteItemArgs struct {
	GroupIndex int `json:"groupIndex"`
	ItemIndex  int `json:"itemIndex"`
}

func (l *LiveEditor) GetCurrentList(listId int64) *views.ListUi {
	list, ok := l.listsById[listId]
	if ok {
		return list.ui
	}
	return nil
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

type Action struct {
	Type *int `json:"actionType"`
	Msg  interface{}
}

type FocusItemAction struct {
	GroupIndex int `json:"groupIndex"`
	ItemIndex  int `json:"itemIndex"`
}

type UnfocusItemAction struct {
	GroupIndex int
	ItemIndex  int
}

type UpdateColorAction struct {
	Color      string `json:"color"`
	UserId     int64  `json:"userId"`
	GroupIndex int    `json:"groupIndex"`
}

type EditGroupAction struct {
	GroupIndex int    `json:"groupIndex"`
	Text       string `json:"text"`
}

type BlurItemAction struct {
	GroupIndex int `json:"groupIndex"`
	ItemIndex  int `json:"itemIndex"`
}

type AddItemAction struct {
	GroupIndex int `json:"groupIndex"`
}

type EditItemAction struct {
	GroupIndex int `json:"groupIndex"`
}

type DeleteGroupArgs struct {
	GroupIndex int `json:"groupIndex"`
}

type EditItemArgs struct {
	GroupIndex  int    `json:"groupIndex"`
	ItemIndex   int    `json:"itemIndex"`
	Description string `json:"description"`
	Quantity    string `json:"quantity"`
}

// type AddGroupAction struct {
// 	Action
// 	GroupIndex int
// 	GroupText  string
// }
//
// type FocusItemAction struct {
// 	Action
// 	GroupIndex int
// 	ItemIndex  int
// }
//
