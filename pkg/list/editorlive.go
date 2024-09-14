package list

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"vilmasoftware.com/colablists/pkg/user"
)

type ListUi struct {
	List
	ColaboratorsOnline []UserUi
}

type LiveEditor struct {
	listsById      map[int64]ListUi
	listRepository ListsRepository
}

type UserUi struct {
	user.User
	Connections []*websocket.Conn
	//*Action
}

func NewLiveEditor(repository ListsRepository) *LiveEditor {
	return &LiveEditor{
		listRepository: repository,
		listsById:      make(map[int64]ListUi),
	}
}

func (l *LiveEditor) GetConnectionsOfList(listId int64) []*websocket.Conn {
	result := make([]*websocket.Conn, 0)
	listUi, ok := l.listsById[listId]
	if !ok {
		return result
	}
	for _, c := range listUi.ColaboratorsOnline {
		result = append(result, c.Connections...)
	}
	return result
}

func (l *LiveEditor) HandleAddGroup(listId int64, groupText string) {
	editList := l.GetCurrentList(listId)
	if editList == nil {
		return
	}
	editList.Groups = append(editList.Groups, Group{Name: groupText, Items: []Item{}})

	for _, conn := range l.GetConnectionsOfList(listId) {
		conn.WriteJSON(editList)
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

func (l *LiveEditor) HandleWebsocketConn(conn *websocket.Conn) {
	conn.WriteMessage(websocket.TextMessage, []byte("Hello"))
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		switch messageType {
		case websocket.CloseMessage:
		case websocket.PingMessage:
		case websocket.PongMessage:
		case websocket.BinaryMessage:
			var msg json.RawMessage
			action := Action{Msg: &msg}
			if err := json.Unmarshal(p, &action); err != nil {
				log.Fatal(err)
			}
			switch action.Type {
			case ADD_GROUP_ACTION:
				var addGroupAction AddItemAction
				if err := json.Unmarshal(msg, &addGroupAction); err != nil {
					log.Fatal(err)
				}
				l.HandleAddGroup(addGroupAction.ListId, addGroupAction.GroupText)
			}
		case websocket.TextMessage:
		}
	}
}

func (l *LiveEditor) SetupList(listId int64, user user.User, conn *websocket.Conn) {
	list, err := l.listRepository.Get(listId)
	panicIfError(err)
	listUi, ok := l.listsById[listId]
	if !ok {
		l.listsById[listId] = ListUi{
			List:               list,
			ColaboratorsOnline: []UserUi{{User: user, Connections: []*websocket.Conn{conn}}},
		}
	} else {
		for _, u := range listUi.ColaboratorsOnline {
			if u.Id == user.Id {
				u.Connections = append(u.Connections, conn)
			}
		}
	}
	go l.HandleWebsocketConn(conn)
}

const (
	ADD_GROUP    = iota
	EDIT_GROU    = iota
	FOCUS_ITEM   = iota
	UNFOCUS_ITEM = iota
	ADD_ITEM     = iota
)

func (l *LiveEditor) HandleFocusItem(listId int64, itemId int64) {
}

func (l *LiveEditor) HandleUnfocusItem(listId int64, itemId int64) {
}

func (l *LiveEditor) HandleAddItem(listId int64, groupIndex int, itemText string) {
	editList := l.GetCurrentList(listId)
	items := editList.Groups[groupIndex].Items
	if groupIndex < 0 || groupIndex >= len(editList.Groups) {
		return
	}
	editList.Groups[groupIndex].Items = append(items, Item{Id: 0, Description: itemText, Order: len(items)})
}

func (l *LiveEditor) GetCurrentList(listId int64) *ListUi {
	list, ok := l.listsById[listId]
	if ok {
		return &list
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
	Type int         `json:"type"`
	Msg  interface{} `json:"msg"`
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
