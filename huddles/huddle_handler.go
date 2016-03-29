package huddles

import (
	"net/http"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/labstack/echo"
)

func AutoScheduleHuddlesHandler(c *echo.Context) error {
	huddles, err := AutoScheduleHuddles(DefaultNegativeOutcomeHuddleConfig)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, huddles)
}

var DefaultNegativeOutcomeHuddleConfig = &HuddleConfig{
	Name:      "Weekly Huddle",
	LeaderID:  "1",
	Days:      []time.Weekday{time.Monday},
	LookAhead: 4,
	RiskConfig: &ScheduleByRiskConfig{
		RiskMethod: models.Coding{System: "http://interventionengine.org/risk-assessments", Code: "Simple"},
		FrequencyConfigs: []RiskScoreFrequencyConfig{
			{
				MinScore:              6,
				MaxScore:              10,
				MinDaysBetweenHuddles: 5,
				MaxDaysBetweenHuddles: 7,
			},
			{
				MinScore:              4,
				MaxScore:              5,
				MinDaysBetweenHuddles: 12,
				MaxDaysBetweenHuddles: 14,
			},
			{
				MinScore:              1,
				MaxScore:              3,
				MinDaysBetweenHuddles: 25,
				MaxDaysBetweenHuddles: 28,
			},
		},
	},
}
