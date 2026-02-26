package entities

type SendMessageRequest struct {
	ChatID      int64        `json:"chat_id"`
	Text        string       `json:"text"`
	ParseMode   string       `json:"parse_mode,omitempty"`
	ReplyMarkup *ReplyMarkup `json:"reply_markup,omitempty"`
}

type CallbackQuery struct {
	ID   string `json:"id"`
	Data string `json:"data,omitempty"`
	From *User  `json:"from,omitempty"`
}
