package subscription

import (
	"net/http"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	fhirmodels "github.com/intervention-engine/fhir/models"
)

func GenerateResourceWatch(subUpdateQueue chan<- ResourceUpdateMessage) negroni.HandlerFunc {
	f := func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r)
		resourceType, ok := context.GetOk(r, "Resource")
		if ok {
			resource := context.Get(r, resourceType)
			HandleResourceUpdate(subUpdateQueue, resource)
		}
	}
	return f
}

func HandleResourceUpdate(subUpdateQueue chan<- ResourceUpdateMessage, resource interface{}) {
	var patientID string
	var timestamp time.Time

	switch t := resource.(type) {
	case *fhirmodels.Condition:
		patientID = t.Patient.ReferencedID
		timestamp = t.OnsetDateTime.Time
	case *fhirmodels.MedicationStatement:
		patientID = t.Patient.ReferencedID
		timestamp = t.EffectivePeriod.Start.Time
	case *fhirmodels.Encounter:
		patientID = t.Patient.ReferencedID
		timestamp = t.Period.Start.Time
	case *fhirmodels.Bundle:
		for _, entry := range t.Entry {
			HandleResourceUpdate(subUpdateQueue, entry.Resource)
		}
		return
	default:
		return
	}

	ru := NewResourceUpdateMessage(patientID, timestamp.Format(time.RFC3339))
	subUpdateQueue <- ru
}
