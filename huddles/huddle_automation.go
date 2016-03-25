package huddles

import (
	"fmt"
	"math"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"github.com/intervention-engine/fhir/server"
)

// CreateAutoPopulatedHuddle returns a Group resource representing the patients that should be automatically considered
// for a huddle for the specific date.  Currently it is based on three criteria:
// - Risk scores (which determine frequency)
// - Recent clinical events (such as ED visit)
// - "Leftovers" from previous huddle
func CreateAutoPopulatedHuddle(date time.Time, config *HuddleConfig) *models.Group {
	group := findExistingHuddle(date, config)
	if group == nil {
		tru := true
		group = &models.Group{
			DomainResource: models.DomainResource{
				Resource: models.Resource{
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

	riskPatientIDs, _ := findEligiblePatientIDsByRiskScore(date, config)
	for _, pid := range riskPatientIDs {
		addPatientToHuddle(group, pid, &RiskScoreReason)
	}
	encounterPatientIDs := findEligiblePatientIDsByRecentEncounter(date, config)
	for _, pid := range encounterPatientIDs {
		addPatientToHuddle(group, pid, &RecentEncounterReason)
	}
	carriedOverPatientIDs := findEligibleCarriedOverPatients(date, config)
	for _, pid := range carriedOverPatientIDs {
		addPatientToHuddle(group, pid, &CarriedOverReason)
	}

	return group
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
	Text: "Recent Encounter Warrants Discussion",
}

// CarriedOverReason indicates that the patient was added to the huddle because he/she was scheduled for the last huddle
// but was not actually discussed
var CarriedOverReason = models.CodeableConcept{
	Coding: []models.Coding{
		{System: "http://interventionengine.org/fhir/cs/huddle-member-reason", Code: "CARRIED_OVER"},
	},
	Text: "Carried Over From Last Huddle",
}

func findExistingHuddle(date time.Time, config *HuddleConfig) *models.Group {
	return nil
}

func findEligiblePatientIDsByRiskScore(date time.Time, config *HuddleConfig) ([]string, error) {
	var patientIDs []string
	var firstHuddle *time.Time

	// Now loop through each score-to-frequency configuration, finding the patients that are due.
	for _, frequency := range config.RiskConfig.FrequencyConfigs {
		patientsInRange, err := findPatientsInScoreRange(config.RiskConfig.RiskCode, frequency.MinScore, frequency.MaxScore)
		if err != nil {
			return nil, err
		}

		var huddlelessPatients []string
		for _, p := range patientsInRange {
			if p.LastHuddle == nil {
				// collect the huddleless patients for special processing later
				huddlelessPatients = append(huddlelessPatients, p.PatientID)
				continue
			}
			// If the elapsed time is more than the minimum time allowed, then the patient is eligible
			elapsed := date.Sub(*p.LastHuddle)
			if elapsed > frequency.MinTimeBetweenHuddles {
				patientIDs = append(patientIDs, p.PatientID)
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
			lastDate := firstHuddle.Add(frequency.MaxTimeBetweenHuddles)

			// Now figure out how many huddles we should distribute the patients over.
			numHuddles := calculateNumberOfHuddles(date, lastDate, config.Days)

			// Divide patients by huddles.  If it's not even, round up.
			perHuddle := int(math.Ceil(float64(len(huddlelessPatients)) / float64(numHuddles)))

			// Now add just the number of patients for this huddle
			patientIDs = append(patientIDs, huddlelessPatients[:perHuddle]...)
		}
	}
	return patientIDs, nil
}

func findEligiblePatientIDsByRecentEncounter(date time.Time, config *HuddleConfig) []string {
	var patientIDs []string
	// Find all patients with Encounters of specified config.EncounterCodes with admits or discharges since last huddle
	return patientIDs
}

func findEligibleCarriedOverPatients(date time.Time, config *HuddleConfig) []string {
	var patientIDs []string
	// Find all patients in most recent PAST huddle that were not actually discussed
	return patientIDs
}

// findPatientsInScoreRange finds the patients whose most recent risk assessment for the
// given method is in the given score range, and gets their huddles too -- so we can see
// when the most recent huddle was.
func findPatientsInScoreRange(method models.Coding, min float64, max float64) ([]patientAndLastHuddle, error) {
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
			"huddleDates": "$_groups.extension.activeDateTime.time",
		}},
	}

	var results []struct {
		PatientID   string         `bson:"patientID"`
		HuddleDates [][]*time.Time `bson:"huddleDates"`
	}
	if err := server.Database.C("riskassessments").Pipe(pipeline).All(&results); err != nil {
		return nil, err
	}

	// Take the raw results from the mgo query and translate them into a list of
	// patients and their last huddle dates
	patientsInScoreRange := make([]patientAndLastHuddle, len(results))
	for i := range results {
		p := patientAndLastHuddle{PatientID: results[i].PatientID}
		for _, t := range results[i].HuddleDates {
			for _, t2 := range t {
				if p.LastHuddle == nil || t2.After(*p.LastHuddle) {
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
type patientAndLastHuddle struct {
	PatientID  string
	LastHuddle *time.Time
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
func calculateNumberOfHuddles(startingHuddleDate time.Time, endingDate time.Time, days []time.Weekday) int {
	numHuddles := 1 // start at 1 for the first huddle, regardless if its on a huddle day
	for d := startingHuddleDate.Add(24 * time.Hour); !d.After(endingDate); d = d.Add(24 * time.Hour) {
		for _, wd := range days {
			if d.Weekday() == wd {
				numHuddles++
				break
			}
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
