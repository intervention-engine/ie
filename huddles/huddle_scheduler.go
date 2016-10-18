package huddles

import (
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"github.com/intervention-engine/fhir/server"
	"github.com/labstack/gommon/log"
)

// HuddleScheduler schedules huddles based on the passed in config.
type HuddleScheduler struct {
	Config       *HuddleConfig
	Huddles      []*Huddle
	patientInfos patientInfoMap
}

// NewHuddleScheduler initializes a new huddle scheduler based on the passed in config.
func NewHuddleScheduler(config *HuddleConfig) *HuddleScheduler {
	return &HuddleScheduler{
		Config:       config,
		patientInfos: make(patientInfoMap),
	}
}

type patientInfoMap map[string]*patientInfo

func (p patientInfoMap) SafeGet(key string) *patientInfo {
	pInfo := p[key]
	if pInfo == nil {
		pInfo = &patientInfo{ID: key}
		p[key] = pInfo
	}
	return pInfo
}

type patientInfo struct {
	ID                    string
	Score                 *float64
	LastHuddle            *int
	NextIdealHuddle       *int
	NearestAllowedHuddle  *int
	FurthestAllowedHuddle *int
}

func (p *patientInfo) FindFrequencyConfig(config *HuddleConfig) *RiskScoreFrequencyConfig {
	if p.Score != nil {
		return config.FindRiskScoreFrequencyConfigByScore(*p.Score)
	}
	return nil
}

// SetLastHuddle should always be used to set the last huddle, because it also sets other important properties
func (p *patientInfo) SetLastHuddle(lastHuddle int, config *HuddleConfig) {
	p.LastHuddle = &lastHuddle
	p.UpdateHuddleTargets(config)
}

func (p *patientInfo) UpdateHuddleTargets(config *HuddleConfig) {
	cfg := p.FindFrequencyConfig(config)
	if cfg != nil {
		if p.LastHuddle != nil {
			ideal := *p.LastHuddle + cfg.IdealFrequency
			p.NextIdealHuddle = &ideal
			nearest := *p.LastHuddle + cfg.MinFrequency
			p.NearestAllowedHuddle = &nearest
			furthest := *p.LastHuddle + cfg.MaxFrequency
			p.FurthestAllowedHuddle = &furthest
		} else {
			// Patient has never had a huddle.  The next huddle is really ideal (if possible).
			nearest := 0
			p.NextIdealHuddle = &nearest
			p.NearestAllowedHuddle = &nearest
			furthest := cfg.MaxFrequency - 1
			p.FurthestAllowedHuddle = &furthest
		}
	}
}

// ScheduleHuddles schedules huddles based on the passed in config.  It will schedule out the number
// of huddles as specified in the config.LookAhead.
func (hs *HuddleScheduler) ScheduleHuddles() ([]*Huddle, error) {
	// First populate the structures we need to do the scheduling
	if err := hs.populatePatientInfosWithRiskScores(); err != nil {
		log.Error(err)
		return nil, err
	}

	err := hs.populatePatientInfosWithHuddleInfo()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Then create the populated huddles
	err = hs.createHuddles()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Store the huddles in the database
	var lastErr error
	for i := range hs.Huddles {
		if _, err := server.Database.C("groups").UpsertId(hs.Huddles[i].Id, hs.Huddles[i]); err != nil {
			lastErr = err
			log.Warn("Error storing huddle: %t", err)
		}
	}

	hs.printInfo()

	return hs.Huddles, lastErr
}

