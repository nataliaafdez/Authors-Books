package dto

type CreateBookReq struct {
	AuthorID int64  `json:"authorId"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Genre    string `json:"genre"`
	Language string `json:"language"`
}
type BookResp struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Genre    string `json:"genre"`
	Language string `json:"language"`
}
