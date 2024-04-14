package storage

type Repository interface {
	// Возвращает ошибку, если не удалось добавить URL
	AddURL(data URLData) error
	// Возвращает \URLData и true, если идентификатор найден, иначе возвращает пустую структуру \URLData и false.
	CheckID(id string) (URLData, bool, error)
	// Возвращает \URLData и true, если URL найден, иначе возвращает пустую структуру \URLData и false.
	CheckURL(url string) (URLData, bool, error)
	// Возвращает ошибку, если не удалось удалить \URLData.
	RemoveURL(data URLData) error
	// \Ping проверяет соединение с БД
	// Возвращает 200 или 500
	Ping() error
}

type URLData struct {
	URL string
	ID  string
}
