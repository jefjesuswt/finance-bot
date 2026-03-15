package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Service interface {
	SendMessage(ctx context.Context, chatID int64, message string) error
}

type botService struct {
	token string
	client *http.Client
}

func NewService(t string, c *http.Client) *botService {
	return &botService{
		token: t,
		client: c,
	}
}

func (s *botService) SendMessage(ctx context.Context, chatID int64, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.token)

	payload := SendMessageReq{
		ChatID: chatID,
		Text: message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
