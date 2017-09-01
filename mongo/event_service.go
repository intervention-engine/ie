package mongo

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/ie/app"
	"github.com/intervention-engine/ie/storage"
	"gopkg.in/mgo.v2/bson"
)

// EventService provides functions to interact with huddles.
type EventService struct {
	Service
}

type Codes []Code

type Code struct {
	Text string
}

func (c Codes) Text() *string {
	for i := range c {
		return &c[i].Text
	}
	res := ""
	return &res
}

type Condition struct {
	ID           string `bson:"_id"`
	ResourceType string `bson:"resourceType"`
	Type         struct {
		Text string
	} `bson:"code"`
	OnsetDateTime struct {
		Time time.Time
	} `bson:"onsetDateTime"`
	AbatementDateTime struct {
		Time time.Time
	} `bson:"abatementDateTime"`
}

type Encounter struct {
	ID           string `bson:"_id"`
	ResourceType string `bson:"resourceType"`
	Type         Codes
	Period       struct {
		Start struct {
			Time time.Time
		}
		End struct {
			Time time.Time
		}
	}
}

type Medication struct {
	ID           string `bson:"_id"`
	ResourceType string `bson:"resourceType"`
	Type         struct {
		Text string
	} `bson:"medicationCodeableConcept"`
	EffectivePeriod struct {
		Start struct {
			Time time.Time
		}
		End struct {
			Time time.Time
		}
	} `bson:"effectivePeriod"`
}

// EventsFilterBy lists the events for a patient with the given filters.
// If no filters are given, all events are returned for that patient.
func (s *EventService) EventsFilterBy(query storage.EventFilterQuery) ([]*app.Event, error) {
	defer s.S.Close()
	if !bson.IsObjectIdHex(query.PatientID) {
		return nil, errors.New("bad patient id")
	}
	var types []string
	if query.Type == "" {
		types = []string{"condition", "encounter", "medication", "risk_change"}
	} else {
		types = strings.Split(query.Type, ",")
	}
	var multipleTypes bool
	if len(types) > 1 {
		multipleTypes = true
	}
	var errmessage string
	results := []*app.Event{}
	for _, t := range types {
		switch t {
		case "condition":
			q := bson.M{
				"patient.referenceid": query.PatientID,
			}
			if !query.Start.IsZero() {
				q["onsetDateTime.time"] = struct {
					Gte time.Time `bson:"$gte"`
				}{
					Gte: query.Start,
				}
			}
			if !query.End.IsZero() {
				q["abatementDateTime.time"] = struct {
					Lte time.Time `bson:"$lte"`
				}{
					Lte: query.End,
				}
			}
			cc := []Condition{}
			err := s.S.DB(s.Database).C("conditions").Find(q).All(&cc)
			if err != nil {
				// if this was the only type, we probably want to return err
				if !multipleTypes {
					return nil, err
				}
				// if this is not the only type, need some kind of propegation
				if err.Error() != "not found" {
					errmessage += err.Error() + " | "
				}
				log.Println(err.Error())
				continue
			}
			for i := range cc {
				r := &app.Event{
					ID:          &cc[i].ID,
					Type:        &cc[i].ResourceType,
					DisplayName: &cc[i].Type.Text,
					StartDate:   &cc[i].OnsetDateTime.Time,
					EndDate:     &cc[i].AbatementDateTime.Time,
				}
				results = append(results, r)
			}
		case "encounter":
			q := bson.M{
				"patient.referenceid": query.PatientID,
			}
			if !query.Start.IsZero() {
				q["period.start.time"] = struct {
					Gte time.Time `bson:"$gte"`
				}{
					Gte: query.Start,
				}
			}
			if !query.End.IsZero() {
				q["period.end.time"] = struct {
					Lt time.Time `bson:"$lt"`
				}{
					Lt: query.End,
				}
			}
			ee := []Encounter{}
			err := s.S.DB(s.Database).C("encounters").Find(q).All(&ee)
			if err != nil {
				// if this was the only type, we probably want to return err
				if !multipleTypes {
					return nil, err
				}
				// if this is not the only type, need some kind of propegation
				if err.Error() != "not found" {
					errmessage += err.Error() + " | "
				}
				log.Println(err.Error())
				continue
			}
			for i := range ee {
				r := &app.Event{
					ID:          &ee[i].ID,
					Type:        &ee[i].ResourceType,
					DisplayName: ee[i].Type.Text(),
					StartDate:   &ee[i].Period.Start.Time,
					EndDate:     &ee[i].Period.End.Time,
				}
				results = append(results, r)
			}
		case "medication":
			q := bson.M{
				"patient.referenceid": query.PatientID,
			}
			if !query.Start.IsZero() {
				q["effectivePeriod.start.time"] = struct {
					Gte time.Time `bson:"$gte"`
				}{
					Gte: query.Start,
				}
			}
			if !query.End.IsZero() {
				q["effectivePeriod.end.time"] = struct {
					Lte time.Time `bson:"$lte"`
				}{
					Lte: query.End,
				}
			}
			mm := []Medication{}
			err := s.S.DB(s.Database).C("medicationstatements").Find(q).All(&mm)
			if err != nil {
				// if this was the only type, we probably want to return err
				if !multipleTypes {
					return nil, err
				}
				// if this is not the only type, need some kind of propegation
				if err.Error() != "not found" {
					errmessage += err.Error() + " | "
				}
				log.Println(err.Error())
				continue
			}
			for i := range mm {
				r := &app.Event{
					ID:          &mm[i].ID,
					Type:        &mm[i].ResourceType,
					DisplayName: &mm[i].Type.Text,
					StartDate:   &mm[i].EffectivePeriod.Start.Time,
					EndDate:     &mm[i].EffectivePeriod.End.Time,
				}
				results = append(results, r)
			}
		case "risk_change":
			q := bson.M{
				"subject.referenceid": query.PatientID,
			}
			q["method.coding.code"] = query.RiskServiceID
			if !query.Start.IsZero() && query.End.IsZero() {
				q["date.time"] = struct {
					Gte time.Time `bson:"$gte"`
				}{
					Gte: query.Start,
				}
			} else if !query.Start.IsZero() && !query.End.IsZero() {
				q["date.time"] = struct {
					Gte time.Time `bson:"$gte"`
					Lte time.Time `bson:"$lte"`
				}{
					Gte: query.Start,
					Lte: query.End,
				}
			}
			var rr []models.RiskAssessment
			err := s.S.DB(s.Database).C("riskassessments").Find(q).Sort("date.time").All(&rr)
			if err != nil {
				// if this was the only type, we probably want to return err
				if !multipleTypes {
					return nil, err
				}
				// if this is not the only type, need some kind of propegation
				if err.Error() != "not found" {
					errmessage += err.Error() + " | "
				}
				log.Println(err.Error())
				continue
			}
			rc := calcRiskChanges(rr)
			results = append(results, rc...)
		default:
			return nil, errors.New(errmessage + "query param 'type' contained an invalid type: " + t)
		}
	}

	if errmessage != "" {
		return results, errors.New(errmessage)
	}
	return results, nil
}

