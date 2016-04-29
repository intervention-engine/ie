package huddles

import (
	"time"

	"github.com/intervention-engine/fhir/models"
)

// HuddleConfig represents a configuration for how huddles should be automatically populated.  The LeaderID
// is expected to correspond to a Practitioner.  Days refers to the days of the week on which the huddle meets.
// LookAhead determines how many huddles should be scheduled into the future.  The further out, the more time
// it takes to plan them and the less certain they are (since any changes ripple out into the future). RiskConfig
// specifies how risk scores are converted to huddle frequencies.  SchedulerCronSpec indicates when the auto
// scheduler should be run (for example, nightly) and follows the cron expression format defined by
// https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format
type HuddleConfig struct {
	Name              string
	LeaderID          string
	Days              []time.Weekday
	LookAhead         int
	RiskConfig        *ScheduleByRiskConfig
	EventConfig       *ScheduleByEventConfig
	SchedulerCronSpec string
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

// ScheduleByEventConfig represents how recent events should influence huddle population
type ScheduleByEventConfig struct {
	EncounterConfigs []EncounterEventConfig
}

// EncounterEventConfig represents what types of encounters should cause patients to be scheduled, and how far back
// the algorithm should look for them
type EncounterEventConfig struct {
	LookBackDays int
	TypeCodes    []EventCode
}

// EventCode represents a coded event that should cause a patient to be scheduled.  The Name will be displayed as
// part of the reason the patient was scheduled.  If UseEndDate is set to true, then the end date, rather than the
// start date, will be used in the scheduling algorithm.
type EventCode struct {
	Name       string
	System     string
	Code       string
	UseEndDate bool
}
