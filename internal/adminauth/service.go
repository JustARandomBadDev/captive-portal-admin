package adminauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultSessionTTL = 12 * time.Hour
	sessionTokenSize  = 32
)

const dummyPasswordHash = "$2a$10$65IWc3thDKBv/Q3KEDHFa.G8EPPNRT6swNeS/CqnCD7jyawDkXJkS"

var (
	ErrInvalidCredentials = errors.New("invalid admin credentials")
	ErrInactiveAdmin      = errors.New("admin user is inactive")
	ErrInvalidSession     = errors.New("invalid admin session")
	ErrExpiredSession     = errors.New("expired admin session")
	ErrMissingRepository  = errors.New("admin auth repository is required")
)

type SessionMeta struct {
	RemoteAddr string
	UserAgent  string
}

type LoginResult struct {
	RawToken string
	Admin    AdminUser
}

type Service struct {
	repository    Repository
	sessionSecret string
	sessionTTL    time.Duration
	now           func() time.Time
}

func NewService(sessionSecret string) *Service {
	return &Service{
		sessionSecret: sessionSecret,
		sessionTTL:    DefaultSessionTTL,
		now:           time.Now,
	}
}

func NewServiceWithRepository(repository Repository, sessionTTL time.Duration) *Service {
	if sessionTTL <= 0 {
		sessionTTL = DefaultSessionTTL
	}

	return &Service{
		repository: repository,
		sessionTTL: sessionTTL,
		now:        time.Now,
	}
}

func (s *Service) Login(ctx context.Context, input LoginInput, meta SessionMeta) (LoginResult, error) {
	_ = meta

	if s.repository == nil {
		return LoginResult{}, ErrMissingRepository
	}

	username := strings.TrimSpace(input.Username)
	password := input.Password
	if username == "" || strings.TrimSpace(password) == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	admin, err := s.repository.GetAdminByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrAdminNotFound) {
			VerifyPassword(password, dummyPasswordHash)
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}
	if !admin.IsActive {
		return LoginResult{}, ErrInactiveAdmin
	}
	if !VerifyPassword(password, admin.PasswordHash) {
		return LoginResult{}, ErrInvalidCredentials
	}

	rawToken, err := generateSessionToken()
	if err != nil {
		return LoginResult{}, err
	}

	now := s.now()
	_, err = s.repository.CreateSession(ctx, AdminSession{
		AdminUserID: admin.ID,
		TokenHash:   HashToken(rawToken),
		ExpiresAt:   now.Add(s.sessionTTL),
	})
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		RawToken: rawToken,
		Admin:    admin,
	}, nil
}

func (s *Service) ValidateSession(ctx context.Context, rawToken string) (AdminUser, error) {
	if s.repository == nil {
		return AdminUser{}, ErrMissingRepository
	}

	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return AdminUser{}, ErrInvalidSession
	}

	sessionWithAdmin, err := s.repository.GetSessionByTokenHash(ctx, HashToken(rawToken))
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return AdminUser{}, ErrInvalidSession
		}
		return AdminUser{}, err
	}

	if sessionWithAdmin.Session.RevokedAt != nil {
		return AdminUser{}, ErrInvalidSession
	}
	if !sessionWithAdmin.Session.ExpiresAt.After(s.now()) {
		return AdminUser{}, ErrExpiredSession
	}
	if !sessionWithAdmin.Admin.IsActive {
		return AdminUser{}, ErrInactiveAdmin
	}

	if err := s.repository.TouchSession(ctx, sessionWithAdmin.Session.ID); err != nil && !errors.Is(err, ErrSessionNotFound) {
		return AdminUser{}, err
	}

	return sessionWithAdmin.Admin, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil
	}
	if s.repository == nil {
		return ErrMissingRepository
	}

	sessionWithAdmin, err := s.repository.GetSessionByTokenHash(ctx, HashToken(rawToken))
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil
		}
		return err
	}

	if err := s.repository.RevokeSession(ctx, sessionWithAdmin.Session.ID); err != nil && !errors.Is(err, ErrSessionNotFound) {
		return err
	}

	return nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

func generateSessionToken() (string, error) {
	token := make([]byte, sessionTokenSize)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(token), nil
}
