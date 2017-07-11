package appt

import (
	"fmt"
	"time"

	"github.com/intervention-engine/ie/app"
)

func addPatientsBasedOnRollOvers(huddle *app.Huddle, expiredHuddle *app.Huddle) {
	// Check for unreviewed patients
	for i := range expiredHuddle.Patients {
		if expiredHuddle.Patients[i].Reviewed == nil {
			addPatientDueToRollOver(huddle, expiredHuddle.Patients[i], *expiredHuddle.Date)
		}
	}
}

// addPatientDueToRollOver adds the patient to the huddle using the ROLLOVER and previous reason.
// If the patient is already in the huddle, nothing will be updated.
func addPatientDueToRollOver(huddle *app.Huddle, patient *app.HuddlePatient, from time.Time) {
	var reason string
	// TODO: check whether these Sprintf messages are still ok
	if *patient.Reason == "ROLLOVER" {
		reason = *patient.Reason
	} else if *patient.Reason == "MANUAL_ADDITION" {
		reason = fmt.Sprintf("Rolled Over from %s (Manually Added - %s)", from.Format("Jan 2"), patient.Reason)
	} else {
		reason = fmt.Sprintf("Rolled Over from %s (%s)", from.Format("Jan 2"), patient.Reason)
	}
	addPatient(huddle, *patient.ID, reason, "ROLLOVER")
}
