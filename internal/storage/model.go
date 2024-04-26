package storage

type URLData struct {
	URL string
	ID  string
}

type ReqBatchURLs []ReqBatchURL

type ReqBatchURL struct {
	ExternalID string `json:"correlation_id"`
	ID         string
	URL        string `json:"original_url"`
	Err        error
}

type RespBatchURLs []RespBatchURL

type RespBatchURL struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}

type UserURLs []URLData
