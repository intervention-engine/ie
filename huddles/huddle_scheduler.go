package huddles

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"github.com/intervention-engine/fhir/server"
)

// ScheduleHuddles schedules future huddles based on the passed in config.  It will schedule out the number
// of future huddles as specified in the config.LookAhead.
func ScheduleHuddles(config *HuddleConfig) ([]*models.Group, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.Local)

	// Step through one day at a time, starting today, until we have scheduled the requested number of huddles
	huddles := make([]*models.Group, 0, config.LookAhead)
	for t := start; len(huddles) < config.LookAhead; t = t.AddDate(0, 0, 1) {
		if !config.IsHuddleDay(t) {
			continue
		}

		// Create the huddle
		huddle, err := CreatePopulatedHuddle(t, config)
		if err != nil {
			return nil, err
		}

		// Store the huddle
		if _, err := server.Database.C("groups").UpsertId(huddle.Id, huddle); err != nil {
			return huddles, err
		}
		huddles = append(huddles, huddle)
	}

	printInfo(huddles, config.Name)

	return huddles, nil
}

// CreatePopulatedHuddle returns a Group resource representing the patients that should be automatically considered
// for a huddle for the specific date.  Currently it is based on three criteria:
// - Risk scores (which determine frequency)
// - Recent clinical events (such as ED visit)
// - "Leftovers" from previous huddle
func CreatePopulatedHuddle(date time.Time, config *HuddleConfig) (*models.Group, error) {
	group, err := findExistingHuddle(date, config)
	if err != nil {
		return nil, err
	}

	if group == nil {
		tru := true
		group = &models.Group{
			DomainResource: models.DomainResource{
				Resource: models.Resource{
					Id:           bson.NewObjectId().Hex(),
					ResourceType: "Group",
					Meta: &models.Meta{
						Profile: []string{"http://interventionengine.org/fhir/profile/huddle"},
					},
				},
				Extension: []models.Extension{
					{
						Url:           "http://interventionengine.org/fhir/extension/group/activeDateTime",
						ValueDateTime: &models.FHIRDateTime{Time: date, Precision: models.Precision(models.Date)},
					},
					{
						Url: "http://interventionengine.org/fhir/extension/group/leader",
						ValueReference: &models.Reference{
							Reference:    "Practitioner/" + config.LeaderID,
							ReferencedID: config.LeaderID,
							Type:         "Practitioner",
							External:     new(bool),
						},
					},
				},
			},
			Type:   "person",
			Actual: &tru,
			Code: &models.CodeableConcept{
				Coding: []models.Coding{
					{System: "http://interventionengine.org/fhir/cs/huddle", Code: "HUDDLE"},
				},
				Text: "Huddle",
			},
			Name: config.Name,
		}
	}

	// First find the manually added patients (in case this was existing) and remember them
	var manuallyAddedPatientIDs []string
	for i := range group.Member {
		for j := range group.Member[i].Extension {
			ext := group.Member[i].Extension[j]
			if ext.Url == "http://interventionengine.org/fhir/extension/group/member/reason" &&
				ext.ValueCodeableConcept != nil &&
				len(ext.ValueCodeableConcept.Coding) > 0 &&
				ext.ValueCodeableConcept.Coding[0].Code == ManualAdditionReason.Coding[0].Code {
				manuallyAddedPatientIDs = append(manuallyAddedPatientIDs, group.Member[i].Entity.ReferencedID)
				break
			}
		}
	}

	// Clear the huddle members list since we'll be repopulating it
	group.Member = nil

	// Start repopulating by adding back manually added patients (if applicable)
	for _, pid := range manuallyAddedPatientIDs {
		addPatientToHuddle(group, pid, &ManualAdditionReason)
	}

	encounterPatients, err := findEligiblePatientIDsByRecentEncounter(date, config)
	if err != nil {
		return nil, err
	}
	for i := range encounterPatients {
		reason := RecentEncounterReason
		reason.Text = encounterPatients[i].EventCode.Name
		addPatientToHuddle(group, encounterPatients[i].PatientID, &reason)
	}

	riskPatientIDs, err := findEligiblePatientIDsByRiskScore(date, config)
	if err != nil {
		return nil, err
	}
	for _, pid := range riskPatientIDs {
		addPatientToHuddle(group, pid, &RiskScoreReason)
	}

	carriedOverPatientIDs := findEligibleCarriedOverPatients(date, config)
	for _, pid := range carriedOverPatientIDs {
		addPatientToHuddle(group, pid, &CarriedOverReason)
	}

	return group, nil
}

