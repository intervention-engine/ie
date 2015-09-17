package notifications

import (
	"time"

	"github.com/intervention-engine/fhir/models"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	DefaultNotificationDefinitionRegistry.RegisterAll([]NotificationDefinition{
		AdmissionNotificationDefinition,
		ReadmissionNotificationDefinition,
		ERVisitNotificationDefinition})
}

type EncounterTypeNotificationDefinition struct {
	name                  string
	types                 []models.Coding
	reason                models.Coding
	additionalConstraints func(resource interface{}, action string) bool
}

func (def *EncounterTypeNotificationDefinition) Name() string {
	return def.name
}

func (def *EncounterTypeNotificationDefinition) Triggers(resource interface{}, action string) bool {
	if action != "create" {
		return false
	}

	encounter, ok := resource.(*models.Encounter)
	if ok && models.CodeableConcepts(encounter.Type).AnyMatchesAnyCode(def.types) {
		if def.additionalConstraints != nil {
			return def.additionalConstraints(resource, action)
		}
		return true
	}
	return false
}

func (def *EncounterTypeNotificationDefinition) GetNotification(resource interface{}, action string, baseURL string) *models.CommunicationRequest {
	if def.Triggers(resource, action) {
		encounter := resource.(*models.Encounter)
		cr := models.CommunicationRequest{}
		cr.Id = bson.NewObjectId().Hex()
		cr.Category = &models.CodeableConcept{Coding: make([]models.Coding, 1)}
		cr.Category.Coding[0].System = "http://snomed.info/sct"
		cr.Category.Coding[0].Code = "185087000"
		//cr.Recipient = TODO
		cr.Payload = make([]models.CommunicationRequestPayloadComponent, 1)
		cr.Payload[0].ContentReference = &models.Reference{
			Reference:    baseURL + "/Encounter/" + encounter.Id,
			Type:         "Encounter",
			ReferencedID: encounter.Id}
		cr.Status = "requested"
		cr.Reason = make([]models.CodeableConcept, 1)
		cr.Reason[0] = models.CodeableConcept{Coding: make([]models.Coding, 1)}
		cr.Reason[0].Coding[0] = def.reason
		cr.Subject = encounter.Patient
		cr.RequestedOn = &models.FHIRDateTime{Precision: models.Timestamp, Time: time.Now()}
		return &cr
	}
	return nil
}

// Definition of Encounter-based Notifications.  In the future, this could be externalized to a configuration file or database.

var AdmissionNotificationDefinition = &EncounterTypeNotificationDefinition{
	name:   "Inpatient Admission",
	reason: models.Coding{System: "http://snomed.info/sct", Code: "32485007", Display: "Hospital admission (procedure)"},
	additionalConstraints: func(resource interface{}, action string) bool {
		// Don't trigger the admission notification if it's also a re-admission
		return !ReadmissionNotificationDefinition.Triggers(resource, action)
	},
	types: []models.Coding{
		models.Coding{System: "http://snomed.info/sct", Code: "10378005", Display: "Hospital admission, emergency, from emergency room, accidental injury (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "112689000", Display: "Hospital admission, elective, with complete pre-admission work-up (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "112690009", Display: "Hospital admission, boarder, for social reasons (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "1505002", Display: "Hospital admission for isolation (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "15584006", Display: "Hospital admission, elective, with partial pre-admission work-up (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "18083007", Display: "Hospital admission, emergency, indirect (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "183430001", Display: "Holiday relief admission (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "183452005", Display: "Emergency hospital admission (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "183477006", Display: "Admit cardiothoracic emergency (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "183481006", Display: "Non-urgent hospital admission (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "183497001", Display: "Non-urgent trauma admission (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "19951005", Display: "Hospital admission, emergency, from emergency room, medical nature (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "2252009", Display: "Hospital admission, urgent, 48 hours (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "23473000", Display: "Hospital admission, for research investigation (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "25986004", Display: "Hospital admission, under police custody (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "266938001", Display: "Hospital patient (finding)"},
		models.Coding{System: "http://snomed.info/sct", Code: "2876009", Display: "Hospital admission, type unclassified, explain by report (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "304568006", Display: "Admission for respite care (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "305335007", Display: "Admission to establishment (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "305337004", Display: "Admission to community hospital (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "305338009", Display: "Admission to general practice hospital (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "305339001", Display: "Admission to private hospital (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "305341000", Display: "Admission to tertiary referral hospital (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "305342007", Display: "Admission to ward (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "305343002", Display: "Admission to day ward (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "305344008", Display: "Admission to day hospital (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "308540004", Display: "Inpatient stay (finding)"},
		models.Coding{System: "http://snomed.info/sct", Code: "313385005", Display: "Admit cardiology emergency (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "32485007", Display: "Hospital admission (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "36723004", Display: "Hospital admission, pre-nursing home placement (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "394656005", Display: "Inpatient care (regime/therapy)"},
		models.Coding{System: "http://snomed.info/sct", Code: "405614004", Display: "Unexpected hospital admission (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "416683003", Display: "Admit heart failure emergency (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "4563007", Display: "Hospital admission, transfer from other hospital or health care facility (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "45702004", Display: "Hospital admission, precertified by medical audit action (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "48183000", Display: "Hospital admission, special (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "50699000", Display: "Hospital admission, short-term (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "52748007", Display: "Hospital admission, involuntary (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "55402005", Display: "Hospital admission, for laboratory work-up, radiography, etc. (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "63551005", Display: "Hospital admission, from remote area, by means of special transportation (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "65043002", Display: "Hospital admission, short-term, day care (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "70755000", Display: "Hospital admission, by legal authority (commitment) (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "71290004", Display: "Hospital admission, limited to designated procedures (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "73607007", Display: "Hospital admission, emergency, from emergency room (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "78680009", Display: "Hospital admission, emergency, direct (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "81672003", Display: "Hospital admission, elective, without pre-admission work-up (procedure)"},
		models.Coding{System: "http://snomed.info/sct", Code: "8715000", Display: "Hospital admission, elective (procedure)"}}}

var ReadmissionNotificationDefinition = &EncounterTypeNotificationDefinition{
	name:   "Readmission",
	reason: models.Coding{System: "http://snomed.info/sct", Code: "417005", Display: "Hospital re-admission (procedure)"},
	types:  []models.Coding{models.Coding{System: "http://snomed.info/sct", Code: "417005", Display: "Hospital re-admission (procedure)"}}}

var ERVisitNotificationDefinition = &EncounterTypeNotificationDefinition{
	name:   "ER Visit",
	reason: models.Coding{System: "http://snomed.info/sct", Code: "4525004", Display: "Emergency department patient visit (procedure)"},
	types: []models.Coding{
		models.Coding{System: "http://snomed.info/sct", Code: "4525004", Display: "Emergency department patient visit (procedure)"},
		models.Coding{System: "http://www.ama-assn.org/go/cpt", Code: "99281", Display: "Emergency department visit..."},
		models.Coding{System: "http://www.ama-assn.org/go/cpt", Code: "99282", Display: "Emergency department visit..."},
		models.Coding{System: "http://www.ama-assn.org/go/cpt", Code: "99283", Display: "Emergency department visit..."},
		models.Coding{System: "http://www.ama-assn.org/go/cpt", Code: "99284", Display: "Emergency department visit..."},
		models.Coding{System: "http://www.ama-assn.org/go/cpt", Code: "99285", Display: "Emergency department visit..."}}}
