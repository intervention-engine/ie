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

			var patientID string
			var timestamp time.Time

			switch resource.(type) {
			case *fhirmodels.Condition:
				patientID = resource.(*fhirmodels.Condition).Patient.ReferencedID
				timestamp = resource.(*fhirmodels.Condition).OnsetDateTime.Time
			case *fhirmodels.MedicationStatement:
				patientID = resource.(*fhirmodels.MedicationStatement).Patient.ReferencedID
				timestamp = resource.(*fhirmodels.MedicationStatement).EffectivePeriod.Start.Time
			}

			ru := NewResourceUpdateMessage(patientID, timestamp.Format(time.RFC3339))
			subUpdateQueue <- ru
		}
	}
	return f
}
