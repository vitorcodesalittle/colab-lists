package list

type ListCreationParams struct {
	Title       string
	Description string
	CreatorId   int64
}

const (
	AddItem    = iota
	RemoveItem = iota
	UpdateItem = iota
)

type ItemUpdateParams struct {
	Id          int
	Order       int
	Description string
	Quantity    int
	GroupName   string
	Operation   int
}