func (hs *HuddleScheduler) populatePatientInfosWithRiskScores() error {
	if hs.Config.RiskConfig == nil || len(hs.Config.RiskConfig.FrequencyConfigs) == 0 {
		return nil
	}

	// Find all of the patients in the scoring ranges used to schedule huddles
	// NOTE: We don't use IE search framework because prediction.probabilityDecimal is not a search parameter.
	// That said, we could consider creating a custom search param in the future if we really wanted...
	riskQuery := bson.M{
		"method.coding": bson.M{
			"$elemMatch": bson.M{
				"system": hs.Config.RiskConfig.RiskMethod.System,
				"code":   hs.Config.RiskConfig.RiskMethod.Code,
			},
		},
		"meta.tag": bson.M{
			"$elemMatch": bson.M{
				"system": "http://interventionengine.org/tags/",
				"code":   "MOST_RECENT",
			},
		},
		"subject.external": false,
	}

	if len(hs.Config.RiskConfig.FrequencyConfigs) == 1 {
		frqCfg := hs.Config.RiskConfig.FrequencyConfigs[0]
		riskQuery["prediction.probabilityDecimal"] = bson.M{
			"$gte": frqCfg.MinScore,
			"$lte": frqCfg.MaxScore,
		}
	} else {
		ranges := make([]bson.M, len(hs.Config.RiskConfig.FrequencyConfigs))
		for i := range hs.Config.RiskConfig.FrequencyConfigs {
			frqCfg := hs.Config.RiskConfig.FrequencyConfigs[i]
			ranges[i] = bson.M{
				"prediction.probabilityDecimal": bson.M{
					"$gte": frqCfg.MinScore,
					"$lte": frqCfg.MaxScore,
				},
			}
		}
		riskQuery["$or"] = ranges
	}

	selector := bson.M{
		"_id": 0,
		"subject.referenceid":           1,
		"prediction.probabilityDecimal": 1,
	}
	iter := server.Database.C("riskassessments").Find(riskQuery).Select(selector).Iter()
	result := models.RiskAssessment{}
	for iter.Next(&result) {
		info := hs.patientInfos.SafeGet(result.Subject.ReferencedID)
		info.Score = result.Prediction[0].ProbabilityDecimal
	}

	return iter.Close()
}

func (hs *HuddleScheduler) populatePatientInfosWithHuddleInfo() error {
	// Find all of the huddles by the leader id and dates before today
	searcher := search.NewMongoSearcher(server.Database)
	queryStr := fmt.Sprintf("leader=Practitioner/%s&activedatetime=lt%s", hs.Config.LeaderID, today().Format("2006-01-02T-07:00"))
	mgoQuery := searcher.CreateQueryWithoutOptions(search.Query{Resource: "Group", Query: queryStr})
	selector := bson.M{
		"_id": 0,
		"member.entity.referenceid": 1,
	}
	iter := mgoQuery.Sort("-extension.activeDateTime.time").Select(selector).Iter()

	// Iterate through them, setting the last huddle as appropriate
	i := -1 // Start with -1 because we will record last huddle relative to now (so -1 is one huddle ago)
	result := models.Group{}
	for iter.Next(&result) {
		for _, member := range result.Member {
			pInfo := hs.patientInfos.SafeGet(member.Entity.ReferencedID)
			if pInfo.LastHuddle == nil || i > *pInfo.LastHuddle {
				pInfo.SetLastHuddle(i, hs.Config)
			}
		}
		i--
	}

	if err := iter.Close(); err != nil {
		return err
	}

	// Some of the patients won't have a last huddle, but we should still set their nearest/furthest huddle if possible
	for _, pInfo := range hs.patientInfos {
		if pInfo.Score != nil && pInfo.LastHuddle == nil {
			pInfo.UpdateHuddleTargets(hs.Config)
		}
	}

	return nil
}

