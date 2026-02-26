package clock

import "time"

// SystemClock инфраструктурная реализация источника времени.
type SystemClock struct{}

// Now возвращает текущее UTC-время.
func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}
