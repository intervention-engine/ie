package huddles

import (
	"fmt"
	"hash/fnv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"github.com/intervention-engine/fhir/server"
	"github.com/labstack/gommon/log"
)

type patientInfo struct {
	Huddles []int
	Score   *float64
}

// ScheduleHuddles schedules future huddles based on the passed in config.  It will schedule out the number
// of future huddles as specified in the config.LookAhead.
func ScheduleHuddles(config *HuddleConfig) ([]*Huddle, error) {
	log.Info("New Huddle Scheduler!")
	patientInfos := make(map[string]patientInfo)
	if err := populatePatientInfosWithRiskScore(patientInfos, config); err != nil {
		log.Error(err)
		return nil, err
	}

	numPastHuddles, err := populatePatientInfosWithPastHuddles(patientInfos, config)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Step through one day at a time, starting today, until we have created the requested number of huddles
	newHuddles, err := createInitialHuddles(numPastHuddles, patientInfos, config)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Get the target huddle size
	totalSlots := 0
	for i := range newHuddles {
		totalSlots += len(newHuddles[i].Member)
	}
	targetSize := totalSlots / config.LookAhead
	log.Info("TARGET SIZE: %d", targetSize)

	// Rebalance the huddles
	var maxPull, maxPush int
	for _, fc := range config.RiskConfig.FrequencyConfigs {
		pull := fc.IdealFrequency - fc.MinFrequency
		if pull > maxPull {
			maxPull = pull
		}
		push := fc.MaxFrequency - fc.IdealFrequency
		if push > maxPush {
			maxPush = push
		}
	}
	for i := range newHuddles {
		h := newHuddles[i]
		if len(h.Member) < targetSize {
			// Too small.  Try to pull patients from future huddles.
			for j := i + 1; j < len(newHuddles) && (j-i) <= maxPull && len(h.Member) < targetSize; j++ {
				h2 := newHuddles[j]
				for _, mem := range h2.HuddleMembers() {
					if mem.ReasonIsRiskScore() {
						info := patientInfos[mem.ID()]
						if info.Score != nil {
							rCfg := config.FindRiskScoreFrequencyConfigByScore(*info.Score)
							if (j - i) <= (rCfg.IdealFrequency - rCfg.MinFrequency) {
								// TODO: Make sure patient isn't already in target huddle (e.g. manual)
								// We can pull this patient to the current huddle!
								shiftMember(mem.ID(), j, i, numPastHuddles, newHuddles, patientInfos)
								if len(h.Member) >= targetSize {
									break
								}
							}
						}
					}
				}
			}
		} else if len(h.Member) > targetSize && (i+1) < len(newHuddles) {
			// Too big.  Push patients to future huddles.  Prefer the patients who can be pushed furthest
			for push := maxPush; push > 0 && len(h.Member) > targetSize; push-- {
				for _, mem := range h.HuddleMembers() {
					if mem.ReasonIsRiskScore() {
						info := patientInfos[mem.ID()]
						if info.Score != nil {
							// TODO: Fix bug -- it doesn't check if the patient really can be pushed further.
							// If this patient was already pushed once, it no longer is accurate
							rCfg := config.FindRiskScoreFrequencyConfigByScore(*info.Score)
							if (rCfg.MaxFrequency - rCfg.IdealFrequency) == push {
								shiftMember(mem.ID(), i, i+1, numPastHuddles, newHuddles, patientInfos)
								if len(h.Member) <= targetSize {
									break
								}
							}
						}
					}
				}
			}
		}
	}

	// Store the huddles in the database
	var lastErr error
	for i := range newHuddles {
		if _, err := server.Database.C("groups").UpsertId(newHuddles[i].Id, newHuddles[i]); err != nil {
			lastErr = err
			log.Warn("Error storing huddle: %t", err)
		}
	}

	printInfo(newHuddles, config.Name)

	return newHuddles, lastErr
}

func shiftMember(patientID string, fromNewHuddleIdx, toNewHuddleIdx, numPastHuddles int, newHuddles []*Huddle, patientInfos map[string]patientInfo) {
	pInfo := patientInfos[patientID]

	// First we must find the place in the patients huddle index containing the from huddle
	var fromLocationInPatientHuddleSlice int
	for i := range pInfo.Huddles {
		if pInfo.Huddles[i] == (numPastHuddles + fromNewHuddleIdx) {
			fromLocationInPatientHuddleSlice = i
		}
	}

	// Now start from there and move forward, shifting the huddles appropriately
	for i := fromLocationInPatientHuddleSlice; i < len(pInfo.Huddles); i++ {
		// TODO: This changes a bit if patient is in a huddle due to rollover or manual
		pInfo.Huddles[i] = pInfo.Huddles[i] - (fromNewHuddleIdx - toNewHuddleIdx)
		newHuddles[fromNewHuddleIdx].RemoveHuddleMember(patientID)
		newHuddles[toNewHuddleIdx].AddHuddleMemberDueToRiskScore(patientID)
	}
}

