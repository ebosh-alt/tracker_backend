package shared

import "errors"

var (
	// ErrInvalidInput возвращается, когда входные данные нарушают доменные правила.
	ErrInvalidInput = errors.New("invalid input")
	// ErrUnauthorized возвращается, когда субъект не прошел проверку доступа.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrNotFound возвращается, когда запрошенная сущность не найдена.
	ErrNotFound = errors.New("not found")
	// ErrConflict возвращается, когда операция нарушает уникальность/состояние.
	ErrConflict = errors.New("conflict")
)
