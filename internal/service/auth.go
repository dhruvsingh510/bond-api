package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
	"errors"
)

const (
	TokenLifespan = time.Hour * 24 * 14
	// KeyAuthUserID to use in context
	KeyAuthUserID key = "auth_user_id"
)

type key string

// Login Output response
type LoginOutput struct {
	Token string
	ExpiresAt time.Time
	AuthUser User
}

var (
	// ErrUnauthenticated used when there is no authenticated user in context
	ErrUnauthenticated = errors.New("unauthenticated")
)

// AuthUserID from token
func (s *Service) AuthUserID(token string) (int64, error) {
	str, err := s.Codec.DecodeToString(token)
	if err != nil {
		return 0, fmt.Errorf("could not decode token: %v", err)
	}

	i, err := strconv.ParseInt(str, 10, 64) 
	if err != nil {
		return 0, fmt.Errorf("could not parse auth user id from token: %v", err)
	}

	return i, nil
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
		return out, ErrUserNotFound
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

// AuthUser from context
func (s *Service) AuthUser(ctx context.Context) (User, error) {
	var u User

	uid, auth := ctx.Value(KeyAuthUserID).(int64)
	if !auth {
		return u, ErrUnauthenticated
	}

	query := "SELECT username FROM users WHERE id = $1"
	err := s.Db.QueryRow(ctx, query, uid).Scan(&u.Username)
	if err == sql.ErrNoRows {
		return u, ErrUserNotFound
	}

	if err != nil {
		return u, fmt.Errorf("could not query select auth user: %v", err)
	}

	u.ID = uid
	
	return u, nil
}