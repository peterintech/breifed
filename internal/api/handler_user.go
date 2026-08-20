package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/peterintech/briefed/internal/database"
)

func (ac *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		errorResponse(w, 400, fmt.Sprint("Error parsing JSON:", err))
		return
	}

	user, err := ac.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
	})

	if err != nil {
		errorResponse(w, 500, fmt.Sprint("Error creating user:", err))
		return
	}

	jsonResponse(w, 201, databaseUserToUser(user))
}
func (ac *apiConfig) getUserByApiKey(w http.ResponseWriter, r *http.Request, user database.User) {
	jsonResponse(w, 200, databaseUserToUser(user))
}

func (ac *apiConfig) getPostsForUserHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	limit := int32(10)
	offset := int32(0)

	posts, err := ac.DB.GetPostsForUser(r.Context(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		errorResponse(w, 500, fmt.Sprint("Error fetching posts:", err))
		return
	}

	jsonResponse(w, 200, databasePostsToPosts(posts))
}
