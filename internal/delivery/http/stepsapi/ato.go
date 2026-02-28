package stepsapi

type putRequestATO struct {
	Count  int    `json:"count" binding:"required"`
	Source string `json:"source" binding:"required"`
}

type addRequestATO struct {
	Date   string `json:"date" binding:"required"`
	Delta  int    `json:"delta" binding:"required"`
	Source string `json:"source" binding:"required"`
}
