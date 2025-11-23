package domain

type Book struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Genre    string `json:"genre"`
	Language string `json:"language"`
}
