package handler

import (
	"net/http"
	"github.com/dhruvsingh510/bond_social_api/internal/service"
	"github.com/matryer/way"
)

type handler struct {
	*service.Service
}

// New creates a http.Handler with predefined routing
func New(s *service.Service) http.Handler {
	h := &handler{s}

	api := way.NewRouter()
	api.HandleFunc("POST", "/login", h.login)
	api.HandleFunc("GET", "auth_user", h.authUser)
	api.HandleFunc("POST", "/users", h.createUser)
	api.HandleFunc("GET", "/getusers", h.readUsers)
	api.HandleFunc("GET", "/users/:username", h.user)
	api.HandleFunc("POST", "/posts", h.createPost)
	api.HandleFunc("GET", "/users/:username/posts", h.posts)
	api.HandleFunc("GET", "/posts/:post_id", h.post)

	r := way.NewRouter()
	r.Handle("*", "/api...", http.StripPrefix("/api", h.withAuth(api)))

	return r
}