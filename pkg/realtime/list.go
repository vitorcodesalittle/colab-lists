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

type connection struct {
	ListId int64
	UserId int64
	Conn   *websocket.Conn
}

func (c *connection) String() string {
	return fmt.Sprintf("Connection{ListId: %d, UserId: %d, Conn: %v}", c.ListId, c.UserId, c.Conn)
}


type LiveEditor struct {
	listsById      map[int64]*ListState
	listRepository list.ListsRepository
}

func (l *LiveEditor) Info() {
	println("Live editor info")
	for k, v := range l.listsById {
		println("List ", k, " has ", len(v.Ui.ColaboratorsOnline), " colaborators")
		println("List ", k, " has ", len(v.connections), " connections")
	}
}


func NewLiveEditor(repository list.ListsRepository) *LiveEditor {
	return &LiveEditor{
		listRepository: repository,
		listsById:      make(map[int64]*ListState),
	}
}

func (l *LiveEditor) GetConnectionsOfList(listId int64) []*connection {
	conns, ok := l.listsById[listId]
	if !ok {
		return []*connection{}
	}
	return conns.connections
}

func (l *LiveEditor) removeConnection(conn *websocket.Conn) {
	for k, v := range l.listsById {
		connections := make([]*connection, 0)
		for _, c := range v.connections {
			if c.Conn != conn {
				connections = append(connections, c)
			}
		}
		l.listsById[k].connections = connections
	}
}

