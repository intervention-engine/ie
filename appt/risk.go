package appt

import "github.com/intervention-engine/ie/app"

func addPatientsBasedOnRiskScores(huddle *app.Huddle, huddleCount, targetSize int, itineraries itineraryMap) {
	for _, p := range itineraries.getPrioritizedPatientList(huddleCount) {
		// If we hit (or exceeded) our target, only stop if the patient *can* be put into a further huddle
		if len(huddle.Patients) >= targetSize && (p.FurthestAllowedHuddle == nil || huddleCount < *p.FurthestAllowedHuddle) {
			break
		}
		// If this huddle is before the nearest allowed, then we should not add this patient
		if p.NearestAllowedHuddle != nil && huddleCount < *p.NearestAllowedHuddle {
			continue
		}
		// Otherwise, add the patient to the huddle
		addPatient(huddle, p.ID, "Risk Score Warrants Discussion", "RISK_SCORE")
	}
}
