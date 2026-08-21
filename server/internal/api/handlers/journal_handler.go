package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/lnardon/arete/internal/auth"
	"github.com/lnardon/arete/internal/repository"
)

const (
	defaultJournalListLimit = 30
	maxJournalListLimit     = 365
	maxJournalContentLength = 10000
)

type JournalHandler struct {
	repo *repository.JournalRepository
}

func NewJournalHandler(repo *repository.JournalRepository) *JournalHandler {
	return &JournalHandler{repo: repo}
}

func parseEntryDate(s string) (string, bool) {
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return "", false
	}
	return s, true
}

func (h *JournalHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	limit := defaultJournalListLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > maxJournalListLimit {
			http.Error(w, "limit must be a positive integer up to 365", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	entries, err := h.repo.ListEntries(r.Context(), authUser.ID, limit)
	if err != nil {
		http.Error(w, "failed to list journal entries", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *JournalHandler) GetEntry(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	date, ok := parseEntryDate(mux.Vars(r)["date"])
	if !ok {
		http.Error(w, "date must be in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}

	entry, err := h.repo.GetEntry(r.Context(), authUser.ID, date)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "journal entry not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get journal entry", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

// UpsertEntry creates the entry for {date} or overwrites it if one already
// exists — journal entries are one-per-day, so the frontend can always PUT
// without checking whether today's entry exists yet.
func (h *JournalHandler) UpsertEntry(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	date, ok := parseEntryDate(mux.Vars(r)["date"])
	if !ok {
		http.Error(w, "date must be in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}

	var body struct {
		Mood    int    `json:"mood"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Mood < 1 || body.Mood > 5 {
		http.Error(w, "mood must be between 1 and 5", http.StatusBadRequest)
		return
	}
	if body.Content == "" || len(body.Content) > maxJournalContentLength {
		http.Error(w, "content is required and must be 10000 characters or fewer", http.StatusBadRequest)
		return
	}

	entry, err := h.repo.UpsertEntry(r.Context(), authUser.ID, date, body.Mood, body.Content)
	if err != nil {
		http.Error(w, "failed to save journal entry", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

func (h *JournalHandler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	date, ok := parseEntryDate(mux.Vars(r)["date"])
	if !ok {
		http.Error(w, "date must be in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteEntry(r.Context(), authUser.ID, date); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "journal entry not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete journal entry", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
