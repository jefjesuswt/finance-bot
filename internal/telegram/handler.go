package telegram

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jefjesuswt/finance-bot/internal/processor"
)

type Handler struct {
	tgService Service
	processorService processor.Service

	allowedChatID int64 // id personal de tg
}

func NewHandler(tgs Service, proc processor.Service, acid int64) *Handler {
	return &Handler{
		tgService:    tgs,
		processorService: proc,
		allowedChatID: acid,
	}
}

func (h *Handler) isAuthorized(chatID int64) bool {
	return h.allowedChatID == 0 || h.allowedChatID == chatID
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

	result, err := h.processorService.ProcessTransaction(ctx, text)
	if err != nil {
		h.tgService.SendMessage(ctx, chatID, "❌ Error procesando transacción:\n"+err.Error())
		return
	}

	if len(result.Warnings) > 0 {
		for _, warning := range result.Warnings {
			h.tgService.SendMessage(ctx, chatID, warning)
		}
	}

	successMsg := fmt.Sprintf("✅ Transacción lista.\n📂 Destino: %s/%s\n\n%s",
		result.Note.Folder,
		result.Note.Filename,
		result.Note.Content,
	)
	if result.ExtraMessage != "" {
		successMsg += "\n\n" + result.ExtraMessage
	}
	h.tgService.SendMessage(ctx, chatID, successMsg)
}
