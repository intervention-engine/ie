package appt

import (
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/ie/app"
	"github.com/intervention-engine/ie/storage"
)

func addPatientsBasedOnRecentEncounters(huddle *app.Huddle, huddles []*app.Huddle, results []storage.EncounterForSched, typeCodes []EventCode, lowInclDate, highExclDate time.Time) error {
	// Go through the encounters finding the matches and storing in the patient map.  Note that FHIR search only
	// allows you to search on dates representing a time that happened at some point in the encounter -- so we must
	// post-process to see if the date is a real match.
	for _, result := range results {
		if findPatient(huddle, result.PatientID) != nil {
			// Patient is already scheduled, so skip
			continue
		}
		for _, code := range typeCodes {
			if !codeMatches(result.Type, code) {
				continue
			}
			d, match := dateMatches(result.Period, code, lowInclDate, highExclDate)
			if !match {
				continue
			}
			// Collect the PAST huddles, already in the database
			var hh []*app.Huddle
			for i := range result.Huddles {
				h := app.Huddle(result.Huddles[i])
				if h.Date != nil && h.Date.Before(*huddle.Date) {
					hh = append(hh, &h)
				}
			}
			if !isAlreadyDiscussed(hh, huddles, result.PatientID, d) {
				addPatient(huddle, result.PatientID, code.Name, "RECENT_ENCOUNTER")
				break
			}
		}
	}
	return nil
}

func isScheduledForSpecificEncounterReason(huddles []*app.Huddle, patientID string, encounterDate time.Time) bool {
	// Go through the huddles backwards since it's more likely a discussed date is recent (although we can't guarantee the huddles are sorted)
	for i := len(huddles) - 1; i >= 0; i-- {
		h := huddles[i]
		if h.Date != nil && !h.Date.Equal(encounterDate) && h.Date.After(encounterDate) {
			p := findPatient(h, patientID)
			// Only consider it already discussed if the patient was discussed for this same reason
			if p != nil && *p.Reason == "RECENT_ENCOUNTER" {
				return true
			}
		}
	}
	return false
}

func isAlreadyDiscussed(hh, huddles []*app.Huddle, patientID string, date time.Time) bool {
	// If the patient has been discussed in a huddle since the date, then don't schedule again
	alreadyDiscussed := isScheduledForSpecificEncounterReason(hh, patientID, date)
	if !alreadyDiscussed {
		// Then check the huddles we've already scheduled in this session
		alreadyDiscussed = isScheduledForSpecificEncounterReason(huddles, patientID, date)
	}
	return alreadyDiscussed
}

func codeMatches(encType []models.CodeableConcept, code EventCode) bool {
	return models.CodeableConcepts(encType).AnyMatchesCode(code.System, code.Code)
}

func dateMatches(encPeriod *models.Period, code EventCode, lowIncl, highExcl time.Time) (time.Time, bool) {
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
