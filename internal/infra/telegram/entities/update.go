package entities

type ApiResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

type Update struct {
	UpdateID      int64          `json:"update_id"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

type SetWebhookRequest struct {
	URL         string `json:"url"`
	SecretToken string `json:"secret_token,omitempty"`
}
