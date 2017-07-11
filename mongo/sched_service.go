package mongo

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"github.com/intervention-engine/ie/app"
	"github.com/intervention-engine/ie/storage"
	"gopkg.in/mgo.v2/bson"
)

// HuddleService provides functions to interact with huddles.
type SchedService struct {
	Service
}

func (s *SchedService) Close() {
	s.S.Close()
}

func (s *SchedService) CreateHuddles(huddles []*app.Huddle) error {
	// Store the huddles in the database
	var batchErr error
	for i := range huddles {
		_, err := s.C.UpsertId(huddles[i].ID, huddles[i])
		if err != nil {
			log.Printf("Error storing huddle: %s\n", err)
			batchErr = err
		}
	}
	return batchErr
}

func (s *SchedService) FindCareTeamHuddleOnDate(careTeamID string, date time.Time) (*app.Huddle, error) {
	mh := Huddle{}
	query := bson.M{
		"careteamid": careTeamID,
		"date": date.Format("2006-01-02T-07:00"),
	}
	err := s.C.Find(query).One(&mh)
	if err != nil {
		return nil, err
	}
	h := &mh.Huddle
	h.ID = &mh.ID
	return h, nil
}

func (s *SchedService) FindCareTeamHuddlesBefore(careTeamID string, date time.Time) ([]*app.Huddle, error) {
	mhh := []Huddle{}
	query := bson.M{
		"name": careTeamID,
		"date": struct {
			Lt  time.Time `bson:"$lt"`
		}{
			Lt: date,
		},
	}
	err := s.C.Find(query).All(&mhh)
	if err != nil {
		return nil, err
	}
	hh := make([]*app.Huddle, len(mhh), len(mhh))
	for i := range mhh {
		hh[i] = &mhh[i].Huddle
		hh[i].ID = &mhh[i].ID
	}
	return hh, nil
}

func (s *SchedService) RiskAssessmentsFilterBy(query storage.RiskFilterQuery) ([]*app.RiskAssessment, error) {
	mrr := []*models.RiskAssessment{}
	err := s.S.DB(s.Database).C(riskAssessmentCollection).Find(query).All(&mrr)
	if err != nil {
		return nil, err
	}
	return newAssessments(mrr), nil
}

func (s *SchedService) FindEncounters(typeCodes []string, earliestDate, latestDate time.Time) ([]storage.EncounterForSched, error) {
	// Don't bother looking for events in the future!
	if earliestDate.After(time.Now()) {
		return nil, errors.New("earliestDate is after current time")
	}
	pipeline := s.buildRecentEncounterQuery(typeCodes, earliestDate, latestDate)
	var results []storage.EncounterForSched
	err := s.S.DB(s.Database).C("encounters").Pipe(pipeline).All(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *SchedService) buildRecentEncounterQuery(codeVals []string, earliestDate, latestDate time.Time) []bson.M {
	// Build up the query to get all possible encounters that might trigger a huddle
	fmt := "2006-01-02T15:04:05.000-07:00"
	queryStr := "date=ge" + earliestDate.Format(fmt) + "&date=lt" + latestDate.Format(fmt) + "&status=arrived,in-progress,onleave,finished"
	if len(codeVals) > 0 {
		queryStr += "&type=" + strings.Join(codeVals, ",")
	}
	searcher := search.NewMongoSearcher(s.S.DB(s.Database))
	encQuery := searcher.CreateQueryObject(search.Query{Resource: "Encounter", Query: queryStr})

	// FOR NOW: We essentially copy/paste the encounter code we used in the previous version of the scheduler,
	// but we should revisit at some point since we may be able to streamline this WITHOUT a pipeline.

	// This pipeline starts with the encounter date/code query, sorts them by date, left-joins the huddles and then
	// returns only the info we care about.
	pipeline := []bson.M{
		{"$match": encQuery},
		{"$sort": bson.M{"period.start": -1}},
		{"$lookup": bson.M{
			"from":         "huddles",
			"localField":   "patient.ID",
			"foreignField": "patients.ID",
			"as":           "_huddles",
		}},
		{"$project": bson.M{
			"_id":       0,
			"patientID": "$patient.ID",
			"type":      1,
			"period":    1,
			"huddles":   "$_huddles",
		}},
	}
	return pipeline
}
