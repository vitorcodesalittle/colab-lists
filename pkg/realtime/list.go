package realtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

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
	ACTION_EDIT_GROU    = iota
	ACTION_ADD_ITEM     = iota
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

func (l *LiveEditor) HandleAddGroup(listId int64, groupText string) {
	editList := l.GetCurrentList(listId)
	if editList == nil {
		return
	}
	editList.Groups = append(editList.Groups, list.Group{Name: groupText, Items: []list.Item{}})
	for _, conn := range l.GetConnectionsOfList(listId) {
		conn.Conn.WriteJSON(editList)
	}
}

func (l *LiveEditor) HandleEditGroupItem(listId int64, groupIndex int, groupText string) {
	editList := l.GetCurrentList(listId)
	if editList == nil {
		return
	}
	group := editList.Groups[groupIndex]
	group.Name = groupText
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
				l.HandleFocusItem(focusItemAction, conn)
			case ACTION_UNFOCUS_ITEM:
				var unfocusItemAction UnfocusItemAction
				if err := json.Unmarshal(p, &unfocusItemAction); err != nil {
					log.Println("Error unmarshalling action", err)
					continue
				}
				l.HandleUnfocusItem(unfocusItemAction, conn)
			case ACTION_UPDATE_COLOR:
				var updateColorAction UpdateColorAction
				if err := json.Unmarshal(p, &updateColorAction); err != nil {
					log.Println("Error unmarshalling action", err)
					continue
				}
				l.HandleUpdateColor(updateColorAction, conn)
			}
		case ACTION_ADD_GROUP:
			l.HandleAddGroup(conn.ListId, "New Group")
		}
	}
}

func (l *LiveEditor) SetupList(listId int64, user user.User, conn *websocket.Conn) {
	defer l.Info()
	listUi, ok := l.listsById[listId]
	conn2 := &Connection{ListId: listId, UserId: user.Id, Conn: conn}
	if !ok {
		list, err := l.listRepository.Get(listId)
		panicIfError(err)
		l.listsById[listId] = &ListState{ui: &views.ListUi{
			List:               list,
			ColaboratorsOnline: []views.UserUi{{User: user, Color: "#18d825"}},
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
			listUi.ui.ColaboratorsOnline = append(listUi.ui.ColaboratorsOnline, views.UserUi{User: user, Color: "#18d825"})
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

func (l *LiveEditor) HandleFocusItem(action FocusItemAction, conn *Connection) {
	log.Println("Handling focus Item")
	list, ok := l.listsById[conn.ListId]
	if !ok {
		return
	}
	item := list.ui.List.Groups[action.GroupIndex].Items[action.ItemIndex]
	userUi := &views.UserUi{Color: "green"}
	for _, c := range list.ui.ColaboratorsOnline {
		if c.User.Id == conn.UserId {
			userUi = &c
		}
	}
	args := views.IndexedItem{
		GroupIndex: action.GroupIndex,
		ItemIndex:  action.ItemIndex,
		Item:       item,
		Color:      userUi.Color,
		ActionType: ACTION_FOCUS_ITEM,
	}
	conns := l.GetConnectionsOfList(conn.ListId)
	for _, conn2 := range conns {
		conn2.Conn.WriteJSON(&args)
	}
}

func (l *LiveEditor) HandleUnfocusItem(action UnfocusItemAction, conn *Connection) {
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

func (l *LiveEditor) HandleUpdateColor(action UpdateColorAction, conn *Connection) {
	log.Println("Handling update color")
	listUi, ok := l.listsById[conn.ListId]
	if !ok {
		return
	}
	for i, userUi := range listUi.ui.ColaboratorsOnline {
		if userUi.Id == action.UserId {
			listUi.ui.ColaboratorsOnline[i].Color = action.Color
		}
	}
	s := ""
	buf := bytes.NewBufferString(s)
	views.Templates.RenderCollaboratorsList(buf, listUi.ui.ColaboratorsOnline)
	conns := l.GetConnectionsOfList(conn.ListId)
	for _, conn2 := range conns {
		conn2.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l *LiveEditor) HandleAddItem(listId int64, groupIndex int, itemText string) {
	editList := l.GetCurrentList(listId)
	items := editList.Groups[groupIndex].Items
	if groupIndex < 0 || groupIndex >= len(editList.Groups) {
		return
	}
	editList.Groups[groupIndex].Items = append(items, list.Item{Id: 0, Description: itemText, Order: len(items)})
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
	Color  string `json:"color"`
	UserId int64  `json:"userId"`
}

type BlurItemAction struct {
	GroupIndex int `json:"groupIndex"`
	ItemIndex  int `json:"itemIndex"`
}

type AddItemAction struct {
	GroupText string `json:"groupText"`
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
