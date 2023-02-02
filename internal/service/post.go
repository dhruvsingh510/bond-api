package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	// "encoding/json"
)

var (
	ErrInvalidTitle = errors.New("invalid title")
	ErrInvalidBody = errors.New("invalid body")
	ErrInvalidLink = errors.New("invalid link")
)

// Post Model
type Post struct {
	ID        int64     `json:"id,omitempty"`
	UserID    int64     `json:"user_id,omitempty"`
	Title     string    `json:"title,omitempty"`
	Body      string    `json:"body,omitempty"`
	Link      string    `json:"link,omitempty"`
	User      *User     `json:"user,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	// album     []string  `json:"album,omitempty"`
	// poll      []string  `json:"poll,omitempty"`
}

func (s *Service) CreatePost (
	ctx context.Context,
	title string,
	body string,
	link string,
	// album []string,
	// poll []string
) (TimelineItem, error) {
	var ti TimelineItem

	uid, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return ti, ErrUnauthenticated
	}

	title = strings.TrimSpace(title)
	if title == "" || len([]rune(title)) > 480 {
		return ti, ErrInvalidTitle
	}

	body = strings.TrimSpace(body)
	if len([]rune(body)) > 480 {
		return ti, ErrInvalidBody
	}

	link = strings.TrimSpace(link)
	if len([]rune(link)) > 480 {
		return ti, ErrInvalidLink
	}

	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return ti, fmt.Errorf("could not begin transaction: %v", err)
	}

	defer tx.Rollback(ctx)

	query := "INSERT INTO posts (user_id, title, body, link) VALUES ($1, $2, $3, $4) "+
	"RETURNING id, created_at"

	if err = tx.QueryRow(ctx, query, uid, title, body, link).Scan(&ti.Post.ID, &ti.Post.CreatedAt); err != nil {
		return ti, fmt.Errorf("could not insert post: %v", err)
	}

	ti.Post.UserID = uid
	ti.Post.Title = title
	ti.Post.Body = body
	ti.Post.Link = link

	query = "INSERT INTO timeline (user_id, post_id) VALUES ($1, $2) RETURNING id"
	if err = tx.QueryRow(ctx, query, uid, ti.Post.ID).Scan(&ti.ID); err != nil {
		return ti, fmt.Errorf("could not insert into timeline: %v", err)
	}

	ti.UserID = uid
	ti.PostID = ti.Post.ID

	tx.Commit(ctx)

	return ti, nil
}



