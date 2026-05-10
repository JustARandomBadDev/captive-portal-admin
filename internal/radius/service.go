package radius

import "github.com/JustARandomBadDev/captive-portal-admin/internal/database"

type Service struct {
	db *database.Handle
}

func NewService(db *database.Handle) *Service {
	return &Service{db: db}
}
