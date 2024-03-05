package storage

type Repositories interface {
	AddURL(data URLData) error
	CheckID(id string) (URLData, bool)
	CheckURL(url string) (URLData, bool)
	RemoveURL(data URLData) error
}

type TempStorage struct {
}
