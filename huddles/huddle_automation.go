package huddles

import (
	"fmt"
	"time"

	"github.com/intervention-engine/fhir/models"
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

	riskPatientIDs := findEligiblePatientIDsByRiskScore(date, config)
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

func findEligiblePatientIDsByRiskScore(date time.Time, config *HuddleConfig) []string {
	var patientIDs []string
	for _, frequency := range config.RiskConfig.FrequencyConfigs {
		// Find all patients who's most recent score is in selected range AND last huddle date was more than the min
		// duration between huddles
		fmt.Println(frequency)
	}
	return patientIDs
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
