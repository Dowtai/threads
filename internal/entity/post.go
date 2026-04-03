package entity

type Post struct {
	ID              string `json:"id"`
	Author          string `json:"author"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	CommentsAllowed bool   `json:"commentsAllowed"`
}
