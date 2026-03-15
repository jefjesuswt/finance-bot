package telegram

import (
	"github.com/go-chi/chi/v5"
)

func Routes(h *Handler) chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.HandleWebhook)

	return r
}
