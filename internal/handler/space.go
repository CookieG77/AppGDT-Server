package handler

import (
	"net/http"

	"github.com/CookieG77/AppGDT-Server/internal/httpjson"
	"github.com/CookieG77/AppGDT-Server/internal/service"
)

// SpaceHandler exposes the routes managing the spaces of the authenticated user.
type SpaceHandler struct {
	spaces *service.SpaceService
}

// NewSpaceHandler creates a SpaceHandler using the given service.
func NewSpaceHandler(spaces *service.SpaceService) *SpaceHandler {
	return &SpaceHandler{spaces: spaces}
}

// spaceInput is the request body of the creation and update routes,
// as defined by the SpaceInput schema of the API contract.
type spaceInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// List handles GET /spaces.
func (h *SpaceHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	spaces, err := h.spaces.List(r.Context(), userID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, spaces)
}

// Create handles POST /spaces.
func (h *SpaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	var input spaceInput
	if err := httpjson.DecodeJSON(w, r, &input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	space, err := h.spaces.Create(r.Context(), userID, input.Name, input.Description)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusCreated, space)
}

// Get handles GET /spaces/{spaceId}.
func (h *SpaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	spaceID, ok := pathID(w, r, "spaceId")
	if !ok {
		return
	}

	space, err := h.spaces.Get(r.Context(), userID, spaceID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, space)
}

// Update handles PUT /spaces/{spaceId}.
func (h *SpaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	spaceID, ok := pathID(w, r, "spaceId")
	if !ok {
		return
	}

	var input spaceInput
	if err := httpjson.DecodeJSON(w, r, &input); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	space, err := h.spaces.Update(r.Context(), userID, spaceID, input.Name, input.Description)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, space)
}

// Delete handles DELETE /spaces/{spaceId}.
func (h *SpaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	spaceID, ok := pathID(w, r, "spaceId")
	if !ok {
		return
	}

	if err := h.spaces.Delete(r.Context(), userID, spaceID); err != nil {
		writeServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
