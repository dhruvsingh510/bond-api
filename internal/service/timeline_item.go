package service

// TImeline Model
type TimelineItem struct {
	ID     int64 `json:"id,omitempty"`
	UserID int64 `json:"user_id"`
	PostID int64 `json:"post_id"`
	Post   Post  `json:"post,omitempty"`
}


