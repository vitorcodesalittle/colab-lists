package list

type ListsRepository interface {
	GetAll(userId int64) ([]List, error)
	Get(id int64) (List, error)
	Create(list *ListCreationParams) (List, error)
	Update(list *List) (*List, error)
	Delete(id int) error
}