func (hs *HuddleScheduler) createHuddles() error {
	targetHuddleSize := hs.getTargetHuddleSize()

	// TODO: Clear future huddles scheduled on the wrong days (unless they have manual additions)?
	// Step through one day at a time, starting today, until we have created the requested number of huddles
	hs.Huddles = make([]*Huddle, 0, hs.Config.LookAhead)
	checkRollOversAndEvents := true
	for t := today(); len(hs.Huddles) < hs.Config.LookAhead; t = t.AddDate(0, 0, 1) {
		if !hs.Config.IsHuddleDay(t) {
			continue
		}

		huddle, err := hs.findExistingHuddle(t)
		if err != nil {
			return err
		}

		huddleIdx := len(hs.Huddles)

		// If this is today's huddle and any patients are marked reviewed already, then do NOT reschedule this huddle!
		if huddle != nil && huddle.ActiveDateTime() != nil && huddle.ActiveDateTime().Time == today() {
			huddleInProgress := false
			for _, member := range huddle.HuddleMembers() {
				if member.Reviewed() != nil {
					huddleInProgress = true
					break
				}
			}
			if huddleInProgress {
				// Need to update the patientInfo last huddles and add the huddle to our slice of huddles
				for _, member := range huddle.HuddleMembers() {
					hs.patientInfos.SafeGet(member.ID()).SetLastHuddle(huddleIdx, hs.Config)
				}
				hs.Huddles = append(hs.Huddles, huddle)
				// Go to the next huddle
				continue
			}
		}

		var originalMembers []HuddleMember
		if huddle == nil {
			// Create a new huddle
			huddle = NewHuddle(hs.Config.Name, hs.Config.LeaderID, t)
		} else {
			// Remember the original members
			originalMembers = huddle.HuddleMembers()
			// Clear the original members so we start from clean slate
			huddle.Member = nil
		}

		// Add back the manually added and rolled over patients
		for _, member := range originalMembers {
			if member.ReasonIsManuallyAdded() || member.ReasonIsRollOver() {
				huddle.addHuddleMember(member.ID(), member.Reason())
			}
		}

		if checkRollOversAndEvents {
			// Add new rollovers if applicable
			if hs.Config.RollOverDelayInDays > 0 {
				hs.addMembersBasedOnRollOvers(huddle)
			}

			// Add members to the huddle who had a recent encounter that triggers huddle discussion
			if hs.Config.EventConfig != nil && len(hs.Config.EventConfig.EncounterConfigs) > 0 {
				hs.addMembersBasedOnRecentEncounters(huddle, huddleIdx)
			}

			checkRollOversAndEvents = false
		}

		// Finally fill out the rest with members based on their risk scores (which determine huddle frequency)
		if hs.Config.RiskConfig != nil && len(hs.Config.RiskConfig.FrequencyConfigs) > 0 {
			hs.addMembersBasedOnRiskScores(huddle, huddleIdx, targetHuddleSize)
		}

		// Now go through all of the assigned huddle members and update that infos
		for _, member := range huddle.HuddleMembers() {
			pInfo := hs.patientInfos.SafeGet(member.ID())
			pInfo.SetLastHuddle(huddleIdx, hs.Config)
		}

		hs.Huddles = append(hs.Huddles, huddle)
	}
	return nil
}

func (hs *HuddleScheduler) getTargetHuddleSize() int {
	frequencyCountMap := make(map[int]int)
	for _, patientInfo := range hs.patientInfos {
		frqCfg := patientInfo.FindFrequencyConfig(hs.Config)
		if frqCfg != nil {
			frequencyCountMap[frqCfg.IdealFrequency] = frequencyCountMap[frqCfg.IdealFrequency] + 1
		}
	}
	var patientsPerHuddle float64
	for frequency, count := range frequencyCountMap {
		patientsPerHuddle += (float64(count) / float64(frequency))
	}
	return int(math.Ceil(patientsPerHuddle))
}

func (hs *HuddleScheduler) findExistingHuddle(date time.Time) (*Huddle, error) {
	searcher := search.NewMongoSearcher(server.Database)
	queryStr := fmt.Sprintf("leader=Practitioner/%s&activedatetime=%s&_sort=activedatetime&_count=1", hs.Config.LeaderID, date.Format("2006-01-02T-07:00"))
	var huddles []*models.Group
	if err := searcher.CreateQuery(search.Query{Resource: "Group", Query: queryStr}).All(&huddles); err != nil {
		return nil, err
	} else if len(huddles) > 0 {
		huddle := Huddle(*huddles[0])
		return &huddle, nil
	}
	return nil, nil
}

func (hs *HuddleScheduler) addMembersBasedOnRiskScores(huddle *Huddle, huddleIdx, targetSize int) {
	for _, p := range hs.getPrioritizedPatientList(huddleIdx) {
		// If we hit (or exceeded) our target, only stop if the patient *can* be put into a further huddle
		if len(huddle.Member) >= targetSize && (p.FurthestAllowedHuddle == nil || huddleIdx < *p.FurthestAllowedHuddle) {
			break
		}
		// If this huddle is before the nearest allowed, then we should not add this patient
		if p.NearestAllowedHuddle != nil && huddleIdx < *p.NearestAllowedHuddle {
			continue
		}
		// Otherwise, add the patient to the huddle
		huddle.AddHuddleMemberDueToRiskScore(p.ID)
	}
}

