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

// HuddleScheduler schedules future huddles based on the passed in config.
type HuddleScheduler struct {
	Config          *HuddleConfig
	FutureHuddles   []*Huddle
	patientInfos    patientInfoMap
	pastHuddleCount int
}

// NewHuddleScheduler initializes a new huddle scheduler based on the passed in config.
func NewHuddleScheduler(config *HuddleConfig) *HuddleScheduler {
	return &HuddleScheduler{
		Config:       config,
		patientInfos: make(patientInfoMap),
	}
}

type patientInfoMap map[string]*patientInfo

func (p patientInfoMap) SafeGet(key string) *patientInfo {
	pInfo := p[key]
	if pInfo == nil {
		pInfo = &patientInfo{}
		p[key] = pInfo
	}
	return pInfo
}

type patientInfo struct {
	Huddles []int
	Score   *float64
}

// ScheduleHuddles schedules future huddles based on the passed in config.  It will schedule out the number
// of future huddles as specified in the config.LookAhead.
func (hs *HuddleScheduler) ScheduleHuddles() ([]*Huddle, error) {
	if err := hs.populatePatientInfosWithRiskScore(); err != nil {
		log.Error(err)
		return nil, err
	}

	err := hs.populatePatientInfosWithPastHuddles()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = hs.createInitialHuddles()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = hs.rebalanceHuddles()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Store the huddles in the database
	var lastErr error
	for i := range hs.FutureHuddles {
		if _, err := server.Database.C("groups").UpsertId(hs.FutureHuddles[i].Id, hs.FutureHuddles[i]); err != nil {
			lastErr = err
			log.Warn("Error storing huddle: %t", err)
		}
	}

	hs.printInfo()

	return hs.FutureHuddles, lastErr
}

