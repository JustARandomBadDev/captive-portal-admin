package app

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/JustARandomBadDev/captive-portal-admin/internal/adminauth"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/config"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/database"
	adminhttp "github.com/JustARandomBadDev/captive-portal-admin/internal/http"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/pitches"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/radius"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/templates"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/tickets"
)

type App struct {
	Config     config.Config
	DB         *database.Handle
	Server     *http.Server
	Tickets    *tickets.Service
	Pitches    *pitches.Service
	Radius     *radius.Service
	AdminAuth  *adminauth.Service
	Templates  *template.Template
	HTTPRouter http.Handler
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	views, err := template.ParseFS(templates.FS, "*.html")
	if err != nil {
		return nil, err
	}

	db, err := database.Connect(ctx, database.Config{URL: cfg.DatabaseURL})
	if err != nil {
		return nil, err
	}
	ticketRepository := tickets.NewPostgresRepository(db)
	pitchRepository := pitches.NewPostgresRepository(db)
	radiusService := radius.NewService(radius.NoopSyncer{})

	app := &App{
		Config:    cfg,
		DB:        db,
		Tickets:   tickets.NewService(ticketRepository, radiusService),
		Pitches:   pitches.NewService(pitchRepository),
		Radius:    radiusService,
		AdminAuth: adminauth.NewService(cfg.SessionSecret),
		Templates: views,
	}

	handler := adminhttp.NewRouter(adminhttp.Dependencies{
		Config:    cfg,
		DB:        db,
		Templates: views,
		Tickets:   app.Tickets,
		Pitches:   app.Pitches,
	})

	app.HTTPRouter = handler
	app.Server = &http.Server{
		Addr:              cfg.AppAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return app, nil
}

func (a *App) Close() {
	a.DB.Close()
}
