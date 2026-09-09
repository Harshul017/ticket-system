package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ticket-system/internal/middleware"
	"ticket-system/internal/models"
	"ticket-system/internal/store"
)

type createTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CreateTicket handles POST /tickets.
func (h *Handlers) CreateTicket(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	var req createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	ticket, err := h.Store.CreateTicket(userID, req.Title, req.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create ticket")
		return
	}

	writeJSON(w, http.StatusCreated, ticket)
}

// ListTickets handles GET /tickets — only the caller's own tickets.
func (h *Handlers) ListTickets(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	tickets, err := h.Store.ListTicketsByUser(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list tickets")
		return
	}

	writeJSON(w, http.StatusOK, tickets)
}

// GetTicket handles GET /tickets/{id}.
// If the ticket doesn't exist OR belongs to someone else, we return the
// same 404 either way — this avoids confirming to a caller that a ticket
// ID exists when it isn't theirs.
func (h *Handlers) GetTicket(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	ticket, err := h.Store.GetTicketByID(id)
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not fetch ticket")
		return
	}

	if ticket.UserID != userID {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

// allowedTransitions defines every valid forward move. A closed ticket
// has no outgoing transitions at all, so it can never be reopened.
var allowedTransitions = map[string][]string{
	models.StatusOpen:       {models.StatusInProgress, models.StatusClosed},
	models.StatusInProgress: {models.StatusClosed},
	models.StatusClosed:     {},
}

func isTransitionAllowed(from, to string) bool {
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// UpdateTicketStatus handles PATCH /tickets/{id}/status.
func (h *Handlers) UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r)

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket id")
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !models.IsValidStatus(req.Status) {
		writeError(w, http.StatusBadRequest, "invalid status value")
		return
	}

	ticket, err := h.Store.GetTicketByID(id)
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "ticket not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not fetch ticket")
		return
	}

	if ticket.UserID != userID {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}

	if !isTransitionAllowed(ticket.Status, req.Status) {
		writeError(w, http.StatusBadRequest, "cannot change status from "+ticket.Status+" to "+req.Status)
		return
	}

	if err := h.Store.UpdateTicketStatus(id, req.Status); err != nil {
		writeError(w, http.StatusInternalServerError, "could not update ticket")
		return
	}

	ticket.Status = req.Status
	ticket.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	writeJSON(w, http.StatusOK, ticket)
}
