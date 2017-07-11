package appt

import (
	"sort"

	"github.com/intervention-engine/ie/app"
	"github.com/intervention-engine/ie/storage"
)

type itineraryMap map[string]itinerary

type itinerary struct {
	ID                    string
	Score                 *float64
	LastHuddle            *int
	NextIdealHuddle       *int
	NearestAllowedHuddle  *int
	FurthestAllowedHuddle *int
}

func (itnMap itineraryMap) getPrioritizedPatientList(huddleCount int) []itinerary {
	itineraries := make([]itinerary, 0, len(itnMap))
	for _, itn := range itnMap {
		if itn.FurthestAllowedHuddle != nil {
			itineraries = append(itineraries, itn)
		}
	}
	p := priority{patients: itineraries, huddleIdx: huddleCount}
	sort.Sort(p)
	return p.patients
}

func (itnMap itineraryMap) populate(service storage.SchedService, config Config) error {
	if err := itnMap.populateWithRiskScores(service, config); err != nil {
		return err
	}
	if err := itnMap.populateWithHuddleInfo(service, config); err != nil {
		return err
	}
	return nil
}

func (itnMap itineraryMap) updatePatientsLastHuddle(huddle *app.Huddle, config Config, pos int) {
	for _, patient := range huddle.Patients {
		key := *patient.ID
		itn, ok := itnMap[key]
		if !ok {
			itn = itinerary{ID: key}
		}
		itn.SetLastHuddle(pos)
		itn.UpdateHuddleTargets(config)
		itnMap[key] = itn
	}
}

func (itnMap itineraryMap) populateWithRiskScores(service storage.SchedService, config Config) error {
	riskQuery, ok := buildRiskQuery(config)
	if !ok {
		return nil
	}
	results, err := service.RiskAssessmentsFilterBy(riskQuery)
	if err != nil {
		return err
	}
	for _, result := range results {
		score := result.Value
		key := *result.RiskServiceID
		itn, ok := itnMap[key]
		if !ok {
			itn = itinerary{ID: key}
		}
		itn.Score = score
		itnMap[key] = itn
	}

	return nil
}

func (itnMap itineraryMap) populateWithHuddleInfo(service storage.SchedService, config Config) error {
	// Find all of the huddles by the care team name and dates before today
	hh, err := service.FindCareTeamHuddlesBefore(config.CareTeamName, today())
	if err != nil {
		return err
	}
	// Iterate through them, setting the last huddle as appropriate
	i := -1 // Start with -1 because we will record last huddle relative to now (so -1 is one huddle ago)
	for _, result := range hh {
		for _, patient := range result.Patients {
			key := *patient.ID
			itn, ok := itnMap[key]
			if !ok {
				itn = itinerary{ID: key}
			}
			if itn.LastHuddle == nil || i > *itn.LastHuddle {
				itn.SetLastHuddle(i)
				itn.UpdateHuddleTargets(config)
			}
			itnMap[key] = itn
		}
		i--
	}
	// Some of the patients won't have a last huddle, but we should still set their nearest/furthest huddle if possible
	for _, itn := range itnMap {
		if itn.Score != nil && itn.LastHuddle == nil {
			itn.UpdateHuddleTargets(config)
		}
	}
	return nil
}

// This is wrapped in a function in order to protect the value of LastHuddle changing.
// Since it's a pointer to an integer, that integer could be changed, which wouldn't be
// the intended behavior of setting LastHuddle. This way, the integer value is copied into
// this function's parameter, making it safe to take its address and trust that it won't be
// modified.
func (it *itinerary) SetLastHuddle(lastHuddle int) {
	it.LastHuddle = &lastHuddle
}

func (it *itinerary) UpdateHuddleTargets(config Config) {
	if it.Score == nil {
		return
	}
	cfg := config.FindRiskScoreFrequenciesByScore(*it.Score)
	if cfg == nil {
		return
	}
	if it.LastHuddle != nil {
		ideal := *it.LastHuddle + cfg.IdealFreq
		it.NextIdealHuddle = &ideal
		nearest := *it.LastHuddle + cfg.MinFreq
		it.NearestAllowedHuddle = &nearest
		furthest := *it.LastHuddle + cfg.MaxFreq
		it.FurthestAllowedHuddle = &furthest
		return
	}
	// Patient has never had a huddle.  The next huddle is really ideal (if possible).
	nearest := 0
	it.NextIdealHuddle = &nearest
	it.NearestAllowedHuddle = &nearest
	furthest := cfg.MaxFreq - 1
	it.FurthestAllowedHuddle = &furthest
}

// TODO: risk assessments are stored outside FHIR now.
// Find all of the patients in the scoring ranges used to schedule huddles
func buildRiskQuery(config Config) (storage.RiskFilterQuery, bool) {
	if config.Risk == nil || len(config.Risk.Frequencies) == 0 {
		return nil, false
	}
	riskQuery := storage.RiskFilterQuery{}

	if len(config.Risk.Frequencies) == 1 {
		riskQuery["value"] = map[string]float64{
			">": config.Risk.Frequencies[0].MinScore,
			"<": config.Risk.Frequencies[0].MaxScore,
		}
	} else {
		ranges := make([]map[string]float64, len(config.Risk.Frequencies))
		for i, frqCfg := range config.Risk.Frequencies {
			ranges[i] = map[string]float64{
				">": frqCfg.MinScore,
				"<": frqCfg.MaxScore,
			}
		}
		riskQuery["values"] = ranges
	}
	return riskQuery, true
}
