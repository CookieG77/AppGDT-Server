package handler

import (
	"net/http"

	"github.com/CookieG77/AppGDT-Server/internal/domain"
	"github.com/CookieG77/AppGDT-Server/internal/httpjson"
	"github.com/CookieG77/AppGDT-Server/internal/service"
)

// NoteHandler exposes the routes managing the notes of the authenticated user.
type NoteHandler struct {
	notes *service.NoteService
}

// NewNoteHandler creates a NoteHandler using the given service.
func NewNoteHandler(notes *service.NoteService) *NoteHandler {
	return &NoteHandler{notes: notes}
}

// noteInput is the request body of the creation and update routes,
// as defined by the NoteInput schema of the API contract.
type noteInput struct {
	Title   string             `json:"title"`
	Content string             `json:"content"`
	Status  *domain.NoteStatus `json:"status"`
}

// toService converts the request body into the input expected by the service.
func (in noteInput) toService() service.NoteInput {
	return service.NoteInput{Title: in.Title, Content: in.Content, Status: in.Status}
}

// List handles GET /spaces/{spaceId}/notes.
func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	spaceID, ok := pathID(w, r, "spaceId")
	if !ok {
		return
	}

	notes, err := h.notes.List(r.Context(), userID, spaceID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, notes)
}

// Create handles POST /spaces/{spaceId}/notes.
func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	spaceID, ok := pathID(w, r, "spaceId")
	if !ok {
		return
	}

	var input noteInput
	if err := httpjson.DecodeJSON(w, r, &input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	note, err := h.notes.Create(r.Context(), userID, spaceID, input.toService())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusCreated, note)
}

// Get handles GET /notes/{noteId}.
func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	noteID, ok := pathID(w, r, "noteId")
	if !ok {
		return
	}

	note, err := h.notes.Get(r.Context(), userID, noteID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, note)
}

// Update handles PUT /notes/{noteId}.
func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	noteID, ok := pathID(w, r, "noteId")
	if !ok {
		return
	}

	var input noteInput
	if err := httpjson.DecodeJSON(w, r, &input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	note, err := h.notes.Update(r.Context(), userID, noteID, input.toService())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, note)
}

// Delete handles DELETE /notes/{noteId}.
func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	noteID, ok := pathID(w, r, "noteId")
	if !ok {
		return
	}

	if err := h.notes.Delete(r.Context(), userID, noteID); err != nil {
		writeServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
