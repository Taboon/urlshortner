package storage

type Repositories interface {
	AddUrl(data UrlData) error
	CheckId(id string) (UrlData, bool)
	CheckUrl(url string) (UrlData, bool)
	RemoveUrl(data UrlData) error
}

type TempStorage struct {
}
