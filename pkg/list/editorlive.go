package list

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"vilmasoftware.com/colablists/pkg/user"
	"vilmasoftware.com/colablists/pkg/views"
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

func (l *LiveEditor) Info() {
	println("Live editor info")
	for k, v := range l.listsById {
		println("List ", k, " has ", len(v.ColaboratorsOnline), " colaborators")
		for _, u := range v.ColaboratorsOnline {
			println("User ", u.Id, " has ", len(u.Connections), " connections")
		}
	}
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

func (l *LiveEditor) removeConnection(conn *websocket.Conn) {
	for listId, v := range l.listsById {
		for userIdx, u := range v.ColaboratorsOnline {
			removeIndex := -1
			for connIdx, conn2 := range u.Connections {
				if conn2 == conn {
					removeIndex = connIdx
					break
				}
			}
			if removeIndex >= 0 {
				l, _ := l.listsById[listId]
				l.ColaboratorsOnline[userIdx].Connections = append(u.Connections[:removeIndex], u.Connections[removeIndex+1:]...)

			}
		}
	}
}

func (l *LiveEditor) HandleWebsocketConn(conn *websocket.Conn) {
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
			return
		case websocket.PingMessage:
			log.Println("PingMessage")
			return
		case websocket.PongMessage:
			log.Println("PongMessage")
			return
		case websocket.BinaryMessage:
			log.Println("BinaryMessage")
			return
		case websocket.TextMessage:
			log.Println("TextMessage")
			var msg json.RawMessage
			action := Action{Msg: &msg}
			if err := json.Unmarshal(p, &action); err != nil {
				log.Fatal(err)
			}
			switch action.Type {
            case FOCUS_ITEM:

			case ADD_GROUP_ACTION:
				var addGroupAction AddItemAction
				if err := json.Unmarshal(msg, &addGroupAction); err != nil {
					log.Fatal(err)
				}
				l.HandleAddGroup(addGroupAction.ListId, addGroupAction.GroupText)
			}
			return
		}
	}
}

func (l *LiveEditor) SetupList(listId int64, user user.User, conn *websocket.Conn) {
	defer l.Info()
	listUi, ok := l.listsById[listId]
	if !ok {
		list, err := l.listRepository.Get(listId)
		panicIfError(err)
		l.listsById[listId] = ListUi{
			List:               list,
			ColaboratorsOnline: []UserUi{{User: user, Connections: []*websocket.Conn{conn}}},
		}
		listUi = l.listsById[listId]
	} else {
		found := false
		for idx, u := range listUi.ColaboratorsOnline {
			if u.Id == user.Id {
				listUi.ColaboratorsOnline[idx].Connections = append(u.Connections, conn)
				found = true
			}
		}
		if !found {
			listUi.ColaboratorsOnline = append(listUi.ColaboratorsOnline, UserUi{User: user, Connections: []*websocket.Conn{conn}})
			l.listsById[listId] = listUi
		}
	}
	s := ""
	buf := bytes.NewBufferString(s)
	views.Templates.RenderCollaboratorsList(buf, listUi.ColaboratorsOnline)
	conns := l.GetConnectionsOfList(listId)
	for _, conn2 := range conns {
		conn2.WriteMessage(websocket.TextMessage, buf.Bytes())
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


type FocusItemAction struct {
	ListId    int64  `json:"listId"`
	GroupIndex int `json:"groupIndex"`
	ItemIndex int `json:"itemIndex"`
}

type BlurItemAction struct {
	ListId    int64  `json:"listId"`
	GroupIndex int `json:"groupIndex"`
	ItemIndex int `json:"itemIndex"`
}

type AddItemAction struct {
	ListId    int64  `json:"listId"`
	GroupText string `json:"groupText"`
}

type HtmxMessage struct {
	CurrentUrl  *string `json:"HX-Current-URL"`
	Request     *string `json:"HX-Request"`
	Target      *string `json:"HX-Target"`
	Trigger     *string `json:"HX-Trigger"`
	TriggerName *string `json:"HX-TriggerName"`
	ActionType  *int    `json:"action-type"`
}

func (h *HtmxMessage) String() string {
    return fmt.Sprintf("CurrentUrl: %s, Request: %s, Target: %s, Trigger: %s, TriggerName: %s, ActionType: %d\n", h.CurrentUrl, h.Request, h.Target, h.Trigger, h.TriggerName, h.ActionType)
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
