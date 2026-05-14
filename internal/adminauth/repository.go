package adminauth

import "context"

type Repository interface {
	GetAdminByUsername(ctx context.Context, username string) (AdminUser, error)
	GetAdminByID(ctx context.Context, id string) (AdminUser, error)
	CreateSession(ctx context.Context, session AdminSession) (AdminSession, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (SessionWithAdmin, error)
	RevokeSession(ctx context.Context, sessionID string) error
	TouchSession(ctx context.Context, sessionID string) error
}
