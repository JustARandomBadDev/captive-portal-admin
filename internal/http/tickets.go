package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/JustARandomBadDev/captive-portal-admin/internal/pitches"
	"github.com/JustARandomBadDev/captive-portal-admin/internal/tickets"
)

type ticketListPageData struct {
	Title       string
	ActiveNav   string
	Heading     string
	Description string
	Tickets     []ticketRow
	Error       string
}

type ticketFormPageData struct {
	Title         string
	ActiveNav     string
	Heading       string
	Description   string
	Pitches       []pitches.Pitch
	PitchID       string
	DurationHours int
	Error         string
}

type ticketRow struct {
	ID         string
	Username   string
	Password   string
	PitchCode  string
	Status     string
	ValidUntil string
	CanRevoke  bool
}

func (r *Router) ticketList(w http.ResponseWriter, req *http.Request) {
	allTickets, err := r.tickets.ListAll(req.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	allPitches, err := r.pitches.ListAll(req.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	r.render(w, "tickets.html", ticketListPageData{
		Title:       "Tickets WiFi",
		ActiveNav:   "tickets",
		Heading:     "Tickets WiFi",
		Description: "Gestion des tickets WiFi temporaires.",
		Tickets:     buildTicketRows(allTickets, allPitches),
		Error:       req.URL.Query().Get("error"),
	})
}

func (r *Router) ticketNew(w http.ResponseWriter, req *http.Request) {
	r.renderTicketForm(w, req, ticketFormPageData{
		Title:         "Nouveau ticket",
		ActiveNav:     "tickets",
		Heading:       "Nouveau ticket WiFi",
		Description:   "Créer un ticket temporaire pour un emplacement.",
		DurationHours: 24,
	})
}

func (r *Router) ticketCreate(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	pitchID := req.PostFormValue("pitch_id")
	durationHours, err := strconv.Atoi(req.PostFormValue("duration_hours"))
	if err != nil || durationHours <= 0 {
		r.renderTicketCreateError(w, req, ticketFormPageData{
			PitchID:       pitchID,
			DurationHours: durationHours,
			Error:         "La durée doit être un nombre d'heures positif.",
		}, http.StatusBadRequest)
		return
	}

	validFrom := time.Now()
	input := tickets.TicketCreateInput{
		PitchID:    pitchID,
		ValidFrom:  validFrom,
		ValidUntil: validFrom.Add(time.Duration(durationHours) * time.Hour),
	}
	if _, err := r.tickets.Create(req.Context(), input); err != nil {
		status := http.StatusBadRequest
		message := ticketCreateError(err)
		if message == "" {
			message = "Le ticket n'a pas pu être créé."
			status = http.StatusInternalServerError
		}
		r.renderTicketCreateError(w, req, ticketFormPageData{
			PitchID:       pitchID,
			DurationHours: durationHours,
			Error:         message,
		}, status)
		return
	}

	http.Redirect(w, req, "/tickets", http.StatusSeeOther)
}

func (r *Router) ticketRevoke(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		http.NotFound(w, req)
		return
	}

	if _, err := r.tickets.Revoke(req.Context(), tickets.TicketRevokeInput{ID: id}); err != nil {
		if errors.Is(err, tickets.ErrTicketNotFound) {
			http.NotFound(w, req)
			return
		}
		http.Redirect(w, req, "/tickets?error=revoke", http.StatusSeeOther)
		return
	}

	http.Redirect(w, req, "/tickets", http.StatusSeeOther)
}

func (r *Router) renderTicketCreateError(w http.ResponseWriter, req *http.Request, data ticketFormPageData, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	r.renderTicketForm(w, req, data)
}

func (r *Router) renderTicketForm(w http.ResponseWriter, req *http.Request, data ticketFormPageData) {
	activePitches, err := r.pitches.ListActive(req.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if data.Title == "" {
		data.Title = "Nouveau ticket"
	}
	if data.ActiveNav == "" {
		data.ActiveNav = "tickets"
	}
	if data.Heading == "" {
		data.Heading = "Nouveau ticket WiFi"
	}
	if data.Description == "" {
		data.Description = "Créer un ticket temporaire pour un emplacement."
	}
	if data.DurationHours == 0 {
		data.DurationHours = 24
	}
	data.Pitches = activePitches

	r.render(w, "ticket_new.html", data)
}

func buildTicketRows(items []tickets.Ticket, allPitches []pitches.Pitch) []ticketRow {
	pitchCodes := make(map[string]string, len(allPitches))
	for _, pitch := range allPitches {
		pitchCodes[pitch.ID] = pitch.Code
	}

	rows := make([]ticketRow, 0, len(items))
	for _, ticket := range items {
		rows = append(rows, ticketRow{
			ID:         ticket.ID,
			Username:   ticket.Username,
			Password:   ticket.CleartextPassword,
			PitchCode:  pitchLabel(ticket.PitchID, pitchCodes),
			Status:     ticketStatusLabel(ticket.Status),
			ValidUntil: ticket.ValidUntil.Format("02/01/2006 15:04"),
			CanRevoke:  ticket.Status == tickets.TicketStatusActive,
		})
	}

	return rows
}

func pitchLabel(id string, pitchCodes map[string]string) string {
	if code, ok := pitchCodes[id]; ok {
		return code
	}
	return id
}

func ticketStatusLabel(status tickets.TicketStatus) string {
	switch status {
	case tickets.TicketStatusActive:
		return "Actif"
	case tickets.TicketStatusExpired:
		return "Expire"
	case tickets.TicketStatusRevoked:
		return "Revoque"
	default:
		return string(status)
	}
}

func ticketCreateError(err error) string {
	switch {
	case errors.Is(err, tickets.ErrPitchRequired):
		return "Selectionnez un emplacement."
	case errors.Is(err, tickets.ErrInvalidTicketDates):
		return "Les dates de validité sont incohérentes."
	case errors.Is(err, tickets.ErrDuplicateUsername):
		return "Un identifiant identique existe déjà. Réessayez."
	default:
		return ""
	}
}
