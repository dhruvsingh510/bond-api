package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dhruvsingh510/bond_social_api/internal/service"
)

type createPostInput struct {
	Title string
	Body string
	Link string
	Album string
	Poll string
}

func (h *handler) createPost(w http.ResponseWriter, r *http.Request) {
	var in createPostInput
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ti, err := h.CreatePost(r.Context(), in.Title, in.Body, in.Link, in.Album, in.Poll)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrInvalidTitle || err == service.ErrInvalidEmail || err == service.ErrInvalidBody || err == service.ErrNoContent {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err != nil {
		respondError(w, err)
		return
	}

	respond(w, ti, http.StatusCreated)
}

