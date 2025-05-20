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