func (hs *HuddleScheduler) getPrioritizedPatientList(huddleIdx int) []*patientInfo {
	patients := make([]*patientInfo, 0, len(hs.patientInfos))
	for _, pInfo := range hs.patientInfos {
		if pInfo.FurthestAllowedHuddle != nil {
			patients = append(patients, pInfo)
		}
	}
	byHP := byHuddlePriority{patients: patients, huddleIdx: huddleIdx}
	sort.Sort(byHP)
	return byHP.patients
}

// Support sorting by huddle priority by using the following sort rules:
// - First check for dueness.  Patients who are due or overdue go first (more overdue before less overdue).
// - If neither is overdue, or they are overdue the same, sort from earliest ideal huddle to latest ideal huddle.
// - If they have the same ideal huddle, sub-sort by risk score (highest goes first).
// - If they have same risk score too, sub-sort by furthest allowed huddle (implies urgency).
// - If they have the same furthest allowed huddle too, sort by the last huddle (oldest first).
// - If they have same last huddle too, just sort by the patients id.
type byHuddlePriority struct {
	patients  []*patientInfo
	huddleIdx int
}

func (hp byHuddlePriority) Len() int {
	return len(hp.patients)
}
func (hp byHuddlePriority) Swap(i, j int) {
	hp.patients[i], hp.patients[j] = hp.patients[j], hp.patients[i]
}
func (hp byHuddlePriority) Less(i, j int) bool {
	// First look at "dueness" by seeing if the furthest allowed huddle is this one or before.
	// If either or both are due, and have different furthest allowed huddles, the more overdue one goes first.
	if hp.patients[i].FurthestAllowedHuddle != hp.patients[j].FurthestAllowedHuddle {
		iDue := hp.patients[i].FurthestAllowedHuddle != nil && *hp.patients[i].FurthestAllowedHuddle <= hp.huddleIdx
		jDue := hp.patients[j].FurthestAllowedHuddle != nil && *hp.patients[j].FurthestAllowedHuddle <= hp.huddleIdx
		if iDue && jDue && *hp.patients[i].FurthestAllowedHuddle != *hp.patients[j].FurthestAllowedHuddle {
			return *hp.patients[i].FurthestAllowedHuddle < *hp.patients[j].FurthestAllowedHuddle
		} else if iDue && !jDue {
			return true
		} else if !iDue && jDue {
			return false
		}
	}

	// If neither is due, or they are overdue the same, compare by the next ideal huddle.
	// If one is nil (which shouldn't happen) assume they should go last.
	if hp.patients[i].NextIdealHuddle != hp.patients[j].NextIdealHuddle {
		if hp.patients[i].NextIdealHuddle == nil {
			return false
		} else if hp.patients[j].NextIdealHuddle == nil {
			return true
		} else if *hp.patients[i].NextIdealHuddle != *hp.patients[j].NextIdealHuddle {
			return *hp.patients[i].NextIdealHuddle < *hp.patients[j].NextIdealHuddle
		}
	}

	// If they have the same next ideal huddle, compare by the score (higher goes first).
	// If one is nil assume they should go last.
	if hp.patients[i].Score != hp.patients[j].Score {
		if hp.patients[i].Score == nil {
			return false
		} else if hp.patients[j].Score == nil {
			return true
		} else if *hp.patients[i].Score != *hp.patients[j].Score {
			return *hp.patients[i].Score > *hp.patients[j].Score
		}
	}

	// If they have the same risk score too, compare by the furthest allowed huddle.
	// If one is nil (which shouldn't happen) assume they should go last.
	if hp.patients[i].FurthestAllowedHuddle != hp.patients[j].FurthestAllowedHuddle {
		if hp.patients[i].FurthestAllowedHuddle == nil {
			return false
		} else if hp.patients[j].FurthestAllowedHuddle == nil {
			return true
		} else if *hp.patients[i].FurthestAllowedHuddle != *hp.patients[j].FurthestAllowedHuddle {
			return *hp.patients[i].FurthestAllowedHuddle < *hp.patients[j].FurthestAllowedHuddle
		}
	}

	// If they have the same furthest allowed huddle, try comparing by the last huddle.
	// If one is nil, they've never had a huddle, so they should be prefered first over someone who has.
	if hp.patients[i].LastHuddle != hp.patients[j].LastHuddle {
		if hp.patients[i].LastHuddle == nil {
			return true
		} else if hp.patients[j].LastHuddle == nil {
			return false
		} else if *hp.patients[i].LastHuddle != *hp.patients[j].LastHuddle {
			return *hp.patients[i].LastHuddle < *hp.patients[j].LastHuddle
		}
	}

	// OK, they're equal for all intents and purposes, but compare by ID so we have consistent order
	return hp.patients[i].ID < hp.patients[j].ID
}