func (hs *HuddleScheduler) populatePatientInfosWithRiskScore() error {
	if hs.Config.RiskConfig == nil || len(hs.Config.RiskConfig.FrequencyConfigs) == 0 {
		return nil
	}

	// Find all of the patients in the scoring ranges used to schedule huddles
	// NOTE: We don't use IE search framework because prediction.probabilityDecimal is not a search parameter.
	// That said, we could consider creating a custom search param in the future if we really wanted...
	riskQuery := bson.M{
		"method.coding": bson.M{
			"$elemMatch": bson.M{
				"system": hs.Config.RiskConfig.RiskMethod.System,
				"code":   hs.Config.RiskConfig.RiskMethod.Code,
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

	if len(hs.Config.RiskConfig.FrequencyConfigs) == 1 {
		frqCfg := hs.Config.RiskConfig.FrequencyConfigs[0]
		riskQuery["prediction.probabilityDecimal"] = bson.M{
			"$gte": frqCfg.MinScore,
			"$lte": frqCfg.MaxScore,
		}
	} else {
		ranges := make([]bson.M, len(hs.Config.RiskConfig.FrequencyConfigs))
		for i := range hs.Config.RiskConfig.FrequencyConfigs {
			frqCfg := hs.Config.RiskConfig.FrequencyConfigs[i]
			ranges[i] = bson.M{
				"prediction.probabilityDecimal": bson.M{
					"$gte": frqCfg.MinScore,
					"$lte": frqCfg.MaxScore,
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
		info := hs.patientInfos.SafeGet(result.Subject.ReferencedID)
		info.Score = result.Prediction[0].ProbabilityDecimal
	}

	return iter.Close()
}

func (hs *HuddleScheduler) populatePatientInfosWithPastHuddles() error {
	// Find all of the huddles by the leader id and dates today or before
	searcher := search.NewMongoSearcher(server.Database)
	queryStr := fmt.Sprintf("leader=Practitioner/%s&activedatetime=le%s&_sort=activedatetime", hs.Config.LeaderID, time.Now().Format("2006-01-02T-07:00"))
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
			info := hs.patientInfos.SafeGet(member.Entity.ReferencedID)
			info.Huddles = append(info.Huddles, i)
		}
		i++
	}

	// Set the pastHuddleCount so we can use it elsewhere
	hs.pastHuddleCount = i

	return iter.Close()
}

func (hs *HuddleScheduler) createInitialHuddles() error {
	// Step through one day at a time, starting today, until we have created the requested number of huddles
	hs.FutureHuddles = make([]*Huddle, 0, hs.Config.LookAhead)
	for t := today(); len(hs.FutureHuddles) < hs.Config.LookAhead; t = t.AddDate(0, 0, 1) {
		if !hs.Config.IsHuddleDay(t) {
			continue
		}

		huddle := NewHuddle(hs.Config.Name, hs.Config.LeaderID, t)
		totalHuddleIdx := hs.pastHuddleCount + len(hs.FutureHuddles)

		// Iterate through all the patientInfos, looking for patients whose next ideal huddle is this one
		for pID, pInfo := range hs.patientInfos {
			if hs.isIdealHuddle(totalHuddleIdx, pID) {
				huddle.AddHuddleMemberDueToRiskScore(pID)
				pInfo.Huddles = append(pInfo.Huddles, totalHuddleIdx)
			}
		}

		hs.FutureHuddles = append(hs.FutureHuddles, huddle)
	}
	return nil
}

// NOTE: This function is designed to be called in incrementing order over huddles.  In other words, you SHOULD NOT
// call this function for huddle 43 if you haven't yet scheduled huddle 42 OR if you already scheduled huddle 44!
func (hs *HuddleScheduler) isIdealHuddle(hIdx int, pID string) bool {
	pInfo := hs.patientInfos.SafeGet(pID)
	if pInfo.Score == nil {
		return false
	}

	frqCfg := hs.Config.FindRiskScoreFrequencyConfigByScore(*pInfo.Score)
	if frqCfg == nil {
		return false
	}
	idealFrequency := frqCfg.IdealFrequency

	if len(pInfo.Huddles) == 0 {
		// This patient never had a huddle.  In this case, try to evenly distribute patients over n huddles
		// (where n = ideal frequency).  This isn't an exact science, but we'll use a hash and mod to do our best!
		// Example: https://play.golang.org/p/7kKKdub2OS
		modP := hash(pID) % uint32(idealFrequency)
		modH := hIdx % idealFrequency
		return modP == uint32(modH)
	}

	// Normal case in which patient does have previous huddles
	// Use >= to catch "overdue" huddles, since they are ideally ASAP
	return hIdx >= pInfo.Huddles[len(pInfo.Huddles)-1]+idealFrequency
}

func (hs *HuddleScheduler) rebalanceHuddles() error {
	targetSize := hs.getTargetHuddleSize()
	log.Info("TARGET SIZE: %d", targetSize)

	maxPull, maxPush := hs.getMaxPullAndPush()

	for i := range hs.FutureHuddles {
		h := hs.FutureHuddles[i]
		if len(h.Member) < targetSize {
			hs.growHuddle(i, targetSize, maxPull)
		} else if len(h.Member) > targetSize {
			hs.shrinkHuddle(i, targetSize, maxPush)
		}
	}
	return nil
}

func (hs *HuddleScheduler) getTargetHuddleSize() int {
	// Get the target huddle size
	totalSlots := 0
	for i := range hs.FutureHuddles {
		totalSlots += len(hs.FutureHuddles[i].Member)
	}
	return totalSlots / hs.Config.LookAhead
}

func (hs *HuddleScheduler) getMaxPullAndPush() (maxPull int, maxPush int) {
	for _, fc := range hs.Config.RiskConfig.FrequencyConfigs {
		pull := fc.IdealFrequency - fc.MinFrequency
		if pull > maxPull {
			maxPull = pull
		}
		push := fc.MaxFrequency - fc.IdealFrequency
		if push > maxPush {
			maxPush = push
		}
	}
	return
}

// growHuddle will attempt to make the huddle larger by "pulling" patients from future huddles
func (hs *HuddleScheduler) growHuddle(huddleIdx, targetSize, maxPull int) {
	h := hs.FutureHuddles[huddleIdx]

	// Determine the furthest huddle out that we can pull patients from
	maxHuddleIdx := huddleIdx + maxPull
	if maxHuddleIdx >= len(hs.FutureHuddles) {
		maxHuddleIdx = len(hs.FutureHuddles) - 1
	}

	// Iterate the next huddles to find patients to pull (until we reach maxHuddleIdx or have enough members)
	for fromHuddleIdx := huddleIdx + 1; fromHuddleIdx < maxHuddleIdx && len(h.Member) < targetSize; fromHuddleIdx++ {
		h2 := hs.FutureHuddles[fromHuddleIdx]
		// Iterate patients in this huddle to see who we can pull
		for _, mem := range h2.HuddleMembers() {
			// Only patients scheduled by risk can be pulled
			if mem.ReasonIsRiskScore() {
				pInfo := hs.patientInfos.SafeGet(mem.ID())
				if pInfo.Score != nil {
					rCfg := hs.Config.FindRiskScoreFrequencyConfigByScore(*pInfo.Score)
					// Compare spread between huddles to spread allowed by frequency config
					if (fromHuddleIdx - huddleIdx) <= (rCfg.IdealFrequency - rCfg.MinFrequency) {
						// TODO: Make sure patient isn't already in target huddle (e.g. manual)
						// We can pull this patient to the current huddle!
						hs.shiftMember(mem.ID(), fromHuddleIdx, huddleIdx)
						if len(h.Member) >= targetSize {
							break
						}
					}
				}
			}
		}
	}
}

// shrinkHuddle will attempt to make the huddle smaller by "pushing" patients to future huddles
func (hs *HuddleScheduler) shrinkHuddle(huddleIdx, targetSize, maxPush int) {
	h := hs.FutureHuddles[huddleIdx]

	// Iterate patients in this huddle, finding ones we can push ahead to next huddle
	// We want to prefer patients who can be pushed furthest, so start looking with maxPush
	for push := maxPush; push > 0 && len(h.Member) > targetSize; push-- {
		// Iterate patients in this huddle to see who we can push
		for _, mem := range h.HuddleMembers() {
			// Only patients scheduled by risk can be pushed
			if mem.ReasonIsRiskScore() {
				pInfo := hs.patientInfos.SafeGet(mem.ID())
				if pInfo.Score != nil {
					// TODO: Fix bug -- it doesn't check if the patient really can be pushed further.
					// If this patient was already pushed once, it no longer is accurate
					pushPotential := hs.getPushPotential(pInfo, huddleIdx)
					// Compare the push potential to the currently desired push
					if pushPotential >= push {
						hs.shiftMember(mem.ID(), huddleIdx, huddleIdx+1)
						if len(h.Member) <= targetSize {
							break
						}
					}
				}
			}
		}
	}
}

func (hs *HuddleScheduler) getPushPotential(pInfo *patientInfo, huddleIdx int) int {
	// If the patient doesn't have a risk score, they can't be shifted
	if pInfo.Score == nil {
		return 0
	}
	// If the patient's risk score doesn't map to a config, they can't be shifted
	frqCfg := hs.Config.FindRiskScoreFrequencyConfigByScore(*pInfo.Score)
	if frqCfg == nil {
		return 0
	}

	// Find out the last huddle index so we know if we've already reached our max distance
	lastHuddleIdx := -1
	totalHuddleIdx := hs.pastHuddleCount + huddleIdx
	for _, prvHuddleIdx := range pInfo.Huddles {
		if prvHuddleIdx >= totalHuddleIdx {
			break
		}
		lastHuddleIdx = prvHuddleIdx
	}
	if lastHuddleIdx < 0 {
		// They never had a huddle before.  We then effectively consider their last huddle
		// to have been the last huddle performed.  It's not TRUE, but it's how we handle
		// new patients or patients who received risk scores for the first time.
		lastHuddleIdx = hs.pastHuddleCount - 1
	}

	offset := huddleIdx - lastHuddleIdx
	if offset >= frqCfg.MaxFrequency {
		return 0
	}
	return frqCfg.MaxFrequency - offset
}

func (hs *HuddleScheduler) shiftMember(patientID string, fromFutureHuddleIdx, toFutureHuddleIdx int) {
	pInfo := hs.patientInfos.SafeGet(patientID)

	// First we must find the place in the patients huddle index containing the from huddleIdx
	// (This is a little confusing, since it's an index of a slice containing an index referencing another slice)
	var fromLocationInPatientHuddleSlice int
	// BTW, totalFromHuddleIdx represents the index the huddle would be if we started from the very first huddle ever
	totalFromHuddleIdx := hs.pastHuddleCount + fromFutureHuddleIdx
	for i := range pInfo.Huddles {
		if pInfo.Huddles[i] == totalFromHuddleIdx {
			fromLocationInPatientHuddleSlice = i
		}
	}

	// Now start from there and move forward, shifting the huddles appropriately
	for i := fromLocationInPatientHuddleSlice; i < len(pInfo.Huddles); i++ {
		// TODO: This changes a bit if patient is in a huddle due to rollover or manual
		pInfo.Huddles[i] = pInfo.Huddles[i] - (fromFutureHuddleIdx - toFutureHuddleIdx)
		hs.FutureHuddles[fromFutureHuddleIdx].RemoveHuddleMember(patientID)
		// Check bounds to ensure this is a huddle we are tracking before adding the patient
		if toFutureHuddleIdx >= 0 && toFutureHuddleIdx < len(hs.FutureHuddles) {
			hs.FutureHuddles[toFutureHuddleIdx].AddHuddleMemberDueToRiskScore(patientID)
		}
	}
}

func (hs *HuddleScheduler) printInfo() {
	fmt.Printf("Scheduled %d huddles with name %s\n", len(hs.FutureHuddles), hs.Config.Name)
	for i := range hs.FutureHuddles {
		fmt.Printf("\t%s: %d patients\n", getStringDate(hs.FutureHuddles[i]), len(hs.FutureHuddles[i].Member))
	}
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

func getStringDate(huddle *Huddle) string {
	dt := huddle.ActiveDateTime()
	if dt != nil {
		return dt.Time.Format("01/02/2006")
	}
	return ""
}
