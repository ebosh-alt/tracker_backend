package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"tracker/internal/infra/telegram/entities"
)

func (cl *Client) SetWebhook(url, secret string) error {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", cl.token)
	payload := entities.SetWebhookRequest{URL: url, SecretToken: secret}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
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
