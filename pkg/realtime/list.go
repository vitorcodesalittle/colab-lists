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


type Connection struct {
	ListId int64
	UserId int64
	Conn   *websocket.Conn
}

func (c *Connection) String() string {
	return fmt.Sprintf("Connection{ListId: %d, UserId: %d, Conn: %v}", c.ListId, c.UserId, c.Conn)
}

type ListState struct {
    ui *views.ListUi
    connections []*Connection
}

type LiveEditor struct {
	listsById           map[int64]*ListState
	listRepository      list.ListsRepository
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

func (l *LiveEditor) HandleWebsocketConn(user user.User, conn *websocket.Conn) {
	conn.WriteMessage(websocket.TextMessage, []byte("Hello"))
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				l.removeConnection(conn)
				log.Println("Unexcepted Close Error: ", err)
				return
			} else if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				l.removeConnection(conn)
				log.Println("Close Error: ", err)
				return
			}
			log.Println("Unexpected error reading websocket message ", err.Error())
			continue
		}
		switch messageType {
		case websocket.CloseMessage:
			log.Println("CloseMessage")
			l.removeConnection(conn)
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
			case FOCUS_ITEM:
				var focusItemAction FocusItemAction
				if err := json.Unmarshal(p, &focusItemAction); err != nil {
					log.Println("Error unmarshalling action", err)
					continue
				}
				l.HandleFocusItem(focusItemAction.ListId, user.Id, focusItemAction.GroupIndex, focusItemAction.ItemIndex)
			case UNFOCUS_ITEM:
				var focusItemAction UnfocusItemAction
				if err := json.Unmarshal(p, &focusItemAction); err != nil {
					log.Println("Error unmarshalling action", err)
					continue
				}
				l.HandleUnfocusItem(focusItemAction.ListId, user.Id, focusItemAction.GroupIndex, focusItemAction.ItemIndex)
			case ADD_GROUP_ACTION:
				var addGroupAction AddItemAction
				if err := json.Unmarshal(p, &addGroupAction); err != nil {
					log.Fatal(err)
				}
				l.HandleAddGroup(addGroupAction.ListId, addGroupAction.GroupText)
			}
			continue
		}
	}
}

func (l *LiveEditor) SetupList(listId int64, user user.User, conn *websocket.Conn) {
	defer l.Info()
	listUi, ok := l.listsById[listId]
	if !ok {
		list, err := l.listRepository.Get(listId)
		panicIfError(err)
        l.listsById[listId] = &ListState{ui:&views.ListUi{
			List:               list,
			ColaboratorsOnline: []views.UserUi{{User: user}},
		}, connections: []*Connection{{ListId: listId, UserId: user.Id, Conn: conn}}}
		listUi = l.listsById[listId]
	} else {
        listUi.connections = append(listUi.connections, &Connection{ListId: listId, UserId: user.Id, Conn: conn})
        l.listsById[listId] = listUi
	}
	s := ""
	buf := bytes.NewBufferString(s)
	views.Templates.RenderCollaboratorsList(buf, listUi.ui.ColaboratorsOnline)
	conns := l.GetConnectionsOfList(listId)
	for _, conn2 := range conns {
		conn2.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
	go l.HandleWebsocketConn(user, conn)
}

const (
	ADD_GROUP    = iota
	FOCUS_ITEM   = iota
	UNFOCUS_ITEM = iota
	EDIT_GROU    = iota
	ADD_ITEM     = iota
)

func (l *LiveEditor) HandleFocusItem(listId int64, userId int64, groupIndex int, itemIndex int) {
	log.Println("Handling focus Item")
	s := "TODO HANDLE FOCUS"
	buf := bytes.NewBufferString(s)
	conns := l.GetConnectionsOfList(listId)
	for _, conn2 := range conns {
		conn2.Conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
}

func (l *LiveEditor) HandleUnfocusItem(listId int64, userId int64, groupIndex int, itemIndex int) {
	log.Println("Handling unfocus Item")
	s := "TODO: HANDLE UNFOCUS"
	buf := bytes.NewBufferString(s)
	conns := l.GetConnectionsOfList(listId)
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

const (
	ADD_GROUP_ACTION = iota
)

type Action struct {
	Type *int         `json:"action-type"`
	Msg  interface{}
}

type FocusItemAction struct {
	ListId     int64 `json:"listId"`
	GroupIndex int   `json:"groupIndex"`
	ItemIndex  int   `json:"itemIndex"`
}

type UnfocusItemAction struct {
	ListId     int64
	GroupIndex int
	ItemIndex  int
}

type BlurItemAction struct {
	ListId     int64 `json:"listId"`
	GroupIndex int   `json:"groupIndex"`
	ItemIndex  int   `json:"itemIndex"`
}

type AddItemAction struct {
	ListId    int64  `json:"listId"`
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
