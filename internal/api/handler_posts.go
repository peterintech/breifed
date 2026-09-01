package api

import (
	"net/http"

	"github.com/peterintech/briefed/internal/database"
)

func (ac *apiConfig) getPostsForUserHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	posts, err := ac.DB.GetPostsForUser(r.Context(), database.GetPostsForUserParams{
		UserID: user.ID, Limit: limit, Offset: offset,
	})
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch posts")
		return
	}
	jsonResponse(w, http.StatusOK, databasePostsToPosts(posts))
}