func calcRiskChanges(rr []models.RiskAssessment) []*app.Event {
	cc := []*app.Event{}
	riskChangeTxt := "risk_change"
	done := map[string]bool{}
	for i := 0; i < len(rr); i++ {
		oldRisk := &rr[i]
		oldCategory := category(oldRisk.Prediction)
		oldValue := probability(oldRisk.Prediction)
		if _, ok := done[oldCategory]; ok {
			continue
		}
		done[oldCategory] = true
		for j := i + 1; j < len(rr); j++ {
			newRisk := &rr[j]
			newCategory := category(newRisk.Prediction)
			newValue := probability(newRisk.Prediction)
			if oldCategory != newCategory {
				continue
			}
			if oldValue == newValue {
				continue
			}
			oV := oldValue
			nV := newValue
			oD := oldRisk.Date.Time
			nD := oldRisk.Date.Time
			r := &app.Event{
				// Since this is a calculated thing, don't think it should return an ID.
				Type:        &riskChangeTxt,
				DisplayName: &oldCategory,
				StartDate:   &oD,
				EndDate:     &nD,
				OldValue:    &oV,
				NewValue:    &nV,
			}
			cc = append(cc, r)
			oldRisk = newRisk
			oldValue = newValue
		}
	}
	return cc
}

func probability(pred []models.RiskAssessmentPredictionComponent) float64 {
	for _, p := range pred {
		if p.ProbabilityDecimal != nil {
			return *p.ProbabilityDecimal
		}
	}
	return 0.0
}

func category(pred []models.RiskAssessmentPredictionComponent) string {
	for _, p := range pred {
		if p.Outcome != nil {
			return p.Outcome.Text
		}
	}
	return ""
}
