package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tracker/internal/infra/telegram/entities"
)

type Client struct {
	token  string
	client *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (cl *Client) SendMessage(chatID int64, text string, markup *entities.ReplyMarkup) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cl.token)
	payload := entities.SendMessageRequest{ChatID: chatID, Text: text, ReplyMarkup: markup}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cl.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResp entities.ApiResponse
	_ = json.NewDecoder(resp.Body).Decode(&apiResp)
	if !apiResp.OK {
		return fmt.Errorf("telegram error: %s", apiResp.Description)
	}

	return nil
}
