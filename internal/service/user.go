package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	rxEmail            = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$`)
	rxUsername         = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{3,30}$`)
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidUsername = errors.New("invalid username")
	ErrEmailTaken      = errors.New("email taken")
	ErrUsernameTaken   = errors.New("username taken")
	ErrHashingPass     = errors.New("error hashing password")
)

// User model
type User struct {
	ID       int64  `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// UserProfile model
type UserProfile struct {
	User            `json:"user,omitempty"`
	Email           string `json:"email,omitempty"`
	Karma           int64 `json:"karma,omitempty"`
	UpvotedPosts	[]int64 `json:"upvoted_posts,omitempty"`
	DownvotedPosts	[]int64 `json:"downvoted_posts,omitempty"`
}

// Inserts a new user in the database
func (s *Service) CreateUser(ctx context.Context, email string, password string, username string) error {
	email = strings.TrimSpace(email)
	if !rxEmail.MatchString(email) {
		return ErrInvalidEmail
	}

	username = strings.TrimSpace(username)
	if !rxUsername.MatchString(username) {
		return ErrInvalidUsername
	}

	hash, b_err := hashPassword(password)
	if b_err != nil {
		return ErrHashingPass
	}

	query := "INSERT INTO users (email, password, username) VALUES ($1, $2, $3)"
	_, err := s.Db.Exec(ctx, query, email, hash, username)
	unique := isUniqueViolation(err)

	if err != nil && !unique && strings.Contains(err.Error(), "email") {
		return ErrEmailTaken
	}

	if err != nil && !unique && strings.Contains(err.Error(), "username") {
		return ErrUsernameTaken
	}

	if err != nil {
		return fmt.Errorf("could not insert user: %v", err)
	}

	return nil
}

// User select on user from the database with the given username
func (s *Service) User(ctx context.Context, username string) (UserProfile, error) {
	var u UserProfile
	
	username = strings.TrimSpace(username)
	if !rxUsername.MatchString(username) {
		return u, ErrInvalidUsername
	}

	uid, auth := ctx.Value(KeyAuthUserID).(int64)
	if !auth {
		return u, ErrUnauthenticated
	}

	query := "SELECT id, email, karma FROM users WHERE username = $1"
	err := s.Db.QueryRow(ctx, query, username).Scan(&u.ID, &u.Email, &u.Karma)
	if err == sql.ErrNoRows {
		return u, ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("could not query select user: %v", err)
	}

	query = "SELECT post_id, vote_type FROM post_votes WHERE user_id = $1"
	rows, err := s.Db.Query(ctx, query, uid)
	if err != nil {
		return u, fmt.Errorf("could not sql query user upvoted posts: %v", err)
	}

	defer rows.Close()

	var postID int64
	var voteType string
	for rows.Next() {
		if err := rows.Scan(&postID, &voteType); err != nil {
			return u, fmt.Errorf("could not iterate over user posts: %v", err)
		}

		if voteType == "upvote" {
			u.UpvotedPosts = append(u.UpvotedPosts, postID)
		} else if voteType == "downvote" {
			u.DownvotedPosts = append(u.DownvotedPosts, postID)
		}
	}
	
	return u, nil
}


 
