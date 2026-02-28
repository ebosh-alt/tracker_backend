package meapi

type updateSettingsRequestATO struct {
	Timezone  *string `json:"timezone"`
	StepsGoal *int    `json:"stepsGoal"`
}
