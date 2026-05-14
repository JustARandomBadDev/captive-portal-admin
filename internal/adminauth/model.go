package adminauth

import "time"

type AdminUser struct {
	ID           string
	Username     string
	PasswordHash string
	DisplayName  string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AdminSession struct {
	ID          string
	AdminUserID string
	TokenHash   string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	LastSeenAt  *time.Time
	RevokedAt   *time.Time
}

type SessionWithAdmin struct {
	Session AdminSession
	Admin   AdminUser
}

type LoginInput struct {
	Username string
	Password string
}
