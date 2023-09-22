package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidTitle  = errors.New("invalid title")
	ErrInvalidBody   = errors.New("invalid body")
	ErrInvalidLink   = errors.New("invalid link")
	ErrNoContent     = errors.New("error: no content to post")
	ErrInvalidPostID = errors.New("error: no such post id exists")
)

// Post Model
type Post struct {
	ID        int64          `json:"id,omitempty"`
	UserID    int64          `json:"user_id,omitempty"`
	Title     string         `json:"title,omitempty"`
	Body      string         `json:"body,omitempty"`
	Link      string         `json:"link,omitempty"`
	Album     sql.NullString `json:"album,omitempty"`
	Poll      sql.NullString `json:"poll,omitempty"`
	Upvotes   int64          `json:"upvotes,omitempty"`
	Downvotes int64          `json:"downvotes,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
	User      *User          `json:"user,omitempty"`
}

func (s *Service) CreatePost(
	ctx context.Context,
	title string,
	body string,
	link string,
	album string,
	poll string,
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

	var albumJSONB json.RawMessage
	if album == "" {
		albumJSONB = nil
	} else {
		err := json.Unmarshal([]byte(album), &albumJSONB)
		if err != nil {
			return ti, fmt.Errorf("error converting album string to jsonb: %v", err)
		}
	}

	var pollJSONB json.RawMessage
	if poll == "" {
		pollJSONB = nil
	} else {
		err := json.Unmarshal([]byte(poll), &pollJSONB)
		if err != nil {
			return ti, fmt.Errorf("error converting poll string to jsonb: %v", err)
		}
	}

	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return ti, fmt.Errorf("could not begin transaction: %v", err)
	}

	defer tx.Rollback(ctx)

	var query string
	if album == "" && poll == "" {
		query = "INSERT INTO posts (user_id, title, body, link) VALUES ($1, $2, $3, $4) RETURNING id, created_at"
		if err = tx.QueryRow(ctx, query, uid, title, body, link).Scan(&ti.Post.ID, &ti.Post.CreatedAt); err != nil {
			return ti, fmt.Errorf("could not insert post: %v", err)
		}
	} else if album == "" {
		query = "INSERT INTO posts (user_id, title, body, link, poll) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at"
		if err = tx.QueryRow(ctx, query, uid, title, body, link, pollJSONB).Scan(&ti.Post.ID, &ti.Post.CreatedAt); err != nil {
			return ti, fmt.Errorf("could not insert post: %v", err)
		}
	} else if poll == "" {
		query = "INSERT INTO posts (user_id, title, body, link, album) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at"
		if err = tx.QueryRow(ctx, query, uid, title, body, link, albumJSONB).Scan(&ti.Post.ID, &ti.Post.CreatedAt); err != nil {
			return ti, fmt.Errorf("could not insert post: %v", err)
		}
	} else {
		query = "INSERT INTO posts (user_id, title, body, link, album, poll) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at"
		if err = tx.QueryRow(ctx, query, uid, title, body, link, albumJSONB, pollJSONB).Scan(&ti.Post.ID, &ti.Post.CreatedAt); err != nil {
			return ti, fmt.Errorf("could not insert post: %v", err)
		}
	}

	ti.Post.UserID = uid
	ti.Post.Title = title
	ti.Post.Body = body
	ti.Post.Link = link

	ti.Post.Album.String = album
	if album != "" {
		ti.Post.Album.Valid = true
	} 

	ti.Post.Poll.String = poll
	if poll != "" {
		ti.Post.Poll.Valid = true
	}

	query = "INSERT INTO timeline (user_id, post_id) VALUES ($1, $2) RETURNING id"
	if err = tx.QueryRow(ctx, query, uid, ti.Post.ID).Scan(&ti.ID); err != nil {
		return ti, fmt.Errorf("could not insert into timeline: %v", err)
	}

	ti.UserID = uid
	ti.PostID = ti.Post.ID

	tx.Commit(ctx)

	return ti, nil
}

// Gets all posts of a user
func (s *Service) Posts(
	ctx context.Context,
	username string,
) ([]Post, error) {

	username = strings.TrimSpace(username)
	if !rxUsername.MatchString(username) {
		return nil, ErrInvalidUsername
	}

	uid, auth := ctx.Value(KeyAuthUserID).(int64)
	if !auth {
		return nil, ErrUnauthenticated
	}

	query := "SELECT title, body, link, album, poll FROM posts WHERE user_id = $1 ORDER BY created_at DESC"

	rows, err := s.Db.Query(ctx, query, uid)
	if err != nil {
		return nil, fmt.Errorf("could not sql query user posts: %v", err)
	}

	defer rows.Close()

	posts := []Post{}
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.Title, &post.Body, &post.Link, &post.Album, &post.Poll); err != nil {
			return nil, fmt.Errorf("could not iterate over user posts: %v", err)
		}

		posts = append(posts, post)
	}

	return posts, nil
}

// Gets a particular post
func (s *Service) Post(
	ctx context.Context,
	postID string,
) (Post, error) {
	var p Post

	p_id, err := strconv.Atoi(postID)
	if err != nil {
		return p, fmt.Errorf("could not convert post string to int: %v", err)
	}

	uid, auth := ctx.Value(KeyAuthUserID).(int64)
	if !auth {
		return p, ErrUnauthenticated
	}

	_ = uid

	query := "SELECT title, body, link, album, poll FROM posts WHERE id = $1"

	err = s.Db.QueryRow(ctx, query, p_id).Scan(&p.Title, &p.Body, &p.Link, &p.Album, &p.Poll)
	if err == sql.ErrNoRows {
		return p, fmt.Errorf("could not sql query user post: %v", err)
	}

	if p.Album.String == "null" {
		p.Album.String = ""
		p.Album.Valid = false
	}

	if p.Poll.String == "null" {
		p.Poll.String = ""
		p.Poll.Valid = false
	}

	return p, nil
}

func (s *Service) PostVote(
	ctx context.Context,
	postID int64,
	action string,
) error {

	uid, auth := ctx.Value(KeyAuthUserID).(int64)
	if !auth {
		return ErrUnauthenticated
	}

	var query string 
	switch action {
	case "removeUpvote":
		query = "UPDATE posts SET upvotes = GREATEST(0, upvotes - 1) WHERE id = $1"
	case "removeDownvote":
		query = "UPDATE posts SET downvotes = GREATEST(0, downvotes - 1) WHERE id = $1"
	case "upvote":
		query = "UPDATE posts SET upvotes = upvotes + 1 WHERE id = $1"
	case "downvote":
		query = "UPDATE posts SET downvotes = downvotes + 1 WHERE id = $1"
	}

	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, postID)
	if err != nil {
		return fmt.Errorf("unable to perform the query update action on post: %v", err)
	}

	if action == "removeUpvote" || action == "removeDownvote" {
		query = "DELETE FROM post_votes WHERE user_id = $1 AND post_id = $2"
		_, err = tx.Exec(ctx, query, uid, postID)
		if err != nil {
			return fmt.Errorf("unable to delete post vote: %v", err)
		}
	} else if action == "upvote" || action == "downvote" {
		query = "INSERT INTO post_votes (user_id, post_id, vote_type) VALUES ($1, $2, $3)"
		_, err = tx.Exec(ctx, query, uid, postID, action)
		if err != nil {
			return fmt.Errorf("unable to insert post vote: %v", err)
		}
	}

	tx.Commit(ctx)

	return nil
}

func (s *Service) PostComment (
	ctx context.Context,
	postID int64,
	parentCommentID int64,
	comment string,
) (error) {
	uid, auth := ctx.Value(KeyAuthUserID).(int64)
	if !auth {
		return ErrUnauthenticated
	}

	_ = uid

	var query string
	var err error
	if parentCommentID != 0 {
		query = "INSERT INTO post_comments (post_id, parent_id, content) VALUES ($1, $2, $3)"
		_, err = s.Db.Exec(ctx, query, postID, parentCommentID, comment)
	} else {
		query = "INSERT INTO post_comments (post_id, content) VALUES ($1, $2)"
		_, err = s.Db.Exec(ctx, query, postID, comment)
	}
	
	if err != nil {
		return fmt.Errorf("unable to insert comment: %v", err)
	}

	return nil
}