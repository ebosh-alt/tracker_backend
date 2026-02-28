package steps

import "tracker/internal/domain/shared"

type Point struct {
	Date  shared.LocalDate
	Steps int
}

func (p Point) Validate() error {
	if err := p.Date.Validate(); err != nil {
		return err
	}
	if p.Steps < 0 {
		return shared.ErrInvalidInput
	}
	return nil
}

type Analytics struct {
	GoalTotal         int
	FactTotal         int
	CompletionPercent float64
	Series            []Point
}

func (a *Analytics) Validate() error {
	if a.GoalTotal < 0 {
		return shared.ErrInvalidInput
	}
	if a.FactTotal < 0 {
		return shared.ErrInvalidInput
	}
	if a.CompletionPercent < 0 {
		return shared.ErrInvalidInput
	}
	for _, point := range a.Series {
		if err := point.Validate(); err != nil {
			return err
		}
	}
	return nil
}
