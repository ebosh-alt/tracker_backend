package shared

import "time"

// Clock абстрагирует источник текущего времени для детерминированных тестов.
type Clock interface {
	Now() time.Time
}
