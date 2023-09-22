package handler

import (
	"net/http"
	"context"
	"github.com/dhruvsingh510/bond_social_api/internal/service"
	"github.com/matryer/way"
)

type handler struct {
	*service.Service
}

type Service interface {
	AuthUser(ctx context.Context) (service.User, error)
}

// New creates a http.Handler with predefined routing
func New(s *service.Service) http.Handler {
	h := &handler{s}

	api := way.NewRouter()
	api.HandleFunc("GET", "auth_user", h.authUser)
	api.HandleFunc("POST", "/login", h.login)

	api.HandleFunc("POST", "/users", h.createUser)
	api.HandleFunc("GET", "/users/:username", h.user)
	api.HandleFunc("GET", "/users/:username/posts", h.posts)

	api.HandleFunc("POST", "/posts", h.createPost)
	api.HandleFunc("GET", "/posts/:post_id", h.post)
	api.HandleFunc("POST", "/posts/action", h.postVote)
	api.HandleFunc("POST", "/posts/comment", h.postComment)

	r := way.NewRouter()
	r.Handle("*", "/api...", http.StripPrefix("/api", h.withAuth(api)))

	return r
}
