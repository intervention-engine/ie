package appt

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"time"

	"github.com/intervention-engine/fhir/models"
)

func readConfigs(paths []string) []Config {
	configs := make([]Config, len(paths))
	for _, path := range paths {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Println("was not able to read file for huddle configuration at path: ", path)
			log.Println(err)
			continue
		}
		var config Config
		if err := json.Unmarshal(data, &config); err != nil {
			log.Println("was not able to parse huddle configuration correctly at path: ", path)
			log.Println(err)
			continue
		}
		// This code will only be executed if we have a correct huddle configuration.
		configs = append(configs, config)
	}
	log.Println(configs)
	return configs
}

// Config represents a configuration for how huddles should be automatically populated.
type Config struct {
	// Only used to give a descriptive name for this configuration.
	Name string
	// The CareTeamName is expected to correspond to a CareTeam.
	CareTeamName string
	// Days refers to the days of the week on which the huddle meets.
	Days []time.Weekday
	// LookAhead determines how many huddles should be scheduled into the future.
	// The further out, the more time it takes to plan them and the less certain
	// they are (since any changes ripple out into the future).
	LookAhead int
	// RiskConfig specifies how risk scores are converted to huddle frequencies.
	Risk  *ScheduleByRisk `json:"riskConfig"`
	Event *ScheduleByEvent `json:"eventConfig"`
	//  RollOverDelayInDays indicates when patients should be rolled over to the next huddle if they weren't discussed.
	// 1 means they will be rolled over to the next huddle the day after their original huddle (2 means they will be rolled
	// over 2 days after their huddle). If RollOverDelayInDays isn't specified in the config, or is less than 1, patients are
	// never rolled over to the next huddle
	RollOverDelayInDays int
	// Interval defines when the scheduler should schedule huddles according to this config.
	// Format "HH:mm:ss". Default is "00:00:00".
	Interval string
}

// IsHuddleDay returns true if the passed in date occurs on one of configured huddle weekdays.
func (c *Config) IsHuddleDay(date time.Time) bool {
	for i := range c.Days {
		if date.Weekday() == c.Days[i] {
			return true
		}
	}
	return false
}

// FindRiskScoreFrequencyByScore finds the proper risk config for a given score
func (c *Config) FindRiskScoreFrequenciesByScore(score float64) *RiskScoreFrequency {
	if c.Risk == nil {
		return nil
	}
	for i, freq := range c.Risk.Frequencies {
		if score >= freq.MinScore && score <= freq.MaxScore {
			return &c.Risk.Frequencies[i]
		}
	}
	return nil
}

func (c *Config) getTargetHuddleSize(itineraries itineraryMap) int {
	frequencyCount := make(map[int]int)
	for _, itn := range itineraries {
		if itn.Score != nil {
			if freq := c.FindRiskScoreFrequenciesByScore(*itn.Score); freq != nil {
				frequencyCount[freq.IdealFreq] = frequencyCount[freq.IdealFreq] + 1
			}
		}
	}
	var patientsPerHuddle float64
	for frequency, count := range frequencyCount {
		patientsPerHuddle += (float64(count) / float64(frequency))
	}
	return int(math.Ceil(patientsPerHuddle))
}

func (e EncounterEventConfig) EarliestLookBackDate(huddleDate time.Time) time.Time {
	// Find the low date representing the earliest time to look back to
	y, m, d := huddleDate.AddDate(0, 0, -1*e.LookBackDays).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, huddleDate.Location())
}

func (e EncounterEventConfig) LatestLookBackDate(huddleDate time.Time) time.Time {
	// Find the high date representing the latest time (exclusive) to look at
	y, m, d := huddleDate.AddDate(0, 0, 1).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, huddleDate.Location())
}

// ScheduleByRisk represents how a risk assessment should influence huddle population
type ScheduleByRisk struct {
	RiskMethod  models.Coding
	Frequencies []RiskScoreFrequency `json:"frequencyConfigs"`
}

// RiskScoreFrequency represents the relationship between risk scores and frequency of huddle discussion
type RiskScoreFrequency struct {
	MinScore  float64
	MaxScore  float64
	IdealFreq int
	MinFreq   int
	MaxFreq   int
}

// ScheduleByEvent represents how recent events should influence huddle population
type ScheduleByEvent struct {
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
