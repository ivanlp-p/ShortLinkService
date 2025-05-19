package models

type ShortURL struct {
	Result string `json:"result"`
}

type OriginalURL struct {
	URL string `json:"url"`
}
