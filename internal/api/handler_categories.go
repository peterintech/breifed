package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
)

func (ac *apiConfig) getCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := ac.DB.GetCategories(r.Context())
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch categories")
		return
	}
	jsonResponse(w, http.StatusOK, databaseCategoriesToCategories(categories))
}

func (ac *apiConfig) replaceUserCategoriesHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	var params struct {
		CategoryIDs []uuid.UUID `json:"category_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := ac.Conn.BeginTx(r.Context(), nil)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not update interests")
		return
	}
	defer tx.Rollback()

	queries := ac.DB.WithTx(tx)
	if err := queries.DeleteUserCategories(r.Context(), user.ID); err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not update interests")
		return
	}
	for _, categoryID := range params.CategoryIDs {
		if err := queries.CreateUserCategory(r.Context(), database.CreateUserCategoryParams{
			UserID: user.ID, CategoryID: categoryID,
		}); err != nil {
			errorResponse(w, http.StatusBadRequest, "one or more category IDs are invalid")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not update interests")
		return
	}

	categories, err := ac.DB.GetCategoriesForUser(r.Context(), user.ID)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch interests")
		return
	}
	jsonResponse(w, http.StatusOK, databaseCategoriesToCategories(categories))
}