func populatePatientInfosWithRiskScore(patientInfos map[string]patientInfo, config *HuddleConfig) error {
	if config.RiskConfig == nil || len(config.RiskConfig.FrequencyConfigs) == 0 {
		return nil
	}

	// Find all of the patients in the scoring ranges used to schedule huddles
	// NOTE: We don't use IE search framework because prediction.probabilityDecimal is not a search parameter.
	// That said, we could consider creating a custom search param in the future if we really wanted...
	riskQuery := bson.M{
		"method.coding": bson.M{
			"$elemMatch": bson.M{
				"system": config.RiskConfig.RiskMethod.System,
				"code":   config.RiskConfig.RiskMethod.Code,
			},
		},
		"meta.tag": bson.M{
			"$elemMatch": bson.M{
				"system": "http://interventionengine.org/tags/",
				"code":   "MOST_RECENT",
			},
		},
		"subject.external": false,
	}

	if len(config.RiskConfig.FrequencyConfigs) == 1 {
		riskQuery["prediction.probabilityDecimal"] = bson.M{
			"$gte": config.RiskConfig.FrequencyConfigs[0].MinScore,
			"$lte": config.RiskConfig.FrequencyConfigs[0].MaxScore,
		}
	} else {
		ranges := make([]bson.M, len(config.RiskConfig.FrequencyConfigs))
		for i := range config.RiskConfig.FrequencyConfigs {
			ranges[i] = bson.M{
				"prediction.probabilityDecimal": bson.M{
					"$gte": config.RiskConfig.FrequencyConfigs[i].MinScore,
					"$lte": config.RiskConfig.FrequencyConfigs[i].MaxScore,
				},
			}
		}
		riskQuery["$or"] = ranges
	}

	selector := bson.M{
		"_id": 0,
		"subject.referenceid":           1,
		"prediction.probabilityDecimal": 1,
	}
	iter := server.Database.C("riskassessments").Find(riskQuery).Select(selector).Iter()
	result := models.RiskAssessment{}
	for iter.Next(&result) {
		info := patientInfos[result.Subject.ReferencedID]
		info.Score = result.Prediction[0].ProbabilityDecimal
		patientInfos[result.Subject.ReferencedID] = info
	}

	return iter.Close()
}

func populatePatientInfosWithPastHuddles(patientInfos map[string]patientInfo, config *HuddleConfig) (numHuddles int, err error) {
	// Find all of the huddles by the leader id and dates today or before
	searcher := search.NewMongoSearcher(server.Database)
	queryStr := fmt.Sprintf("leader=Practitioner/%s&activedatetime=le%s&_sort=activedatetime", config.LeaderID, time.Now().Format("2006-01-02T-07:00"))
	selector := bson.M{
		"_id": 0,
		"member.entity.referenceid": 1,
	}
	iter := searcher.CreateQueryWithoutOptions(search.Query{Resource: "Group", Query: queryStr}).Select(selector).Iter()
	i := 0
	result := models.Group{}
	for iter.Next(&result) {
		// Add the huddle to each member's patientInfo
		for _, member := range result.Member {
			info := patientInfos[member.Entity.ReferencedID]
			info.Huddles = append(info.Huddles, i)
			patientInfos[member.Entity.ReferencedID] = info
		}
		i++
	}

	return i, iter.Close()
}

func createInitialHuddles(numPastHuddles int, patientInfos map[string]patientInfo, config *HuddleConfig) ([]*Huddle, error) {
	// Step through one day at a time, starting today, until we have created the requested number of huddles
	i := numPastHuddles
	newHuddles := make([]*Huddle, config.LookAhead)
	for t := today(); i-numPastHuddles < config.LookAhead; t = t.AddDate(0, 0, 1) {
		if !config.IsHuddleDay(t) {
			continue
		}

		huddle := NewHuddle(config.Name, config.LeaderID, t)

		// Iterate through all the patientInfos, looking for patients whose next ideal huddle is this one
		for pID, pInfo := range patientInfos {
			if isIdealHuddle(i, pID, &pInfo, config) {
				huddle.AddHuddleMemberDueToRiskScore(pID)
				// TODO: Don't update patient info here.  Do it when the huddle is rebalanced.
				pInfo.Huddles = append(pInfo.Huddles, i)
				patientInfos[pID] = pInfo
			}
		}

		// Store it using a zero-based index (thus i-numPastHuddles)
		newHuddles[i-numPastHuddles] = huddle

		i++
	}
	return newHuddles, nil
}

// NOTE: This function is designed to be called in incrementing order over huddles.  In other words, you SHOULD NOT
// call this function for huddle 43 if you haven't yet scheduled huddle 42 OR if you already scheduled huddle 44!
func isIdealHuddle(hIdx int, pID string, pInfo *patientInfo, config *HuddleConfig) bool {
	if pInfo.Score == nil {
		return false
	}

	var idealFrequency *int
	for i := range config.RiskConfig.FrequencyConfigs {
		if *pInfo.Score >= config.RiskConfig.FrequencyConfigs[i].MinScore && *pInfo.Score <= config.RiskConfig.FrequencyConfigs[i].MaxScore {
			idealFrequency = &config.RiskConfig.FrequencyConfigs[i].IdealFrequency
		}
	}
	if idealFrequency == nil {
		return false
	}

	if len(pInfo.Huddles) == 0 {
		// This patient never had a huddle.  In this case, try to evenly distribute patients over n huddles
		// (where n = ideal frequency).  This isn't an exact science, but we'll use a hash and mod to do our best!
		// Example: https://play.golang.org/p/7kKKdub2OS
		modP := hash(pID) % uint32(*idealFrequency)
		modH := hIdx % *idealFrequency
		return modP == uint32(modH)
	}

	// Normal case in which patient does have previous huddles
	// Use >= to catch "overdue" huddles, since they are ideally ASAP
	return hIdx >= pInfo.Huddles[len(pInfo.Huddles)-1]+*idealFrequency
}

func today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func printInfo(huddles []*Huddle, name string) {
	fmt.Printf("Scheduled %d huddles with name %s\n", len(huddles), name)
	for i := range huddles {
		fmt.Printf("\t%s: %d patients\n", getStringDate(huddles[i]), len(huddles[i].Member))
	}
}

func getStringDate(huddle *Huddle) string {
	dt := huddle.ActiveDateTime()
	if dt != nil {
		return dt.Time.Format("01/02/2006")
	}
	return ""
}