func (l *LiveEditor) HandleWebsocketConn(conn *connection) {
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
			l.removeConnection(conn.Conn)
			continue
		case websocket.PingMessage:
			continue
		case websocket.PongMessage:
			continue
		case websocket.BinaryMessage:
			continue
		case websocket.TextMessage:
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
	conn2 := &connection{ListId: listId, UserId: user.Id, Conn: conn}
	if !ok {
		list, err := l.listRepository.Get(listId)
		assertNotNil(err)
		l.listsById[listId] = NewListState(&list, user, conn2)
		listUi = l.listsById[listId]
	} else {
		listUi.connections = append(listUi.connections, conn2)
		found := false
		for _, userUi := range listUi.Ui.ColaboratorsOnline {
			if userUi.Id == user.Id {
				found = true
				break
			}
		}
		if !found {
			listUi.Ui.ColaboratorsOnline = append(listUi.Ui.ColaboratorsOnline, &views.UserUi{User: user, Color: "#18d825"})
		}
		l.listsById[listId] = listUi
	}
	s := ""
	buf := bytes.NewBufferString(s)
	views.Templates.RenderCollaboratorsList(buf, listUi.Ui.ColaboratorsOnline)
	conns := l.GetConnectionsOfList(listId)
	for _, conn2 := range conns {
		conn2.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
	go l.HandleWebsocketConn(conn2)
}

func (l *LiveEditor) HandleFocusItem(action *FocusItemAction, conn *connection) {
	list, ok := l.listsById[conn.ListId]
	if !ok {
		return
	}
	item := list.FindItemById(int64(action.GroupIndex), int64(action.ItemIndex))
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

func (l *LiveEditor) HandleUnfocusItem(action *UnfocusItemAction, conn *connection) {
	list, ok := l.listsById[conn.ListId]
	if !ok {
		return
	}
	item := list.FindItemById(int64(action.GroupIndex), int64(action.ItemIndex))
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
	for _, userUi := range list.Ui.ColaboratorsOnline {
		if userUi.Id == userId {
			return userUi
		}
	}
	return nil
}

func (l *LiveEditor) HandleUpdateColor(action *UpdateColorAction, conn *connection) {
	listUi, ok := l.listsById[conn.ListId]
	if !ok {
		return
	}
	l.GetColaboratorOnline(conn.ListId, action.UserId).Color = action.Color

	s := ""
	buf := bytes.NewBufferString(s)
	views.Templates.RenderCollaboratorsList(buf, listUi.Ui.ColaboratorsOnline)
	conns := l.GetConnectionsOfList(conn.ListId)
	for _, conn2 := range conns {
		conn2.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l *LiveEditor) HandleAddGroup(listId int64, groupText string) {
    listState := l.GetCurrentListState(listId)
    if listState == nil {
        return
    }
    group := listState.AddGroup(groupText)
	editList := listState.Ui
	if editList == nil {
		return
	}
	s := ""
	buf := bytes.NewBufferString(s)
	views.Templates.RenderGroup(buf, *views.NewGroupIndex(group.GroupId, group, "beforeend:#groups"))
	views.Templates.RenderSaveList(buf, &views.ListArgs{List: *editList.List, IsDirty: true})
	for _, conn := range l.GetConnectionsOfList(listId) {
		conn.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l *LiveEditor) HandleEditGroup(action *EditGroupAction, conn *connection) {
    listState := l.GetCurrentListState(conn.ListId)
    if listState == nil {
        return
    }
    group := listState.EditGroup(action)
	editList := listState.Ui
	l.SetDirty(conn.ListId)
	s := ""
	buf := bytes.NewBufferString(s)
	gi := *views.NewGroupIndex(group.GroupId, group, "outerHTML:")
	gi.HxSwapOob = "outerHTML:#" + gi.Id
	views.Templates.RenderGroup(buf, gi)
	views.Templates.RenderSaveList(buf, &views.ListArgs{List: *editList.List, IsDirty: listState.Dirty})

	for _, conn := range l.GetConnectionsOfList(conn.ListId) {
		conn.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l *LiveEditor) HandleAddItem(args *AddItemAction, conn *connection) {
	listState := l.GetCurrentListState(conn.ListId)
    if listState == nil {
        return
    }
    item := listState.AddItem(int64(args.GroupIndex), "New Item")

    if item == nil {
        return
    }
	editList := listState.Ui
	s := ""
	buf := bytes.NewBufferString(s)
	color := l.GetColaboratorOnline(conn.ListId, conn.UserId).Color
    groupIdStr := strconv.FormatInt(int64(args.GroupIndex), 10)
	views.Templates.RenderItem(buf, *views.NewIndexedItem(item.GroupId, item.Id, item, color, "beforeend:#items-"+groupIdStr))
	views.Templates.RenderSaveList(buf, &views.ListArgs{List: *editList.List, IsDirty: true})
	for _, conn := range l.GetConnectionsOfList(conn.ListId) {
		conn.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l *LiveEditor) HandleDeleteGroup(args *DeleteGroupArgs, conn *connection) {
    listState := l.GetCurrentListState(conn.ListId)
    if listState == nil {
        return
    }
    listState.DeleteGroup(args.GroupIndex)
	s := ""
	buf := bytes.NewBufferString(s)
	g := *views.NewGroupIndex(args.GroupIndex, &list.Group{}, "delete")
	g.HxSwapOob = "delete:#" + g.Id
	views.Templates.RenderGroup(buf, g)
	views.Templates.RenderSaveList(buf, &views.ListArgs{List: *listState.Ui.List, IsDirty: true})
	for _, conn := range l.GetConnectionsOfList(conn.ListId) {
		conn.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l *LiveEditor) HandleDeleteItem(args *DeleteItemArgs, conn *connection) {
    listState := l.GetCurrentListState(conn.ListId)
    if listState == nil {
        return
    }
    listState.DeleteItem(args.GroupIndex, args.ItemIndex)
	editList := listState.Ui
	s := ""
	buf := bytes.NewBufferString(s)
	color := l.GetColaboratorOnline(conn.ListId, conn.UserId).Color
	i := *views.NewIndexedItem(args.GroupIndex, args.ItemIndex, &list.Item{}, color, fmt.Sprintf("delete:#desc-%d-%d", args.GroupIndex, args.ItemIndex))
	views.Templates.RenderItem(buf, i)
    views.Templates.RenderSaveList(buf, &views.ListArgs{List: *editList.List, IsDirty: true})
	for _, conn := range l.GetConnectionsOfList(conn.ListId) {
		conn.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l *LiveEditor) HandleEditItem(args *EditItemArgs, conn *connection) {
	listState := l.GetCurrentListState(conn.ListId)
    if listState == nil {
        return
    }
    item := listState.EditItem(args)
    if item == nil {
        return
    }
    println("Item edited: ", item)
	s := ""
	buf := bytes.NewBufferString(s)
	color := l.GetColaboratorOnline(conn.ListId, conn.UserId).Color
	i := *views.NewIndexedItem(args.GroupIndex, args.ItemIndex, item, color, fmt.Sprintf("outerHTML:#desc-%d-%d", args.GroupIndex, args.ItemIndex))
	views.Templates.RenderItem(buf, i)
    views.Templates.RenderSaveList(buf, &views.ListArgs{List: *listState.Ui.List, IsDirty: true})
	for _, conn := range l.GetConnectionsOfList(conn.ListId) {
		conn.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

type DeleteItemArgs struct {
	GroupIndex int64 `json:"groupIndex"`
	ItemIndex  int64 `json:"itemIndex"`
}

func (l *LiveEditor) GetCurrentListUi(listId int64) *views.ListUi {
	list, ok := l.listsById[listId]
	if ok {
		return list.Ui
	}
	return nil
}

func (l *LiveEditor) GetCurrentListState(listId int64) *ListState {
	list, ok := l.listsById[listId]
	if ok {
		return list
	}
	return nil
}

func (l *LiveEditor) SetDirty(listId int64) {
	list, ok := l.listsById[listId]
	if ok {
		list.Dirty = false
	}
}

func assertNotNil(err error) {
	if err != nil {
		panic(err)
	}
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
