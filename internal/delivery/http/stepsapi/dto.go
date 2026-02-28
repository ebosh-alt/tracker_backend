package stepsapi

import (
	"time"

	appsteps "tracker/internal/application/steps"
	domainSteps "tracker/internal/domain/steps"
)

type dailyEntryDTO struct {
	Date      string `json:"date"`
	Count     int    `json:"count"`
	Source    string `json:"source"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type listResponseDTO struct {
	Entries []dailyEntryDTO `json:"entries"`
}

type putResponseDTO struct {
	Entry dailyEntryDTO `json:"entry"`
}

type addResponseDTO struct {
	Entry dailyEntryDTO `json:"entry"`
}

type analyticsPointDTO struct {
	Date  string `json:"date"`
	Steps int    `json:"steps"`
}

type analyticsPeriodDTO struct {
	From              string              `json:"from"`
	To                string              `json:"to"`
	GoalTotal         int                 `json:"goalTotal"`
	FactTotal         int                 `json:"factTotal"`
	CompletionPercent float64             `json:"completionPercent"`
	Series            []analyticsPointDTO `json:"series"`
}

type analyticsResponseDTO struct {
	GoalPerDay int                `json:"goalPerDay"`
	Week       analyticsPeriodDTO `json:"week"`
	Month      analyticsPeriodDTO `json:"month"`
}

func mapDailyEntry(entry domainSteps.DailyEntry) dailyEntryDTO {
	return dailyEntryDTO{
		Date:      entry.Date.String(),
		Count:     entry.Count,
		Source:    string(entry.Source),
		CreatedAt: entry.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: entry.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func mapListResponse(out appsteps.ListOutput) listResponseDTO {
	entries := make([]dailyEntryDTO, 0, len(out.Entries))
	for _, e := range out.Entries {
		entries = append(entries, mapDailyEntry(e))
	}
	return listResponseDTO{Entries: entries}
}

func mapPutResponse(out appsteps.PutOutput) putResponseDTO {
	return putResponseDTO{Entry: mapDailyEntry(out.Entry)}
}

func mapAddResponse(out appsteps.AddOutput) addResponseDTO {
	return addResponseDTO{Entry: mapDailyEntry(out.Entry)}
}

func mapAnalyticsPeriod(period appsteps.Period) analyticsPeriodDTO {
	series := make([]analyticsPointDTO, 0, len(period.Analytics.Series))
	for _, p := range period.Analytics.Series {
		series = append(series, analyticsPointDTO{
			Date:  p.Date.String(),
			Steps: p.Steps,
		})
	}
	return analyticsPeriodDTO{
		From:              period.From.String(),
		To:                period.To.String(),
		GoalTotal:         period.Analytics.GoalTotal,
		FactTotal:         period.Analytics.FactTotal,
		CompletionPercent: period.Analytics.CompletionPercent,
		Series:            series,
	}
}

func mapAnalyticsResponse(out appsteps.AnalyticsOutput) analyticsResponseDTO {
	return analyticsResponseDTO{
		GoalPerDay: out.GoalPerDay,
		Week:       mapAnalyticsPeriod(out.Week),
		Month:      mapAnalyticsPeriod(out.Month),
	}
}