// RiskScoreReason indicates that the patient was added to the huddle because his/her risk score warrants discussion
var RiskScoreReason = models.CodeableConcept{
	Coding: []models.Coding{
		{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "RISK_SCORE"},
	},
	Text: "Risk Score Warrants Discussion",
}

// RecentEncounterReason indicates that the patient was added to the huddle because a recent encounter (such as an ED
// visit) warrants discussion
var RecentEncounterReason = models.CodeableConcept{
	Coding: []models.Coding{
		{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "RECENT_ENCOUNTER"},
	},
	Text: "Recent Encounter",
}

// CarriedOverReason indicates that the patient was added to the huddle because he/she was scheduled for the last huddle
// but was not actually discussed
var CarriedOverReason = models.CodeableConcept{
	Coding: []models.Coding{
		{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "CARRIED_OVER"},
	},
	Text: "Carried Over From Last Huddle",
}

// ManualAdditionReason indicates that the patient was manually added to the huddle by a clinician.
var ManualAdditionReason = models.CodeableConcept{
	Coding: []models.Coding{
		{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "MANUAL_ADDITION"},
	},
	Text: "Manually Added",
}

func findExistingHuddle(date time.Time, config *HuddleConfig) (*models.Group, error) {
	searcher := search.NewMongoSearcher(server.Database)
	queryStr := fmt.Sprintf("leader=Practitioner/%s&activedatetime=%s&_sort=activedatetime&_count=1", config.LeaderID, date.Format("2006-01-02T-07:00"))
	var huddles []*models.Group
	if err := searcher.CreateQuery(search.Query{Resource: "Group", Query: queryStr}).All(&huddles); err != nil {
		return nil, err
	} else if len(huddles) > 0 {
		return huddles[0], nil
	}
	return nil, nil
}

type patientEvent struct {
	PatientID string
	EventCode EventCode
}

func findEligiblePatientIDsByRecentEncounter(date time.Time, config *HuddleConfig) ([]patientEvent, error) {
	var patientEvents []patientEvent
	patientMap := make(map[string]bool)
	if config.EventConfig == nil {
		return patientEvents, nil
	}
	// Loop through the event configs, looking for patients with matching encounters
	for _, eventConfig := range config.EventConfig.EncounterConfigs {
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
			return nil, err
		}

		// Go through the encounters finding the matches and storing in the patient map.  Note that FHIR search only
		// allows you to search on dates representing a time that happened at some point in the encounter -- so we must
		// post-process to see if the date is a real match.
		for _, result := range results {
			if patientMap[result.PatientID] {
				// Patient is already scheduled due to an event, so skip
				continue
			}
			for _, code := range eventConfig.TypeCodes {
				if codeMatches(result.Type, &code) {
					if d, matches := dateMatches(result.Period, &code, lowInclDate, highExclDate); matches {
						// If the patient has been discussed since the date, then don't schedule again
						alreadyDiscussed := false
						for i := range result.Huddles {
							h := Huddle(result.Huddles[i])
							if h.ActiveDateTime() != nil && h.ActiveDateTime().Time.After(d) {
								m := h.FindHuddleMember(result.PatientID)
								// Only consider it already discussed if the patient was discussed for this same reason
								alreadyDiscussed = m != nil && m.Reason().MatchesCode(RecentEncounterReason.Coding[0].System, RecentEncounterReason.Coding[0].Code)
								break
							}
						}
						if !alreadyDiscussed {
							patientEvents = append(patientEvents, patientEvent{
								PatientID: result.PatientID,
								EventCode: code,
							})
							patientMap[result.PatientID] = true
							break
						}
					}
				}
			}
		}
	}
	return patientEvents, nil
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

//func findLatestHuddleInfoForPatient(huddles []models.Group, patientID string, highDate time.Time) (date time.Time, reason models.CodeableConcept)

