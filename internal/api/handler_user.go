package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/sessionauth"
	"golang.org/x/crypto/bcrypt"
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

	params.Name = strings.TrimSpace(params.Name)
	params.Email = strings.ToLower(strings.TrimSpace(params.Email))
	if params.Name == "" || params.Email == "" || len(params.Password) < 8 {
		errorResponse(w, http.StatusBadRequest, "name, email, and a password of at least 8 characters are required")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not secure password")
		return
	}
	token, err := sessionauth.NewToken()
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not create session")
		return
	}

	tx, err := ac.Conn.BeginTx(r.Context(), nil)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not start registration")
		return
	}
	defer tx.Rollback()

	queries := ac.DB.WithTx(tx)
	now := time.Now().UTC()
	user, err := queries.CreateUser(r.Context(), database.CreateUserParams{
		ID: uuid.New(), CreatedAt: now, UpdatedAt: now, Name: params.Name,
		Email: params.Email, PasswordHash: string(passwordHash),
	})
	if err != nil {
		if strings.Contains(err.Error(), "users_email_key") {
			errorResponse(w, http.StatusConflict, "an account with that email already exists")
			return
		}
		errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("could not create user: %v", err))
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
	for _, feedID := range params.FeedIDs {
		if err := queries.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
			UserID: user.ID, FeedID: feedID,
		}); err != nil {
			errorResponse(w, http.StatusBadRequest, "one or more feed IDs are invalid")
			return
		}
	}

	expiresAt := now.Add(sessionauth.Duration)
	if _, err := queries.CreateSession(r.Context(), database.CreateSessionParams{
		ID: uuid.New(), UserID: user.ID, Token: token, CreatedAt: now, ExpiresAt: expiresAt,
	}); err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not create session")
		return
	}
	if err := tx.Commit(); err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not finish registration")
		return
	}

	sessionauth.SetCookie(w, r, token, expiresAt)
	jsonResponse(w, http.StatusCreated, databaseUserToUser(user))
}

func (ac *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	var params authParameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email := strings.ToLower(strings.TrimSpace(params.Email))
	user, err := ac.DB.GetUserByEmail(r.Context(), email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(params.Password)) != nil {
		errorResponse(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := sessionauth.NewToken()
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not create session")
		return
	}
	now := time.Now().UTC()
	expiresAt := now.Add(sessionauth.Duration)
	if _, err := ac.DB.CreateSession(r.Context(), database.CreateSessionParams{
		ID: uuid.New(), UserID: user.ID, Token: token, CreatedAt: now, ExpiresAt: expiresAt,
	}); err != nil {
		errorResponse(w, http.StatusInternalServerError, "could not create session")
		return
	}

	sessionauth.SetCookie(w, r, token, expiresAt)
	jsonResponse(w, http.StatusOK, databaseUserToUser(user))
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
