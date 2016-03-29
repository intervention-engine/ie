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
				MinTimeBetweenHuddles: 5 /*days*/ * 24 * time.Hour,
				MaxTimeBetweenHuddles: 7 /*days*/ * 24 * time.Hour,
			},
			{
				MinScore:              4,
				MaxScore:              5,
				MinTimeBetweenHuddles: 12 /*days*/ * 24 * time.Hour,
				MaxTimeBetweenHuddles: 14 /*days*/ * 24 * time.Hour,
			},
			{
				MinScore:              1,
				MaxScore:              3,
				MinTimeBetweenHuddles: 25 /*days*/ * 24 * time.Hour,
				MaxTimeBetweenHuddles: 28 /*days*/ * 24 * time.Hour,
			},
		},
	},
}
