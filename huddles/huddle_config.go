package huddles

import (
	"time"

	"github.com/intervention-engine/fhir/models"
)

// HuddleConfig represents a configuration for how huddles should be automatically populated
type HuddleConfig struct {
	Name       string
	LeaderID   string
	Days       []time.Weekday
	RiskConfig *ScheduleByRiskConfig
}

// ScheduleByRiskConfig represents how a risk assessment should influence huddle population
type ScheduleByRiskConfig struct {
	RiskCode         string
	FrequencyConfigs []RiskScoreFrequencyConfig
	EncounterCodes   []models.Coding
}

// RiskScoreFrequencyConfig represents the relationship between risk scores and frequency of huddle discussion
type RiskScoreFrequencyConfig struct {
	MinScore              int
	MaxScore              int
	MinTimeBetweenHuddles time.Duration
	MaxTimeBetweenHuddles time.Duration
}
