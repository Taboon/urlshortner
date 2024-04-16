package storage

type Repository interface {
	// AddURL Возвращает ошибку, если не удалось добавить URL
	AddURL(data URLData) error
	// AddBatchURL Возвращает ошибку, если не удалось добавить массив URL
	AddBatchURL(map[string]ReqBatchJSON) error
	// CheckID Возвращает \URLData и true, если идентификатор найден, иначе возвращает пустую структуру \URLData и false.
	CheckID(id string) (URLData, bool, error)
	// CheckURL Возвращает \URLData и true, если URL найден, иначе возвращает пустую структуру \URLData и false.
	CheckURL(url string) (URLData, bool, error)
	// CheckBatchURL Проверяет url на наличие в базе. Если присутствует в базе, то свойство Exist = false
	CheckBatchURL(urls *[]ReqBatchJSON) (*[]ReqBatchJSON, error)
	// RemoveURL Возвращает ошибку, если не удалось удалить URLData.
	RemoveURL(data URLData) error
	// Ping проверяет соединение с БД
	// Возвращает 200 или 500
	Ping() error
}

type URLData struct {
	URL string
	ID  string
}

type ReqBatchJSON struct {
	ID    string `json:"correlation_id"`
	URL   string `json:"original_url"`
	Valid bool
	Exist bool
}

type RespBatchJSON struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}
