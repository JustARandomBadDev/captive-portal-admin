package adminauth

type Service struct {
	sessionSecret string
}

func NewService(sessionSecret string) *Service {
	return &Service{sessionSecret: sessionSecret}
}
