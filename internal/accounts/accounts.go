package accounts

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
	"github.com/peterintech/briefed/internal/sessionauth"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidRegistration = errors.New("name, email, and a password of at least 8 characters are required")
	ErrEmailExists         = errors.New("an account with that email already exists")
	ErrInvalidChoices      = errors.New("one or more category or feed IDs are invalid")
	ErrInvalidCredentials  = errors.New("invalid email or password")
)

type RegisterInput struct {
	Name        string
	Email       string
	Password    string
	CategoryIDs []uuid.UUID
	FeedIDs     []uuid.UUID
}

type Session struct {
	User      database.User
	Token     string
	ExpiresAt time.Time
}

func Register(ctx context.Context, conn *sql.DB, db *database.Queries, input RegisterInput) (Session, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Name == "" || input.Email == "" || len(input.Password) < 8 {
		return Session{}, ErrInvalidRegistration
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return Session{}, err
	}
	token, err := sessionauth.NewToken()
	if err != nil {
		return Session{}, err
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, err
	}
	defer tx.Rollback()

	queries := db.WithTx(tx)
	now := time.Now().UTC()
	user, err := queries.CreateUser(ctx, database.CreateUserParams{
		ID: uuid.New(), CreatedAt: now, UpdatedAt: now, Name: input.Name,
		Email: input.Email, PasswordHash: string(passwordHash),
	})
	if err != nil {
		if strings.Contains(err.Error(), "users_email_key") {
			return Session{}, ErrEmailExists
		}
		return Session{}, err
	}

	for _, categoryID := range input.CategoryIDs {
		if err := queries.CreateUserCategory(ctx, database.CreateUserCategoryParams{
			UserID: user.ID, CategoryID: categoryID,
		}); err != nil {
			return Session{}, ErrInvalidChoices
		}
	}
	for _, feedID := range input.FeedIDs {
		if err := queries.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
			UserID: user.ID, FeedID: feedID,
		}); err != nil {
			return Session{}, ErrInvalidChoices
		}
	}

	expiresAt := now.Add(sessionauth.Duration)
	if _, err := queries.CreateSession(ctx, database.CreateSessionParams{
		ID: uuid.New(), UserID: user.ID, Token: token, CreatedAt: now, ExpiresAt: expiresAt,
	}); err != nil {
		return Session{}, err
	}
	if err := tx.Commit(); err != nil {
		return Session{}, err
	}
	return Session{User: user, Token: token, ExpiresAt: expiresAt}, nil
}

func Login(ctx context.Context, db *database.Queries, email, password string) (Session, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := db.GetUserByEmail(ctx, email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return Session{}, ErrInvalidCredentials
	}
	token, err := sessionauth.NewToken()
	if err != nil {
		return Session{}, err
	}
	now := time.Now().UTC()
	expiresAt := now.Add(sessionauth.Duration)
	if _, err := db.CreateSession(ctx, database.CreateSessionParams{
		ID: uuid.New(), UserID: user.ID, Token: token, CreatedAt: now, ExpiresAt: expiresAt,
	}); err != nil {
		return Session{}, err
	}
	return Session{User: user, Token: token, ExpiresAt: expiresAt}, nil
}
