package adminauth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestHashPasswordAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("secret-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "secret-password" {
		t.Fatal("expected password hash, got cleartext password")
	}
	if !VerifyPassword("secret-password", hash) {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword("wrong-password", hash) {
		t.Fatal("expected wrong password to fail")
	}
}

func TestLoginOK(t *testing.T) {
	repository := newFakeRepository(t, true)
	service := newTestService(repository)

	result, err := service.Login(context.Background(), LoginInput{
		Username: "admin",
		Password: "secret-password",
	}, SessionMeta{RemoteAddr: "127.0.0.1", UserAgent: "test"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if result.RawToken == "" {
		t.Fatal("expected raw session token")
	}
	if result.Admin.Username != "admin" {
		t.Fatalf("expected admin username, got %q", result.Admin.Username)
	}
	if len(repository.sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(repository.sessions))
	}
	for _, session := range repository.sessions {
		if session.TokenHash == result.RawToken {
			t.Fatal("expected stored token hash, got raw token")
		}
		if session.TokenHash != HashToken(result.RawToken) {
			t.Fatal("expected stored token hash to match raw token")
		}
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	repository := newFakeRepository(t, true)
	service := newTestService(repository)

	_, err := service.Login(context.Background(), LoginInput{
		Username: "admin",
		Password: "wrong-password",
	}, SessionMeta{})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginRejectsInactiveAdmin(t *testing.T) {
	repository := newFakeRepository(t, false)
	service := newTestService(repository)

	_, err := service.Login(context.Background(), LoginInput{
		Username: "admin",
		Password: "secret-password",
	}, SessionMeta{})
	if !errors.Is(err, ErrInactiveAdmin) {
		t.Fatalf("expected ErrInactiveAdmin, got %v", err)
	}
}

func TestValidateSessionOK(t *testing.T) {
	repository := newFakeRepository(t, true)
	service := newTestService(repository)
	rawToken := "raw-session-token"
	session := repository.addSession(repository.admin.ID, HashToken(rawToken), service.now().Add(time.Hour), nil)

	admin, err := service.ValidateSession(context.Background(), rawToken)
	if err != nil {
		t.Fatalf("validate session: %v", err)
	}
	if admin.ID != repository.admin.ID {
		t.Fatalf("expected admin %q, got %q", repository.admin.ID, admin.ID)
	}
	if repository.touchedID != session.ID {
		t.Fatalf("expected touched session %q, got %q", session.ID, repository.touchedID)
	}
}

func TestValidateSessionRejectsExpiredSession(t *testing.T) {
	repository := newFakeRepository(t, true)
	service := newTestService(repository)
	rawToken := "raw-session-token"
	repository.addSession(repository.admin.ID, HashToken(rawToken), service.now().Add(-time.Minute), nil)

	_, err := service.ValidateSession(context.Background(), rawToken)
	if !errors.Is(err, ErrExpiredSession) {
		t.Fatalf("expected ErrExpiredSession, got %v", err)
	}
}

func TestValidateSessionRejectsRevokedSession(t *testing.T) {
	repository := newFakeRepository(t, true)
	service := newTestService(repository)
	rawToken := "raw-session-token"
	revokedAt := service.now().Add(-time.Minute)
	repository.addSession(repository.admin.ID, HashToken(rawToken), service.now().Add(time.Hour), &revokedAt)

	_, err := service.ValidateSession(context.Background(), rawToken)
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected ErrInvalidSession, got %v", err)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	repository := newFakeRepository(t, true)
	service := newTestService(repository)
	rawToken := "raw-session-token"
	session := repository.addSession(repository.admin.ID, HashToken(rawToken), service.now().Add(time.Hour), nil)

	if err := service.Logout(context.Background(), rawToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if repository.revokedID != session.ID {
		t.Fatalf("expected revoked session %q, got %q", session.ID, repository.revokedID)
	}
	if repository.sessions[session.ID].RevokedAt == nil {
		t.Fatal("expected revoked_at to be set")
	}
}

func TestLogoutAllowsEmptyToken(t *testing.T) {
	service := newTestService(newFakeRepository(t, true))
	if err := service.Logout(context.Background(), " "); err != nil {
		t.Fatalf("expected empty logout to succeed, got %v", err)
	}
}

func newTestService(repository *fakeRepository) *Service {
	service := NewServiceWithRepository(repository, 2*time.Hour)
	service.now = func() time.Time {
		return time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	}
	return service
}

func newFakeRepository(t *testing.T, active bool) *fakeRepository {
	t.Helper()

	hash, err := HashPassword("secret-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	admin := AdminUser{
		ID:           "admin-id",
		Username:     "admin",
		PasswordHash: hash,
		DisplayName:  "Admin",
		IsActive:     active,
		CreatedAt:    time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC),
	}

	return &fakeRepository{
		admin:    admin,
		sessions: make(map[string]AdminSession),
	}
}

type fakeRepository struct {
	admin     AdminUser
	sessions  map[string]AdminSession
	nextID    int
	touchedID string
	revokedID string
}

func (r *fakeRepository) GetAdminByUsername(ctx context.Context, username string) (AdminUser, error) {
	if username != r.admin.Username {
		return AdminUser{}, ErrAdminNotFound
	}
	return r.admin, nil
}

func (r *fakeRepository) GetAdminByID(ctx context.Context, id string) (AdminUser, error) {
	if id != r.admin.ID {
		return AdminUser{}, ErrAdminNotFound
	}
	return r.admin, nil
}

func (r *fakeRepository) CreateSession(ctx context.Context, session AdminSession) (AdminSession, error) {
	r.nextID++
	if session.ID == "" {
		session.ID = "session-id"
		if r.nextID > 1 {
			session.ID = "session-id-2"
		}
	}
	session.CreatedAt = time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	r.sessions[session.ID] = session
	return session, nil
}

func (r *fakeRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (SessionWithAdmin, error) {
	for _, session := range r.sessions {
		if session.TokenHash == tokenHash {
			admin := r.admin
			admin.ID = session.AdminUserID
			return SessionWithAdmin{Session: session, Admin: admin}, nil
		}
	}
	return SessionWithAdmin{}, ErrSessionNotFound
}

func (r *fakeRepository) RevokeSession(ctx context.Context, sessionID string) error {
	session, ok := r.sessions[sessionID]
	if !ok || session.RevokedAt != nil {
		return ErrSessionNotFound
	}
	now := time.Date(2026, 5, 13, 12, 30, 0, 0, time.UTC)
	session.RevokedAt = &now
	r.sessions[sessionID] = session
	r.revokedID = sessionID
	return nil
}

func (r *fakeRepository) TouchSession(ctx context.Context, sessionID string) error {
	session, ok := r.sessions[sessionID]
	if !ok || session.RevokedAt != nil {
		return ErrSessionNotFound
	}
	now := time.Date(2026, 5, 13, 12, 30, 0, 0, time.UTC)
	session.LastSeenAt = &now
	r.sessions[sessionID] = session
	r.touchedID = sessionID
	return nil
}

func (r *fakeRepository) addSession(adminID, tokenHash string, expiresAt time.Time, revokedAt *time.Time) AdminSession {
	session := AdminSession{
		ID:          "session-id",
		AdminUserID: adminID,
		TokenHash:   tokenHash,
		CreatedAt:   time.Date(2026, 5, 13, 11, 0, 0, 0, time.UTC),
		ExpiresAt:   expiresAt,
		RevokedAt:   revokedAt,
	}
	r.sessions[session.ID] = session
	return session
}
