package sessionauth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/peterintech/briefed/internal/database"
)

const (
	CookieName = "briefed_session"
	Duration   = 30 * 24 * time.Hour
)

func NewToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func CurrentUser(ctx context.Context, db *database.Queries, r *http.Request) (*database.User, error) {
	cookie, err := r.Cookie(CookieName)
	if err == http.ErrNoCookie || (err == nil && cookie.Value == "") {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	user, err := db.GetUserBySessionToken(ctx, cookie.Value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func SetCookie(w http.ResponseWriter, r *http.Request, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name: CookieName, Value: token, Path: "/", Expires: expiresAt,
		MaxAge: int(time.Until(expiresAt).Seconds()), HttpOnly: true,
		Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode,
	})
}

func ClearCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: CookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode,
	})
}
