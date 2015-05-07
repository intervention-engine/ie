package middleware

import (
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	fhirmodels "github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
	"github.com/intervention-engine/ie/models"
	"gopkg.in/mgo.v2/bson"
)

func FactHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(rw, r)
	resourceType, ok := context.GetOk(r, "Resource")
	if ok {
		f := models.Fact{}
		resource := context.Get(r, resourceType)

		actionType := context.Get(r, "Action")
		if isFactAction(actionType.(string)) {
			switch t := resource.(type) {
			default:
				log.Printf("type of resource is %T", t)
			case *fhirmodels.Patient:
				f = models.FactFromPatient(resource.(*fhirmodels.Patient))
			case *fhirmodels.Condition:
				f = models.FactFromCondition(resource.(*fhirmodels.Condition))
			case *fhirmodels.Encounter:
				f = models.FactFromEncounter(resource.(*fhirmodels.Encounter))
			case *fhirmodels.Observation:
				f = models.FactFromObservation(resource.(*fhirmodels.Observation))
			case *fhirmodels.MedicationStatement:
				f = models.FactFromMedicationStatement(resource.(*fhirmodels.MedicationStatement), fhirmodels.BindMedicationLookup(server.Database))
			case string:
				//in this case we are handling the Delete method so we don't need to create a fact (we have no json resource to parse, just a targetID of the fact to delete)
			}
			ManageFactStorage(f, actionType.(string), rw, r)
		}
	}

}

func isFactAction(actionType string) bool {
	return actionType != "search" && actionType != "read"
}

func ManageFactStorage(f models.Fact, actionType string, rw http.ResponseWriter, r *http.Request) {

	var err error
	factCollection := server.Database.C("facts")

	switch actionType {
	case "create":
		err = factCollection.Insert(f)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	case "update":
		tempFact := models.Fact{}
		err = factCollection.Find(bson.M{"targetid": f.TargetID}).One(&tempFact)

		f.Id = tempFact.Id

		err = factCollection.Update(bson.M{"targetid": f.TargetID}, f)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	case "delete":
		err = factCollection.Remove(bson.M{"targetid": mux.Vars(r)["id"]})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}
}
