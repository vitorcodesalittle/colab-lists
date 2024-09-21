package realtime


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
type Action struct {
	Type *int `json:"actionType"`
	Msg  interface{}
}

type FocusItemAction struct {
	GroupIndex int64 `json:"groupIndex"`
	ItemIndex  int64 `json:"itemIndex"`
}

type UnfocusItemAction struct {
	GroupIndex int64
	ItemIndex  int64
}

type UpdateColorAction struct {
	Color      string `json:"color"`
	UserId     int64  `json:"userId"`
	GroupIndex int64    `json:"groupIndex"`
}

type EditGroupAction struct {
	GroupIndex int64    `json:"groupIndex"`
	Text       string `json:"text"`
}

type BlurItemAction struct {
	GroupIndex int64 `json:"groupIndex"`
	ItemIndex  int64 `json:"itemIndex"`
}

type AddItemAction struct {
	GroupIndex int64 `json:"groupIndex"`
}

type EditItemAction struct {
	GroupIndex int64 `json:"groupIndex"`
}

type DeleteGroupArgs struct {
	GroupIndex int64 `json:"groupIndex"`
}

type EditItemArgs struct {
	GroupIndex  int64    `json:"groupIndex"`
	ItemIndex   int64    `json:"itemIndex"`
	Description string `json:"description"`
	Quantity    string `json:"quantity"`
}