func findEligiblePatientIDsByRiskScore(date time.Time, config *HuddleConfig) ([]string, error) {
	if config.RiskConfig == nil {
		return []string{}, nil
	}

	var patientInfos []patientHuddleInfo
	var firstHuddle *time.Time

	// Now loop through each score-to-frequency configuration, finding the patients that are due.
	for _, frequency := range config.RiskConfig.FrequencyConfigs {
		patientsInRange, err := findPatientsInScoreRange(date, config.RiskConfig.RiskMethod, frequency.MinScore, frequency.MaxScore)
		if err != nil {
			return nil, err
		}

		var huddlelessPatients []patientHuddleInfo
		for _, p := range patientsInRange {
			if p.LastHuddle == nil {
				// collect the huddleless patients for special processing later
				huddlelessPatients = append(huddlelessPatients, p)
				continue
			}
			// Find the earliest allowed next huddle date and see if this huddle on or after it
			earliestNext := p.LastHuddle.AddDate(0, 0, frequency.MinDaysBetweenHuddles)
			if !date.Before(earliestNext) {
				patientInfos = append(patientInfos, p)
			}
		}

		if len(huddlelessPatients) > 0 {
			// To properly distribute the patients among the huddles, we must figure out how many huddles
			// to distribute the patients over.  Ideally, we want to distribute them equally over the
			// number of huddles representing the max time between huddles for this risk score.  For example,
			// if this is the first huddle, and patients should have a max of four weeks between huddles, then
			// we find how many huddles there are in the next four weeks (including this one), then we divide
			// the number of patients over that number of huddles, so we know how many patients should go in
			// each huddle.  That's the number we assign to *this* huddle.  If this is not the first huddle,
			// then we should use the first huddle to determine the max permissible huddle date, and then count
			// the huddles left from this date until that max date.  Capiche?

			// We only need to find the first huddle once, hence the pointer scoped out of the for loop
			if firstHuddle == nil {
				firstHuddle = findFirstHuddleDate("Practitioner/"+config.LeaderID, &date)
			}

			// Find the last possible date that is ok for the patients to be discussed
			lastDate := firstHuddle.AddDate(0, 0, frequency.MaxDaysBetweenHuddles)

			// Now figure out how many huddles we should distribute the patients over.
			numHuddles := calculateNumberOfHuddles(date, lastDate, config)

			// Divide patients by huddles.  If it's not even, round up.
			perHuddle := int(math.Ceil(float64(len(huddlelessPatients)) / float64(numHuddles)))

			// Now add just the number of patients for this huddle
			patientInfos = append(patientInfos, huddlelessPatients[:perHuddle]...)
		}
	}

	sort.Sort(byScoreAndHuddle(patientInfos))

	patientIDs := make([]string, len(patientInfos))
	for i := range patientInfos {
		patientIDs[i] = patientInfos[i].PatientID
	}
	return patientIDs, nil
}

func findEligibleCarriedOverPatients(date time.Time, config *HuddleConfig) []string {
	var patientIDs []string
	// Find all patients in most recent PAST huddle that were not actually discussed
	return patientIDs
}

// findPatientsInScoreRange finds the patients whose most recent risk assessment for the
// given method is in the given score range, and gets their huddles too -- so we can see
// when the most recent huddle was.
func findPatientsInScoreRange(huddleDate time.Time, method models.Coding, min float64, max float64) ([]patientHuddleInfo, error) {
	// For now we go straight to the database rather than using the MongoSearcher for a few reasons:
	// (1) The RiskAssessment resource doesn't define a search parameter for the score,
	// (2) using a pipeline and left join will allow us to do this all in one swoop, which isn't
	//     possible with MongoSearch because revinclude semantics don't apply here.
	riskQuery := bson.M{
		"method.coding": bson.M{
			"$elemMatch": bson.M{
				"system": method.System,
				"code":   method.Code,
			},
		},
		"prediction.probabilityDecimal": bson.M{
			"$gte": min,
			"$lte": max,
		},
		"meta.tag": bson.M{
			"$elemMatch": bson.M{
				"system": "http://interventionengine.org/tags/",
				"code":   "MOST_RECENT",
			},
		},
		"subject.external": false,
	}

	// This pipeline starts with the risk assessments in range, sorts them by risk score,
	// left-joins the huddles and then returns only the info we care about.
	pipeline := []bson.M{
		{"$match": riskQuery},
		{"$sort": bson.M{"prediction.probabilityDecimal": -1}},
		{"$lookup": bson.M{
			"from":         "groups",
			"localField":   "subject.referenceid",
			"foreignField": "member.entity.referenceid",
			"as":           "_groups",
		}},
		{"$project": bson.M{
			"_id":         0,
			"patientID":   "$subject.referenceid",
			"scores":      "$prediction.probabilityDecimal",
			"huddleDates": "$_groups.extension.activeDateTime.time",
		}},
	}

	var results []struct {
		PatientID   string         `bson:"patientID"`
		Scores      []float64      `bson:"scores"`
		HuddleDates [][]*time.Time `bson:"huddleDates"`
	}
	if err := server.Database.C("riskassessments").Pipe(pipeline).All(&results); err != nil {
		return nil, err
	}

	// Take the raw results from the mgo query and translate them into a list of
	// patients and their last huddle dates
	patientsInScoreRange := make([]patientHuddleInfo, len(results))
	for i := range results {
		p := patientHuddleInfo{PatientID: results[i].PatientID}
		if len(results[i].Scores) > 0 {
			p.Score = results[i].Scores[0]
		}
		for _, t := range results[i].HuddleDates {
			for _, t2 := range t {
				if t2.Before(huddleDate) && (p.LastHuddle == nil || t2.After(*p.LastHuddle)) {
					p.LastHuddle = t2
				}
			}
		}
		patientsInScoreRange[i] = p
	}

	return patientsInScoreRange, nil
}

