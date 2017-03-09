package controllers

type Meta struct {
	CurrentPage *int `bson:"current_page" json:"current_page"`
	NextPage    *int `bson:"next_page" json:"next_page"`
	PrevPage    *int `bson:"prev_page" json:"prev_page"`
	TotalPages  *int `bson:"total_pages" json:"total_pages"`
	TotalCount  *int `bson:"total_count" json:"total_count"`
	Offset      *int `bson:"offset" json:"offset"`
}
