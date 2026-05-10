package http

import (
	"html/template"
	"net/http"

	"github.com/JustARandomBadDev/captive-portal-admin/internal/config"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/pitches"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/tickets"
)

type Dependencies struct {
	Config    config.Config
	Templates *template.Template
	Tickets   *tickets.Service
	Pitches   *pitches.Service
}

type Router struct {
	cfg       config.Config
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
		templates: deps.Templates,
		tickets:   deps.Tickets,
		pitches:   deps.Pitches,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", router.dashboard)
	mux.HandleFunc("GET /healthz", router.healthz)
	mux.HandleFunc("GET /tickets", router.ticketList)
	mux.HandleFunc("GET /pitches", router.pitchList)

	return mux
}

func (r *Router) dashboard(w http.ResponseWriter, req *http.Request) {
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

func (r *Router) pitchList(w http.ResponseWriter, req *http.Request) {
	r.render(w, "pitches.html", pageData{
		Title:       "Emplacements",
		ActiveNav:   "pitches",
		Heading:     "Emplacements",
		Description: "Preparation de la gestion des emplacements du camping.",
	})
}

func (r *Router) healthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (r *Router) render(w http.ResponseWriter, name string, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := r.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
