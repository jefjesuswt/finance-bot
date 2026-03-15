package rates

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) GetRates(w http.ResponseWriter, r *http.Request) {
	currentRates, err := h.service.GetCurrentRates(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error fetching rates: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(currentRates); err != nil {
		http.Error(w, `{"error": "Error encoding rates"}`, http.StatusInternalServerError)
		return
	}
}
