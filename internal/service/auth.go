package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	TokenLifespan = time.Hour * 24 * 14
)

// Login Output response
type LoginOutput struct {
	Token string
	ExpiresAt time.Time
	AuthUser User
}

// Login insecurely
func(s *Service) Login(ctx context.Context, email string, password string) (LoginOutput, error) {
	var out LoginOutput

	email = strings.TrimSpace(email)
	if !rxEmail.MatchString(email) {
		return out, ErrInvalidEmail
	}

	var hash string
	query := "SELECT id, username, password FROM users WHERE email = $1"
	err := s.Db.QueryRow(ctx, query, email).Scan(&out.AuthUser.ID, &out.AuthUser.Username, &hash)

	if err == sql.ErrNoRows {
		return out, ErrorUserNotFound
	}

	if err != nil {
		return out, fmt.Errorf("could not query select user: %v", err)
	}

	if !checkPasswordHash(password, hash) {
		return out, ErrInvalidPassword	
	}

	out.Token, err = s.Codec.EncodeToString(strconv.FormatInt(out.AuthUser.ID, 10))
	if err != nil {
		return out, fmt.Errorf("could not create token: %v", err)
	}

	out.ExpiresAt = time.Now().Add(TokenLifespan)

	return out, nil
}