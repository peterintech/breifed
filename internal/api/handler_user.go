package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/accounts"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/sessionauth"
)

type authParameters struct {
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	Password    string      `json:"password"`
	CategoryIDs []uuid.UUID `json:"category_ids"`
	FeedIDs     []uuid.UUID `json:"feed_ids"`
}

func (ac *apiConfig) registerHandler(w http.ResponseWriter, r *http.Request) {
	var params authParameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := accounts.Register(r.Context(), ac.Conn, ac.DB, accounts.RegisterInput{
		Name: params.Name, Email: params.Email, Password: params.Password,
		CategoryIDs: params.CategoryIDs, FeedIDs: params.FeedIDs,
	})
	switch {
	case errors.Is(err, accounts.ErrInvalidRegistration), errors.Is(err, accounts.ErrInvalidChoices):
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	case errors.Is(err, accounts.ErrEmailExists):
		errorResponse(w, http.StatusConflict, err.Error())
		return
	case err != nil:
		errorResponse(w, http.StatusInternalServerError, "could not finish registration")
		return
	}

	sessionauth.SetCookie(w, r, result.Token, result.ExpiresAt)
	jsonResponse(w, http.StatusCreated, databaseUserToUser(result.User))
}

func (ac *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	var params authParameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := accounts.Login(r.Context(), ac.DB, params.Email, params.Password)
	if errors.Is(err, accounts.ErrInvalidCredentials) {
		errorResponse(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not create session")
		return
	}

	sessionauth.SetCookie(w, r, result.Token, result.ExpiresAt)
	jsonResponse(w, http.StatusOK, databaseUserToUser(result.User))
}

func (ac *apiConfig) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionauth.CookieName); err == nil && cookie.Value != "" {
		_ = ac.DB.DeleteSessionByToken(r.Context(), cookie.Value)
	}
	sessionauth.ClearCookie(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (ac *apiConfig) getMeHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	profile, err := ac.profileResponse(r.Context(), user)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not fetch profile")
		return
	}
	jsonResponse(w, http.StatusOK, profile)
}

func (ac *apiConfig) profileResponse(ctx context.Context, user database.User) (Profile, error) {
	categories, err := ac.DB.GetCategoriesForUser(ctx, user.ID)
	if err != nil {
		return Profile{}, err
	}
	feeds, err := ac.DB.GetFeedsForUser(ctx, user.ID)
	if err != nil {
		return Profile{}, err
	}
	feedResponse, err := ac.feedResponses(ctx, feeds)
	if err != nil {
		return Profile{}, err
	}
	return Profile{
		User: databaseUserToUser(user), Categories: databaseCategoriesToCategories(categories), Feeds: feedResponse,
	}, nil
}
