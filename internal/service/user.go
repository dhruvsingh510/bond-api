package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"encoding/json"
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
	InteractedPosts `json:"interacted_posts,omitempty"`
	Karma           int64 `json:"karma,omitempty"`
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

	var interactedPostsJSONB json.RawMessage = nil

	query := "INSERT INTO users (email, password, username, interacted_posts) VALUES ($1, $2, $3, $4)"
	_, err := s.Db.Exec(ctx, query, email, hash, username, interactedPostsJSONB)
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
func (s *Service) User (ctx context.Context, username string) (UserProfile, error) {
	var u UserProfile
	
	username = strings.TrimSpace(username)
	if !rxUsername.MatchString(username) {
		return u, ErrInvalidUsername
	}

	uid, auth := ctx.Value(KeyAuthUserID).(int64)

	query := "SELECT id, email, karma FROM users WHERE username = $1"
	err := s.Db.QueryRow(ctx, query, username).Scan(&u.ID, &u.Email, &u.Karma)
	if err == sql.ErrNoRows {
		return u, ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("could not query select user: %v", err)
	}

	u.Username = username
	if !auth || uid != u.ID {
		u.ID = 0
		u.Email = ""
	}

	return u, nil
}

func (s *Service) ReadUsers(ctx context.Context) error {
	query := "SELECT * FROM users"
	rows, err := s.Db.Query(ctx, query)

	if err != nil {
		return fmt.Errorf("could not execute get users query: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var username, password, email string
		var id int
		if err := rows.Scan(&id, &username, &email, &password); err != nil {
			fmt.Printf("could not read users : %v", err)
		}

		fmt.Printf("id: %d\n username: %s\n email: %s\n password: %s\n---\n\n", id, username, email, password)
	}

	return nil
}

 
