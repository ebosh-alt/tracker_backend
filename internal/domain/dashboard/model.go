package dashboard

import (
	"time"

	"tracker/internal/domain/shared"
)

// Model агрегирует ключевые показатели дня для главного экрана.
type Model struct {
	Date        shared.LocalDate
	Streak      int
	Steps       StepsWidget
	Medications MedicationsWidget
	Workouts    WorkoutsWidget
	Alerts      []Alert
}

// StepsWidget данные блока шагов на дашборде.
type StepsWidget struct {
	Today     int
	Goal      int
	Progress  float64
	Remaining int
	History7D []StepsPoint
}

// StepsPoint точка временного ряда шагов для UI.
type StepsPoint struct {
	Date  string
	Steps int
}

// MedicationsWidget данные приверженности по БАД.
type MedicationsWidget struct {
	Taken   int
	Total   int
	Missed  int
	Pending int
}

// WorkoutsWidget данные по тренировкам на выбранную дату.
type WorkoutsWidget struct {
	TodayCount int
}

// AlertType тип предупреждения для UI.
type AlertType string

const (
	AlertMedicationMissed AlertType = "medication_missed"
)

// Alert единичный сигнал для карточки "внимание".
type Alert struct {
	Type         AlertType
	MedicationID *int64
	ScheduledAt  *time.Time
}