func (hs *HuddleScheduler) addMembersBasedOnRecentEncounters(huddle *Huddle, huddleIdx int) error {
	if hs.Config.EventConfig == nil {
		return nil
	}
	date := huddle.ActiveDateTime().Time
	// Loop through the event configs, looking for patients with matching encounters
	for _, eventConfig := range hs.Config.EventConfig.EncounterConfigs {
		// Find the low date representing the earliest time to look back to
		y, m, d := date.AddDate(0, 0, -1*eventConfig.LookBackDays).Date()
		lowInclDate := time.Date(y, m, d, 0, 0, 0, 0, date.Location())

		// Don't bother looking for events in the future!
		if lowInclDate.After(time.Now()) {
			continue
		}

		// Find the high date representing the latest time (exclusive) to look at
		y, m, d = date.AddDate(0, 0, 1).Date()
		highExclDate := time.Date(y, m, d, 0, 0, 0, 0, date.Location())

		// Build up the query to get all possible encounters that might trigger a huddle
		fmt := "2006-01-02T15:04:05.000-07:00"
		queryStr := "date=ge" + lowInclDate.Format(fmt) + "&date=lt" + highExclDate.Format(fmt) + "&status=arrived,in-progress,onleave,finished"
		if len(eventConfig.TypeCodes) > 0 {
			codeVals := make([]string, len(eventConfig.TypeCodes))
			for i, code := range eventConfig.TypeCodes {
				codeVals[i] = code.System + "|" + code.Code
			}
			queryStr += "&type=" + strings.Join(codeVals, ",")
		}

		searcher := search.NewMongoSearcher(server.Database)
		encQuery := searcher.CreateQueryObject(search.Query{Resource: "Encounter", Query: queryStr})

		// FOR NOW: We essentially copy/paste the encounter code we used in the previous version of the scheduler,
		// but we should revisit at some point since we may be able to streamline this WITHOUT a pipeline.

		// This pipeline starts with the encounter date/code query, sorts them by date, left-joins the huddles and then
		// returns only the info we care about.
		pipeline := []bson.M{
			{"$match": encQuery},
			{"$sort": bson.M{"period.start": -1}},
			{"$lookup": bson.M{
				"from":         "groups",
				"localField":   "patient.referenceid",
				"foreignField": "member.entity.referenceid",
				"as":           "_groups",
			}},
			{"$project": bson.M{
				"_id":       0,
				"patientID": "$patient.referenceid",
				"type":      1,
				"period":    1,
				"huddles":   "$_groups",
			}},
		}

		var results []struct {
			PatientID string                   `bson:"patientID"`
			Type      []models.CodeableConcept `bson:"type"`
			Period    *models.Period           `bson:"period"`
			Huddles   []models.Group           `bson:"huddles"`
		}
		if err := server.Database.C("encounters").Pipe(pipeline).All(&results); err != nil {
			return err
		}

		// Go through the encounters finding the matches and storing in the patient map.  Note that FHIR search only
		// allows you to search on dates representing a time that happened at some point in the encounter -- so we must
		// post-process to see if the date is a real match.
		for _, result := range results {
			if huddle.FindHuddleMember(result.PatientID) != nil {
				// Patient is already scheduled, so skip
				continue
			}
			for _, code := range eventConfig.TypeCodes {
				if codeMatches(result.Type, &code) {
					if d, matches := dateMatches(result.Period, &code, lowInclDate, highExclDate); matches {
						// Collect the PAST huddles, already in the database
						var huddles []*Huddle
						for i := range result.Huddles {
							h := Huddle(result.Huddles[i])
							if h.ActiveDateTime() != nil && h.ActiveDateTime().Time.Before(huddle.ActiveDateTime().Time) {
								huddles = append(huddles, &h)
							}
						}
						// If the patient has been discussed in a huddle since the date, then don't schedule again
						alreadyDiscussed := isPatientScheduledForSpecificEncounterReason(huddles, result.PatientID, d)
						if !alreadyDiscussed {
							// Then check the huddles we've already scheduled in this session
							alreadyDiscussed = isPatientScheduledForSpecificEncounterReason(hs.Huddles, result.PatientID, d)
						}

						if !alreadyDiscussed {
							huddle.AddHuddleMemberDueToRecentEvent(result.PatientID, code)
							break
						}
					}
				}
			}
		}
	}
	return nil
}

