package storage

type Repository interface {
	// AddURL добавляет URL в репозиторий.
	// Возвращает ошибку, если не удалось добавить URL
	AddURL(data URLData) error
	// CheckID проверяет наличие URLData с указанным идентификатором.
	// Возвращает URLData и true, если идентификатор найден, иначе возвращает пустую структуру URLData и false.
	CheckID(id string) (URLData, bool, error)
	// CheckURL проверяет наличие URLData с указанным URL.
	// Возвращает URLData и true, если URL найден, иначе возвращает пустую структуру URLData и false.
	CheckURL(url string) (URLData, bool, error)
	// RemoveURL удаляет указанный URLData из репозитория.
	// Возвращает ошибку, если не удалось удалить URLData.
	RemoveURL(data URLData) error
	// Ping проверяет соединение с БД
	// Возвращает 200 или 500
	Ping() error
}

type URLData struct {
	URL string
	ID  string
}
