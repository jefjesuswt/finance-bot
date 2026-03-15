package telegram

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jefjesuswt/finance-bot/internal/parser"
	"github.com/jefjesuswt/finance-bot/internal/rates"
)

type Handler struct {
	tgService Service
	ratesService rates.Service
	// todo: github.Service

	allowedChatID int64 // id personal de tg
}

func NewHandler(tgs Service, rs rates.Service, acid int64) *Handler {
	return &Handler{
		tgService:    tgs,
		ratesService: rs,
		allowedChatID: acid,
	}
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	defer w.WriteHeader(http.StatusOK)

	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Println("error decoding update.", err)
		return
	}

	if !update.HasTextMessage() {
		log.Println("mensaje vacio.")
		return
	}

	chatID := update.Message.Chat.ID

	log.Printf("🔍 DEBUG - ID Permitido: %d | ID Entrante: %d", h.allowedChatID, chatID)

	if !h.isAuthorized(chatID) {
		log.Printf("⚠️ Intento de uso no autorizado desde el chat ID: %d", chatID)
		return
	}

	ctx := r.Context()
	text := update.Message.Text

	tx, err := parser.Parse(text)
	if err != nil {
		h.tgService.SendMessage(ctx, chatID, "❌ Error de parseo:\n"+err.Error())
		return
	}

	if err := tx.Validate(); err != nil {
		h.tgService.SendMessage(ctx, chatID, "❌ Error de validación:\n"+err.Error())
		return
	}

	if len(tx.Warnings) > 0 {
		for _, warning := range tx.Warnings {
			h.tgService.SendMessage(ctx, chatID, warning)
		}
	}

	currentRates, err := h.ratesService.GetCurrentRates(ctx)
	if err != nil {
		h.tgService.SendMessage(ctx, chatID, "❌ Error obteniendo tasas:\n"+err.Error())
		return
	}

	// markdownFile, err := reports.BuildNote(tx, currentRates)
	// if err != nil {
	// 	h.tgService.SendMessage(ctx, chatID, "💥 Error armando reporte:\n"+err.Error())
	// 	return
	// }

	successMsg := fmt.Sprintf("Transaccion procesada correctamente. con rates: %v", currentRates)
	h.tgService.SendMessage(ctx, chatID, successMsg)
}

func (h *Handler) isAuthorized(chatID int64) bool {
	return h.allowedChatID == 0 || h.allowedChatID == chatID
}