func isPatientScheduledForSpecificEncounterReason(huddles []*Huddle, patientID string, encounterDate time.Time) bool {
	// Go through the huddles backwards since it's more likely a discussed date is recent (although we can't guarantee the huddles are sorted)
	for i := len(huddles) - 1; i >= 0; i-- {
		h := huddles[i]
		if h.ActiveDateTime() != nil && !h.ActiveDateTime().Time.Equal(encounterDate) && h.ActiveDateTime().Time.After(encounterDate) {
			m := h.FindHuddleMember(patientID)
			// Only consider it already discussed if the patient was discussed for this same reason
			if m != nil && m.ReasonIsRecentEncounter() {
				return true
			}
		}
	}
	return false
}

func (hs *HuddleScheduler) addMembersBasedOnRollOvers(huddle *Huddle) {
	if hs.Config.RollOverDelayInDays <= 0 {
		return
	}

	// Find the patients that need to roll over (i.e., the ones not reviewed in the huddle x days ago)
	expiredHuddleDay := today().AddDate(0, 0, -1*hs.Config.RollOverDelayInDays)
	expiredHuddle, err := hs.findExistingHuddle(expiredHuddleDay)
	if err != nil {
		fmt.Printf("Error searching on previous huddle (%s) to detect rollover patients\n", expiredHuddleDay.Format("Jan 2"))
	} else if expiredHuddle != nil {
		// Check for unreviewed patients
		eh := Huddle(*expiredHuddle)
		for _, member := range eh.HuddleMembers() {
			if member.Reviewed() == nil {
				huddle.AddHuddleMemberDueToRollOver(member.ID(), expiredHuddleDay, member.Reason())
			}
		}
	}
}

func (hs *HuddleScheduler) printInfo() {
	fmt.Printf("Scheduled %d huddles with name %s\n", len(hs.Huddles), hs.Config.Name)
	for i := range hs.Huddles {
		fmt.Printf("\t%s: %d patients\n", getStringDate(hs.Huddles[i]), len(hs.Huddles[i].Member))
	}
}

var _nowValueForTestingOnly *time.Time

func now() time.Time {
	if _nowValueForTestingOnly != nil {
		return *_nowValueForTestingOnly
	}
	return time.Now()
}
func today() time.Time {
	now := now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func codeMatches(encType []models.CodeableConcept, code *EventCode) bool {
	return models.CodeableConcepts(encType).AnyMatchesCode(code.System, code.Code)
}

func dateMatches(encPeriod *models.Period, code *EventCode, lowIncl, highExcl time.Time) (time.Time, bool) {
	if encPeriod == nil {
		return time.Time{}, false
	}
	d := encPeriod.Start
	if code.UseEndDate {
		d = encPeriod.End
	}
	if d != nil && !d.Time.Before(lowIncl) && d.Time.Before(highExcl) {
		return d.Time, true
	}
	return time.Time{}, false
}

func getStringDate(huddle *Huddle) string {
	dt := huddle.ActiveDateTime()
	if dt != nil {
		return dt.Time.Format("01/02/2006")
	}
	return ""
}
