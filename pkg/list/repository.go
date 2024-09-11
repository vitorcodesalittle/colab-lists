package list

type ListsRepository interface {
	GetAll() ([]List, error)
	Get(id int64) (List, error)
	Create(list *ListCreationParams) (List, error)
	Update(list *ListUpdateParams) (List, error)
	Delete(id int) error
}
