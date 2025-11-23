package dto

type CreateAuthorReq struct {
	Name string `json:"name"`
}

type AddBookReq struct {
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

type AuthorResp struct {
	ID    int64      `json:"id"`
	Name  string     `json:"name"`
	Books []BookResp `json:"books"`
}
