package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/peterintech/briefed/components"
	"github.com/peterintech/briefed/internal/accounts"
	"github.com/peterintech/briefed/internal/sessionauth"
	webtypes "github.com/peterintech/briefed/types"
	"github.com/peterintech/briefed/views"
)

func (h *Handler) RegisterAuthRoutes(router *chi.Mux) {
	router.Get("/login", h.loginPage)
	router.Post("/auth/login", h.login)
	router.Post("/auth/logout", h.logout)
	router.Post("/auth/register", h.register)
	router.Get("/partials/onboarding/interests", h.onboardingInterests)
	router.Get("/partials/onboarding/feeds", h.onboardingFeeds)
	router.Get("/partials/onboarding/account", h.onboardingAccount)
}

func (h *Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	user, _, err := h.viewer(r)
	if err != nil {
		http.Error(w, "could not load login", http.StatusInternalServerError)
		return
	}
	if user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	categories, err := h.DB.GetCategories(r.Context())
	if err != nil {
		http.Error(w, "could not load login", http.StatusInternalServerError)
		return
	}
	render(w, r, http.StatusOK, views.Login(webtypes.LoginData{Categories: categories}))
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		render(w, r, http.StatusOK, components.InlineError("Check the form and try again."))
		return
	}
	result, err := accounts.Login(r.Context(), h.DB, r.FormValue("email"), r.FormValue("password"))
	if errors.Is(err, accounts.ErrInvalidCredentials) {
		render(w, r, http.StatusOK, components.InlineError("That email and password do not match."))
		return
	}
	if err != nil {
		render(w, r, http.StatusOK, components.InlineError("We could not sign you in. Try again."))
		return
	}
	sessionauth.SetCookie(w, r, result.Token, result.ExpiresAt)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionauth.CookieName); err == nil && cookie.Value != "" {
		_ = h.DB.DeleteSessionByToken(r.Context(), cookie.Value)
	}
	sessionauth.ClearCookie(w, r)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid registration", http.StatusBadRequest)
		return
	}
	categoryIDs, categoryErr := parseUUIDs(r.Form["category_ids"])
	feedIDs, feedErr := parseUUIDs(r.Form["feed_ids"])
	data := webtypes.AccountStepData{CategoryIDs: categoryIDs, FeedIDs: feedIDs, Name: r.FormValue("name"), Email: r.FormValue("email")}
	if categoryErr != nil || feedErr != nil || len(categoryIDs) == 0 || len(feedIDs) == 0 {
		data.Error = "Your selected interests or sources are invalid. Please go back and choose again."
		render(w, r, http.StatusOK, components.AccountStep(data))
		return
	}
	result, err := accounts.Register(r.Context(), h.Conn, h.DB, accounts.RegisterInput{Name: data.Name, Email: data.Email, Password: r.FormValue("password"), CategoryIDs: categoryIDs, FeedIDs: feedIDs})
	switch {
	case errors.Is(err, accounts.ErrEmailExists):
		data.Error = "An account with that email already exists."
	case errors.Is(err, accounts.ErrInvalidRegistration):
		data.Error = "Add your name, a valid email, and a password with at least 8 characters."
	case errors.Is(err, accounts.ErrInvalidChoices):
		data.Error = "One of your selected interests or sources is no longer available."
	case err != nil:
		data.Error = "We could not create your account. Try again."
	}
	if err != nil {
		render(w, r, http.StatusOK, components.AccountStep(data))
		return
	}
	sessionauth.SetCookie(w, r, result.Token, result.ExpiresAt)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
