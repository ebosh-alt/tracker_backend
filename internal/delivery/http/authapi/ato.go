package authapi

// telegramAuthRequestATO формат входного payload для POST /api/auth/telegram.
type telegramAuthRequestATO struct {
	InitData string `json:"initData" binding:"required"`
}

type refreshRequestATO struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}
