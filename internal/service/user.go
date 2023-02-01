package service

import (
	"context"
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