// patientAndLastHuddle is a simple container for holding the info we need to schedule patients
// to huddles based on risk scores
type patientHuddleInfo struct {
	PatientID  string
	Score      float64
	LastHuddle *time.Time
}

// Support sorting by score and last huddle.  Higher scores go first.  When scores are equal,
// then the patient who has had the longest time since the last huddle (or no huddle at all)
// goes first.
type byScoreAndHuddle []patientHuddleInfo

func (s byScoreAndHuddle) Len() int {
	return len(s)
}
func (s byScoreAndHuddle) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byScoreAndHuddle) Less(i, j int) bool {
	if s[i].Score == s[j].Score {
		if s[i].LastHuddle == nil || s[j].LastHuddle == nil {
			return s[j].LastHuddle != nil
		}
		return s[i].LastHuddle.Before(*s[j].LastHuddle)
	}

	// Don't get tricked.  A higher score means you are earlier in the sort order.
	return s[i].Score > s[j].Score
}

// findFirstHuddleDate returns the date of the first huddle for this leader.  If no
// huddle was found, then it returns the default date passed in.
func findFirstHuddleDate(leaderRef string, defaultDate *time.Time) *time.Time {
	searcher := search.NewMongoSearcher(server.Database)
	queryStr := fmt.Sprintf("leader=%s&_sort=activedatetime&_count=1", leaderRef)
	var huddles []*models.Group
	if err := searcher.CreateQuery(search.Query{Resource: "Group", Query: queryStr}).All(&huddles); err != nil {
		return defaultDate
	}
	if len(huddles) != 0 {
		for i := range huddles[0].Extension {
			e := huddles[0].Extension[i]
			if e.Url == "http://interventionengine.org/fhir/extension/group/activeDateTime" && e.ValueDateTime != nil {
				return &e.ValueDateTime.Time
			}
		}
	}
	// No huddle date found, return the default
	return defaultDate
}

// calculateNumberOfHuddles figures out how many huddles there are between a starting huddle and an
// ending date, given the days of the week on which huddles occur.  There's a more efficient algorithm
// for this, but this one wins for simplicity!
func calculateNumberOfHuddles(startingHuddleDate time.Time, endingDate time.Time, config *HuddleConfig) int {
	numHuddles := 1 // start at 1 for the first huddle, regardless if its on a huddle day
	for d := startingHuddleDate.AddDate(0, 0, 1); !d.After(endingDate); d = d.AddDate(0, 0, 1) {
		if config.IsHuddleDay(d) {
			numHuddles++
		}
	}
	return numHuddles
}

func addPatientToHuddle(group *models.Group, id string, reason *models.CodeableConcept) {
	// First look to see if the patient is already in the group
	for i := range group.Member {
		if id == group.Member[i].Entity.ReferencedID {
			return
		}
	}

	// The patient is not yet in the group, so add him/her
	group.Member = append(group.Member, models.GroupMemberComponent{
		BackboneElement: models.BackboneElement{
			Element: models.Element{
				Extension: []models.Extension{
					{
						Url:                  "http://interventionengine.org/fhir/extension/group/member/reason",
						ValueCodeableConcept: reason,
					},
				},
			},
		},
		Entity: &models.Reference{
			Reference:    "Patient/" + id,
			ReferencedID: id,
			Type:         "Patient",
			External:     new(bool),
		},
	})
}

func printInfo(huddles []*models.Group, name string) {
	fmt.Printf("Scheduled %d huddles with name %s\n", len(huddles), name)
	for i := range huddles {
		fmt.Printf("\t%s: %d patients\n", getStringDate(huddles[i]), len(huddles[i].Member))
	}
}

func getStringDate(huddle *models.Group) string {
	for i := range huddle.Extension {
		if huddle.Extension[i].Url == "http://interventionengine.org/fhir/extension/group/activeDateTime" {
			t := huddle.Extension[i].ValueDateTime.Time
			return t.Format("01/02/2006")
		}
	}
	return ""
}
