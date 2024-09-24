package house

type HouseRepository struct{}

func (h *HouseRepository) Save(house *House) {

}

func (h *HouseRepository) FindById(id int64) *House {
	return &House{}
}

func (h *HouseRepository) FindMyHouses(userId int64) []*House {
	return []*House{}
}
