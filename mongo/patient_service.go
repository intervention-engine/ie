package mongo

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/ie/app"
	"gopkg.in/mgo.v2/bson"
)

// PatientService provides a mongo implementation of a
// Storage Service for patients.
type PatientService struct {
	Service
}

// Patient gets a patient with the given id.
func (s *PatientService) Patient(id string) (*app.Patient, error) {
	defer s.S.Close()
	if !bson.IsObjectIdHex(id) {
		return nil, errors.New("bad id")
	}
	var data models.Patient
	err := s.C.FindId(id).One(&data)
	if err != nil {
		return nil, err
	}
	current := getActiveForPatient(s.C.Database, id)
	p := newPatient(data, &current)
	recentRisk, err := s.findRecentRiskAssessment(id)
	if err != nil && err.Error() == "not found" {
		return p, nil
	}
	if err != nil {
		return nil, err
	}
	p.RecentRiskAssessment = newAssessment(&recentRisk)
	p, err = s.addNextHuddleToPatient(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PatientService) addNextHuddleToPatient(patient *app.Patient) (*app.Patient, error) {
	// Go through every patient, execute a query for that patient and populate
	// their next huddle field if they have one.
	var nh app.Huddle
	q := bson.M{
		"patients.id": patient.ID,
		"date": struct {
			Gte time.Time `bson:"$gte"`
		}{
			Gte: time.Now(),
		},
	}
	err := s.S.DB(s.Database).C("huddles").Find(q).One(&nh)
	if err != nil {
		if err.Error() == "not found" {
			return patient, nil
		}
		return nil, err
	}
	var details *app.PatientHuddle
	for _, p := range nh.Patients {
		if *p.ID == patient.ID {
			details = p
			break
		}
	}
	patient.NextHuddle = &app.NextHuddle{}
	patient.NextHuddle.HuddleID = nh.ID
	patient.NextHuddle.HuddleDate = nh.Date
	if details != nil {
		patient.NextHuddle.Reason = details.Reason
		patient.NextHuddle.ReasonType = details.ReasonType
		patient.NextHuddle.Reviewed = details.Reviewed
		patient.NextHuddle.ReviewedAt = details.ReviewedAt
	}
	var ct app.CareTeam
	err = s.S.DB(s.Database).C("care_teams").FindId(nh.CareTeamID).One(&ct)
	if err != nil {
		if err.Error() == "not found" {
			return patient, nil
		}
		return nil, err
	}
	patient.NextHuddle.CareTeamName = ct.Name

	return patient, nil
}

// Patients gets all the patients in the db.
func (s *PatientService) Patients(filter map[string]string) ([]*app.Patient, error) {
	defer s.S.Close()
	var data []models.Patient
	err := s.C.Find(nil).All(&data)
	if err != nil {
		return nil, err
	}
	pp := newPatients(data)
	if len(filter) != 0 {
		if id, ok := filter["care_team_id"]; ok {
			pp, err = s.filterCareTeam(id, pp)
			if err != nil {
				return nil, err
			}
		}
		if id, ok := filter["huddle_id"]; ok {
			pp, err = s.filterHuddle(id, pp)
			if err != nil {
				return nil, err
			}
		}
		if term, ok := filter["search_term"]; ok {
			pp, err = s.filterName(term, pp)
			if err != nil {
				return nil, err
			}
		}
	}
	pp, err = s.addNextHuddleToPatients(pp)
	err = s.addRecentRiskAssessments(pp)
	return pp, err
}

// SortBy gets patients sorted by the fields given.
func (s *PatientService) PatientsSortBy(filter map[string]string, fields ...string) ([]*app.Patient, error) {
	defer s.S.Close()
	var data []models.Patient
	log.Println("fields is: ", fields)
	query, err := convertQuery(fields...)
	if err != nil {
		return nil, err
	}
	log.Println("query is:", query)
	err = s.C.Find(nil).Sort(query...).All(&data)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	pp := newPatients(data)
	if len(filter) != 0 {
		if id, ok := filter["care_team_id"]; ok {
			pp, err = s.filterCareTeam(id, pp)
			if err != nil {
				return nil, err
			}
		}
		if id, ok := filter["huddle_id"]; ok {
			pp, err = s.filterHuddle(id, pp)
			if err != nil {
				return nil, err
			}
		}
		if term, ok := filter["search_term"]; ok {
			pp, err = s.filterName(term, pp)
			if err != nil {
				return nil, err
			}
		}
	}
	pp, err = s.addNextHuddleToPatients(pp)
	if err != nil {
		return nil, err
	}
	err = s.addRecentRiskAssessments(pp)
	return pp, err
}

func (s *PatientService) addNextHuddleToPatients(pp []*app.Patient) ([]*app.Patient, error) {
	// Go through every patient, execute a query for that patient and populate
	// their next huddle field if they have one.
	for i := range pp {
		var nh app.Huddle
		q := bson.M{
			"patients.id": pp[i].ID,
			"date": struct {
				Gte time.Time `bson:"$gte"`
			}{
				Gte: time.Now(),
			},
		}
		err := s.S.DB(s.Database).C("huddles").Find(q).One(&nh)
		if err != nil {
			if err.Error() == "not found" {
				continue
			}
			return nil, err
		}
		var details *app.PatientHuddle
		for _, p := range nh.Patients {
			if *p.ID == pp[i].ID {
				details = p
				break
			}
		}
		pp[i].NextHuddle = &app.NextHuddle{}
		pp[i].NextHuddle.HuddleID = nh.ID
		pp[i].NextHuddle.HuddleDate = nh.Date
		if details != nil {
			pp[i].NextHuddle.Reason = details.Reason
			pp[i].NextHuddle.ReasonType = details.ReasonType
			pp[i].NextHuddle.Reviewed = details.Reviewed
			pp[i].NextHuddle.ReviewedAt = details.ReviewedAt
		}
		var ct app.CareTeam
		err = s.S.DB(s.Database).C("care_teams").FindId(nh.CareTeamID).One(&ct)
		if err != nil {
			if err.Error() == "not found" {
				continue
			}
			return nil, err
		}
		pp[i].NextHuddle.CareTeamName = ct.Name
	}

	return pp, nil
}

func (s *PatientService) filterCareTeam(ID string, patients []*app.Patient) ([]*app.Patient, error) {
	// Go find all the patients in this care team. Then iterate over patients
	// and only keep the ones in this care team.
	var rel []struct {
		CareTeamID string `bson:"care_team_id"`
		PatientID  string `bson:"patient_id"`
	}
	err := s.S.DB(s.Database).C("care_team_patient").Find(bson.M{"care_team_id": ID}).All(&rel)
	if err != nil {
		return nil, err
	}
	filtered := []*app.Patient{}
	for _, p := range patients {
		for _, r := range rel {
			if p.ID == r.PatientID {
				filtered = append(filtered, p)
			}
		}
	}
	return filtered, nil
}

func (s *PatientService) filterHuddle(ID string, patients []*app.Patient) ([]*app.Patient, error) {
	// Go find all the patients in this huddle. Then iterate over patients
	// and only keep the ones in this huddle.
	var huddle app.Huddle
	err := s.S.DB(s.Database).C("huddles").FindId(ID).One(&huddle)
	if err != nil {
		return nil, err
	}
	filtered := []*app.Patient{}
	for _, p := range patients {
		for _, hp := range huddle.Patients {
			if p.ID == *hp.ID {
				filtered = append(filtered, p)
			}
		}
	}
	return filtered, nil
}

func (s *PatientService) filterName(term string, patients []*app.Patient) ([]*app.Patient, error) {
	// Ierate over patients and only keep the ones whose names match the term in
	// some way.
	filtered := []*app.Patient{}
	for _, p := range patients {
		if p.Name != nil {
			if nameContains(*p.Name, term) {
				filtered = append(filtered, p)
			}
		}
	}
	return filtered, nil
}

func nameContains(name app.Name, term string) bool {
	if name.Family != nil {
		if strings.Contains(*name.Family, term) {
			return true
		}
	}
	if name.Given != nil {
		if strings.Contains(*name.Given, term) {
			return true
		}
	}
	if name.Full != nil {
		if strings.Contains(*name.Full, term) {
			return true
		}
	}
	if name.MiddleInitial != nil {
		if strings.Contains(*name.MiddleInitial, term) {
			return true
		}
	}
	return false
}

func (s *PatientService) addRecentRiskAssessments(pp []*app.Patient) error {
	for i, p := range pp {
		recentRisk, err := s.findRecentRiskAssessment(p.ID)
		if err != nil && err.Error() == "not found" {
			continue
		}
		if err != nil {
			return err
		}
		pp[i].RecentRiskAssessment = newAssessment(&recentRisk)
	}
	return nil
}

func (s *PatientService) findRecentRiskAssessment(id string) (models.RiskAssessment, error) {
	var recentRisk models.RiskAssessment
	rCol := s.S.DB(s.Database).C(riskAssessmentCollection)
	err := rCol.Find(bson.M{"subject.referenceid": id}).Sort("-date.time").One(&recentRisk)
	return recentRisk, err
}

func newPatients(patients []models.Patient) []*app.Patient {
	pp := make([]*app.Patient, len(patients))
	for i, patient := range patients {
		pp[i] = newPatient(patient, nil)
	}
	return pp
}

var conversions = map[string]string{
	"name":       "name",
	"-name":      "-name",
	"birthdate":  "birthDate.time",
	"-birthdate": "-birthDate.time",
	"gender":     "gender",
	"-gender":    "-gender",
	// Sorting by address evidently wasn't useful and sorting by name and address causes an error in mongo.
	//"address": "address.postalCode",
	//"-address": "-address.postalCode",
	// Haven't technically been implemented yet.
	//"riskScore": "riskScore",
	//"-riskScore": "-riskScore",
	//"notifications": "notifications",
	//"-notifications": "-notifications",
}

func convertQuery(fields ...string) ([]string, error) {
	query := make([]string, len(fields))
	for i, field := range fields {
		log.Println("field given is: ", field)
		conv, ok := conversions[field]
		if !ok {
			return nil, errors.New("had no conversion for: " + field)
		}
		query[i] = conv
		log.Println("query now is: ", query)
	}
	return query, nil
}

func newPatient(fhirPatient models.Patient, current *CurrentActiveElements) *app.Patient {
	p := app.Patient{}
	p.ID = fhirPatient.Id
	if len(fhirPatient.Address) > 0 {
		p.Address = newAddress(fhirPatient.Address[0])
	}
	if fhirPatient.BirthDate != nil {
		age := age(fhirPatient.BirthDate)
		p.Age = &age
	}
	p.Gender = &fhirPatient.Gender
	p.BirthDate = &fhirPatient.BirthDate.Time
	if len(fhirPatient.Name) > 0 {
		p.Name = newName(fhirPatient.Name[0])
	}

	if current != nil {
		p.CurrentAllergies = current.Allergies
		p.CurrentConditions = current.Conditions
		p.CurrentMedications = current.Medications
	}

	// p.NextHuddleID = &fhirPatient.NextHuddleID
	// p.RecentRiskAssessments = fhirPatient.RiskAssessments
	return &p
}

func newName(name models.HumanName) *app.Name {
	n := app.Name{}
	var family, given string
	if len(name.Given) > 0 {
		given = name.Given[0]
		n.Given = &given
	}
	if len(name.Family) > 0 {
		family = name.Family[0]
		n.Family = &family
	}
	if (given != "") && (family != "") {
		full := fmt.Sprintf("%s %s", given, family)
		n.Full = &full
	} else if given != "" {
		n.Full = &given
	} else if family != "" {
		n.Full = &family
	}
	return &n
}

func newAddress(address models.Address) *app.Address {
	a := app.Address{}
	a.Street = address.Line
	a.City = &address.City
	a.State = &address.State
	a.PostalCode = &address.PostalCode
	return &a
}

func age(birthday *models.FHIRDateTime) int {
	now := time.Now()
	years := now.Year() - birthday.Time.Year()

	if now.YearDay() < birthday.Time.YearDay() {
		years--
	}

	return years
}
