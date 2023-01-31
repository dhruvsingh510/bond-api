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
	ErrForbiddenFollow = errors.New("can not follow yourself")
)

// User model
type User struct {
	ID       int64  `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// ToggleFollow output response
type ToggleFollowOutput struct {
	Following      bool `json:"following,omitempty"`
	FollowersCount int  `json:"followers_count,omitempty"`
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

func (s *Service) ToggleFollow(ctx context.Context, username string) (ToggleFollowOutput, error) {
	var out ToggleFollowOutput
	followerID, ok := ctx.Value(KeyAuthUserID).(int64)
	if !ok {
		return out, ErrUnauthenticated
	}

	username = strings.TrimSpace(username)
	if !rxUsername.MatchString(username) {
		return out, ErrInvalidUsername
	}

	tx, err := s.Db.Begin(ctx)
	if err != nil {
		return out, fmt.Errorf("could not begin tx: %v", err)
	}

	defer tx.Rollback(ctx)

	var followeeID int64
	query := "SELECT id FROM users WHERE username = $1"
	err = tx.QueryRow(ctx, query, username).Scan(&followeeID)
	if err == sql.ErrNoRows {
		return out, ErrUserNotFound
	}

	if err != nil {
		return out, fmt.Errorf("could not query select id from followee username: %v", err)
	}

	if followeeID == followerID {
		return out, ErrForbiddenFollow
	}

	query = "SELECT EXISTS (SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = $2)"
	if err = tx.QueryRow(ctx, query, followerID, followeeID).Scan(&out.Following); err != nil {
		return out, fmt.Errorf("could not query select existance of user: %v", err)
	}

	if out.Following {
		query = "DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2"
		if _, err = tx.Exec(ctx, query, followerID, followeeID); err != nil {
			return out, fmt.Errorf("could not delete follow: %v", err)
		}

		query = "UPDATE users SET followees_count = followees_count - 1 WHERE id = $1"
		if _, err = tx.Exec(ctx, query, followerID); err != nil {
			return out, fmt.Errorf("could not update follower followees count: %v", err)
		}

		query = "UPDATE users SET followers_count = followers_count - 1 WHERE id = $1 RETURNING followers_count"
		if err = tx.QueryRow(ctx, query, followeeID).Scan(&out.FollowersCount); err != nil {
			return out, fmt.Errorf("could not update followee followers count: %v", err)
		}
	} else {
		query = "INSERT INTO follows (follower_id, followee_id) VALUES ($1, $2)"
		if _, err = tx.Exec(ctx, query, followerID, followeeID); err != nil {
			return out, fmt.Errorf("could not insert follow: %v", err)
		}

		query = "UPDATE users SET followees_count = followees_count + 1 WHERE id = $1"
		if _, err = tx.Exec(ctx, query, followerID); err != nil {
			return out, fmt.Errorf("could not update follower followees count: %v", err)
		}

		query = "UPDATE users SET followers_count = followers_count + 1 WHERE id = $1 RETURNING followers_count"
		if err = tx.QueryRow(ctx, query, followeeID).Scan(&out.FollowersCount); err != nil {
			return out, fmt.Errorf("could not update followee followers count: %v", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return out, fmt.Errorf("could not commit toggle: %v", err)
	}

	out.Following = !out.Following

	// if out.Following {
	// 	// TODO : notify followee 
	// }

	return out, nil

}
