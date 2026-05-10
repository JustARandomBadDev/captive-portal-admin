package database

type Config struct {
	URL string
}

type Handle struct {
	config Config
}

func New(config Config) *Handle {
	return &Handle{config: config}
}

func (h *Handle) Config() Config {
	return h.config
}
