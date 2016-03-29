package huddles

import (
	"time"

	"github.com/intervention-engine/fhir/models"
)

// HuddleConfig represents a configuration for how huddles should be automatically populated.  The LeaderID
// is expected to correspond to a Practitioner.  Days refers to the days of the week on which the huddle meets.
// LookAhead determines how many huddles should be scheduled into the future.  The further out, the more time
// it takes to plan them and the less certain they are (since any changes ripple out into the future).
type HuddleConfig struct {
	Name       string
	LeaderID   string
	Days       []time.Weekday
	LookAhead  int
	RiskConfig *ScheduleByRiskConfig
}

// IsHuddleDay returns true if the passed in date occurs on one of configured huddle weekdays.
func (hc *HuddleConfig) IsHuddleDay(date time.Time) bool {
	for i := range hc.Days {
		if date.Weekday() == hc.Days[i] {
			return true
		}
	}
	return false
}

// ScheduleByRiskConfig represents how a risk assessment should influence huddle population
type ScheduleByRiskConfig struct {
	RiskMethod       models.Coding
	FrequencyConfigs []RiskScoreFrequencyConfig
}

// RiskScoreFrequencyConfig represents the relationship between risk scores and frequency of huddle discussion
type RiskScoreFrequencyConfig struct {
	MinScore              float64
	MaxScore              float64
	MinDaysBetweenHuddles int
	MaxDaysBetweenHuddles int
}
