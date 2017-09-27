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

func (s *SchedService) FindCareTeamID(name string) (string, error) {
	var ct app.CareTeam
	err := s.S.DB(s.Database).C("care_teams").Find(bson.M{"name": name}).One(&ct)
	if err != nil {
		return "", err
	}
	if ct.ID != nil {
		return *ct.ID, nil
	}
	return "", nil
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

func (s *SchedService) FindCareTeamHuddlesBefore(careTeamName string, date time.Time) ([]*app.Huddle, error) {
	id, err := s.FindCareTeamID(careTeamName)
	log.Println("careteamID for FindCareTeamHuddlesBefore: ", id)
	if err != nil {
		return nil, err
	}
	mhh := []Huddle{}
	query := bson.M{
		"careteamid": id,
		"date": struct {
			Lt  time.Time `bson:"$lt"`
		}{
			Lt: date,
		},
	}
	err = s.C.Find(query).All(&mhh)
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

func (s *SchedService) RiskAssessmentsFilterBy(query storage.RiskFilterQuery) ([]*storage.RiskAssessment, error) {
	mrr := []*models.RiskAssessment{}
	q := parseRiskQuery(query)
	err := s.S.DB(s.Database).C(riskAssessmentCollection).Find(q).All(&mrr)
	if err != nil {
		return nil, err
	}
	return newStorageAssessments(mrr), nil
}

// Find all of the patients in the scoring ranges used to schedule huddles
func parseRiskQuery(query storage.RiskFilterQuery) bson.M {
	q := bson.M{
		"method.coding": bson.M{
					"$elemMatch": bson.M{
						"system": query.System,
						"code":   query.Code,
					},
				},
				//"meta.tag": bson.M{
				//	"$elemMatch": bson.M{
				//		"system": "http://interventionengine.org/tags/",
				//		"code":   "MOST_RECENT",
				//	},
				//},
		"subject.external": false,
	}
	if len(query.Values) > 0 {
		ranges := make([]map[string]interface{}, len(query.Values))
		for i, value := range query.Values {
			ranges[i] = bson.M{
				"prediction.probabilityDecimal": bson.M{
					"$gte": value[">"],
					"$lte": value["<"],
				},
			}
		}
		q["$or"] = ranges
		log.Println("mongo query for risk assessments: ", q)
		return q
	} else {
		q["prediction.probabilityDecimal"] = map[string]interface{}{
			"$gte": query.Value[">"],
			"$lte": query.Value["<"],
		}
	}
	log.Println("mongo query for risk assessments: ", q)
	return q
}

func newStorageAssessments(r []*models.RiskAssessment) []*storage.RiskAssessment {
	ra := make([]*storage.RiskAssessment, len(r))

	for i := 0; i < len(r); i++ {
		ra[i] = newStorageAssessment(r[i])
	}

	return ra
}

func newStorageAssessment(r *models.RiskAssessment) *storage.RiskAssessment {
	if r == nil {
		return nil
	}
	ra := &storage.RiskAssessment{}
	if (r.Prediction != nil) && (len(r.Prediction) > 0) {
		ra.Value = *r.Prediction[0].ProbabilityDecimal
	}
	ra.PatientID = r.Subject.ReferencedID
	return ra
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
