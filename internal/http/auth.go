package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/JustARandomBadDev/captive-portal-admin/internal/adminauth"
)

const adminSessionCookieName = "admin_session"

type adminContextKey struct{}

type loginPageData struct {
	viewData
	Title       string
	ActiveNav   string
	Heading     string
	Description string
	Username    string
	Error       string
}

func (r *Router) loginForm(w http.ResponseWriter, req *http.Request) {
	if _, err := r.adminFromCookie(req); err == nil {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	r.render(w, "login.html", loginPageData{
		Title:       "Connexion admin",
		Heading:     "Connexion admin",
		Description: "Accès réservé au panel admin local.",
	})
}

func (r *Router) loginSubmit(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, 16<<10)
	if err := req.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	input := adminauth.LoginInput{
		Username: req.PostFormValue("username"),
		Password: req.PostFormValue("password"),
	}
	result, err := r.adminAuth.Login(req.Context(), input, adminauth.SessionMeta{
		RemoteAddr: req.RemoteAddr,
		UserAgent:  req.UserAgent(),
	})
	if err != nil {
		status := http.StatusUnauthorized
		message := "Identifiant ou mot de passe incorrect."
		if errors.Is(err, adminauth.ErrInactiveAdmin) {
			status = http.StatusForbidden
			message = "Ce compte admin est désactivé."
		} else if !errors.Is(err, adminauth.ErrInvalidCredentials) {
			status = http.StatusInternalServerError
			message = "La connexion a échoué."
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		r.render(w, "login.html", loginPageData{
			Title:       "Connexion admin",
			Heading:     "Connexion admin",
			Description: "Accès réservé au panel admin local.",
			Username:    input.Username,
			Error:       message,
		})
		return
	}

	http.SetCookie(w, r.sessionCookie(result.RawToken, int(r.cfg.AdminSessionTTL.Seconds())))
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func (r *Router) logoutSubmit(w http.ResponseWriter, req *http.Request) {
	if cookie, err := req.Cookie(adminSessionCookieName); err == nil {
		if err := r.adminAuth.Logout(req.Context(), cookie.Value); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(w, r.expiredSessionCookie())
	http.Redirect(w, req, "/login", http.StatusSeeOther)
}

func (r *Router) authMe(w http.ResponseWriter, req *http.Request) {
	admin, ok := CurrentAdmin(req)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":           admin.ID,
		"username":     admin.Username,
		"display_name": admin.DisplayName,
	})
}

func (r *Router) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		admin, err := r.adminFromCookie(req)
		if err != nil {
			http.SetCookie(w, r.expiredSessionCookie())
			if isAPIRequest(req) {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			http.Redirect(w, req, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(req.Context(), adminContextKey{}, admin)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func CurrentAdmin(req *http.Request) (adminauth.AdminUser, bool) {
	admin, ok := req.Context().Value(adminContextKey{}).(adminauth.AdminUser)
	return admin, ok
}

func (r *Router) viewData(req *http.Request) viewData {
	admin, ok := CurrentAdmin(req)
	if !ok {
		return viewData{}
	}
	return viewData{CurrentAdmin: &admin}
}

func (r *Router) adminFromCookie(req *http.Request) (adminauth.AdminUser, error) {
	cookie, err := req.Cookie(adminSessionCookieName)
	if err != nil {
		return adminauth.AdminUser{}, err
	}
	return r.adminAuth.ValidateSession(req.Context(), cookie.Value)
}

func (r *Router) sessionCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     adminSessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.cfg.AdminCookieSecure,
	}
}

func (r *Router) expiredSessionCookie() *http.Cookie {
	cookie := r.sessionCookie("", -1)
	cookie.Expires = time.Unix(0, 0)
	return cookie
}

func isAPIRequest(req *http.Request) bool {
	return len(req.URL.Path) >= len("/api/") && req.URL.Path[:len("/api/")] == "/api/"
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
