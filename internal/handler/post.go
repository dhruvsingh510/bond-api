package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dhruvsingh510/bond_social_api/internal/service"
	"github.com/matryer/way"
)

type createPostInput struct {
	Title string
	Body string
	Link string
	Album string
	Poll string
}

type postEngagementInput struct {
	PostID int64
	Action string
}

type postCommentInput struct {
	PostID int64
	ParentCommentID int64
	Comment string
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

func (h *handler) posts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := way.Param(ctx, "username")

	pp, err := h.Posts(ctx, username) 

	if err == service.ErrInvalidUsername {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err != nil {
		respondError(w, err)
	}

	respond(w, pp, http.StatusOK)
}

func (h *handler) post(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	postID := way.Param(ctx, "post_id")

	p, err := h.Post(ctx, postID) 

	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err != nil {
		respondError(w, err)
	}

	respond(w, p, http.StatusOK)
}

func (h *handler) postVote(w http.ResponseWriter, r *http.Request) {
	var in postEngagementInput
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.PostVote(r.Context(), in.PostID, in.Action)

	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrInvalidPostID {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		respondError(w, err)
		return
	}
	
	respond(w, "success", http.StatusNoContent)
}

func (h *handler) postComment(w http.ResponseWriter, r *http.Request) {
	var in postCommentInput
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.PostComment(r.Context(), in.PostID, in.ParentCommentID, in.Comment)

	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrInvalidPostID {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		respondError(w, err)
		return
	}
	
	respond(w, "success", http.StatusNoContent)
}



