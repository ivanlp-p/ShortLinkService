package models

type ShortURL struct {
	Result string `json:"result"`
}

type OriginalURL struct {
	URL string `json:"url"`
}

type ShortLink struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchRequest struct {
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
