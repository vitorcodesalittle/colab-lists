package list

type ListCreationParams struct {
	Title       string
	Description string
	CreatorId   int
}

const (
	Add    = iota
	Remove = iota
	Update = iota
)

type ItemUpdateParams struct {
	Id          int
	Order       int
	Description string
	Quantity    int
	GroupName   string
	Operation   int
}

type ListUpdateParams struct {
	Id int
	ListCreationParams
	Items []ItemUpdateParams
}
