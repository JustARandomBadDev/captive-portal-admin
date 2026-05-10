package http

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/JustARandomBadDev/captive-portal-admin/internal/config"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/database"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/pitches"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/tickets"
)

type Dependencies struct {
	Config    config.Config
	DB        *database.Handle
	Templates *template.Template
	Tickets   *tickets.Service
	Pitches   *pitches.Service
}

type Router struct {
	cfg       config.Config
	db        *database.Handle
	templates *template.Template
	tickets   *tickets.Service
	pitches   *pitches.Service
}

type pageData struct {
	Title              string
	ActiveNav          string
	Heading            string
	Description        string
	DatabaseConfigured bool
}

func NewRouter(deps Dependencies) http.Handler {
	router := &Router{
		cfg:       deps.Config,
		db:        deps.DB,
		templates: deps.Templates,
		tickets:   deps.Tickets,
		pitches:   deps.Pitches,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", router.dashboard)
	mux.HandleFunc("GET /healthz", router.healthz)
	mux.HandleFunc("GET /tickets", router.ticketList)
	mux.HandleFunc("GET /pitches", router.pitchList)
	mux.HandleFunc("GET /pitches/new", router.pitchNew)
	mux.HandleFunc("POST /pitches", router.pitchCreate)
	mux.HandleFunc("POST /pitches/{id}/disable", router.pitchDisable)

	return mux
}

func (r *Router) dashboard(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	r.render(w, "dashboard.html", pageData{
		Title:              "Camping WiFi Admin",
		ActiveNav:          "dashboard",
		Heading:            "Dashboard",
		Description:        "Vue d'ensemble du panel admin local.",
		DatabaseConfigured: r.cfg.DatabaseURL != "",
	})
}

func (r *Router) ticketList(w http.ResponseWriter, req *http.Request) {
	r.render(w, "tickets.html", pageData{
		Title:       "Tickets WiFi",
		ActiveNav:   "tickets",
		Heading:     "Tickets WiFi",
		Description: "Preparation de la gestion des tickets temporaires et de leur synchronisation FreeRADIUS.",
	})
}

func (r *Router) healthz(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()
	if err := r.db.Ping(ctx); err != nil {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (r *Router) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := r.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
