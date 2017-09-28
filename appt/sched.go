package appt

import (
	"log"
	"time"
	"context"
	"errors"

	"github.com/intervention-engine/ie/app"
	"github.com/intervention-engine/ie/storage"
)

type Scheduler struct {
	service     storage.SchedService
	config      Config
	itineraries itineraryMap
}

func NewScheduler(service storage.SchedService, config Config) (*Scheduler, error) {
	s := Scheduler{service: service, config: config, itineraries: make(itineraryMap)}
	// First populate the structures we need to do the scheduling
	if err := s.itineraries.populate(service, config); err != nil {
		return nil, err
	}
	return &s, nil
}

func Boot(factory storage.ServiceFactory, settings []string) {
	configs := readConfigs(settings)
	for _, config := range configs {
		service := factory.NewSchedService()
		go func() {
			interval, err := time.ParseInLocation("3:04", config.Interval, time.Local)
			if err != nil {
				log.Println("failed to parse interval, not scheduling " + config.Name)
				return
			}
			for {
				triggerScheduler(service, config, time.Until(interval))
			}
		}()
	}
}

func triggerScheduler(service storage.SchedService, config Config, interval time.Duration) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(interval))
	defer cancel()
	select {
	// When deadline has expired, fire off Scheduler
	case <-ctx.Done():
		s, err := NewScheduler(service, config)
		if err != nil {
			log.Printf("error when trying to create scheduler with config %s: %v", config.Name, err)
			return
		}
		_, err = s.Schedule()
		if err != nil {
			log.Printf("error when trying to schedule huddles with config %s: %v", config.Name, err)
			return
		}
	}
}

func ManualSchedule(service storage.SchedService, files []string) ([]*app.Huddle, error) {
	defer service.Close()
	configs := readConfigs(files)
	log.Println("configs: ", configs)
	var scheduled []*app.Huddle
	for _, config := range configs {
		log.Println("Processing config: ", config.Name)
		log.Println(config)
		s, err := NewScheduler(service, config)
		if err != nil {
			return nil, err
		}
		huddles, err := s.Schedule()
		if err != nil {
			log.Printf("error when trying to schedule huddles with config %s: %v", config.Name, err)
			continue
		}
		scheduled = append(scheduled, huddles...)
	}
	return scheduled, nil
}

// Schedule schedules huddles based on its config.  It will schedule out the number
// of huddles as specified in the config.LookAhead.
func (s *Scheduler) Schedule() ([]*app.Huddle, error) {
	hh, err := s.planHuddles()
	if err != nil {
		return nil, err
	}
	log.Println("Planned ", len(hh), " huddles:")
	log.Println(hh)
	err = s.service.CreateHuddles(hh)
	if err != nil {
		return nil, err
	}
	// TODO: Seems to be for debug purposes
	printInfo(s.config, hh)

	return hh, nil
}

// TODO: Clear future huddles scheduled on the wrong days (unless they have manual additions)?
func (s *Scheduler) planHuddles() ([]*app.Huddle, error) {
	// Step through one day at a time, starting today, until we have created the requested number of huddles (LookAhead)
	hh := make([]*app.Huddle, 0, s.config.LookAhead)
	huddleCount := 0
	// This is only true once, and should just be evaluated on the first huddle.
	// It's for rolling patients over that haven't been reviewed
	// in a past huddle they were scheduled for and for patients that have recently
	// had an event that indicates that they should be
	// scheduled for a huddle sooner than usual.
	checkRollOversAndEvents := true
	currentDate := today()
	targetHuddleSize := s.config.getTargetHuddleSize(s.itineraries)
	log.Println("target huddle size: ", targetHuddleSize)
	careTeamID, err := s.service.FindCareTeamID(s.config.CareTeamName)
	if err != nil {
		return nil, errors.New(err.Error() + "\ncould not find care team with name: " + s.config.CareTeamName)
	}
	for len(hh) < s.config.LookAhead {
		if !s.config.IsHuddleDay(currentDate) {
			log.Println(currentDate, " is not a huddle day, continuing.")
			currentDate = currentDate.AddDate(0,0,1)
			continue
		}
		huddle, err := s.service.FindCareTeamHuddleOnDate(careTeamID, currentDate)
		if err != nil && err.Error() != "not found" {
			return nil, err
		}
		if huddle == nil {
			hd := currentDate
			huddle = &app.Huddle{CareTeamID: &careTeamID, Date: &hd}
		}
		// If this is today's huddle and any patients are marked reviewed already, then do NOT reschedule this huddle!
		if inProgress(huddle) {
			// Need to update the itinerary's last huddle and add the huddle to our slice of huddles
			s.itineraries.updatePatientsLastHuddle(huddle, s.config, huddleCount)
			hh = append(hh, huddle)
			huddleCount++
			// Go to the next huddle
			continue
		}

		// Remember the original members
		origPatients := huddle.Patients
		// Clear the original patients so we start from clean slate
		huddle.Patients = nil
		// Add back the manually added and rolled over patients
		for _, patient := range origPatients {
			if (patient.Reason != nil) && (*patient.Reason == "MANUAL_ADDITION" || *patient.Reason == "ROLLOVER") {
				reasonType := ""
				if patient.ReasonType != nil {
					reasonType = *patient.ReasonType
				}
				addPatient(huddle, *patient.ID, *patient.Reason, reasonType)
			}
		}

		if checkRollOversAndEvents {
			// Add new rollovers if applicable
			if s.config.RollOverDelayInDays > 0 {
				// Find the patients that need to roll over (i.e., the ones not reviewed in the huddle x days ago)
				expiredDate := today().AddDate(0, 0, -1*s.config.RollOverDelayInDays)
				expiredHuddle, err := s.service.FindCareTeamHuddleOnDate(careTeamID, expiredDate)
				if err != nil && err.Error() != "not found" {
					log.Printf("Error searching on previous huddle (%s) to detect rollover patients\n", expiredDate.Format("Jan 2"))
					log.Println(err.Error())
				}
				if err == nil && expiredHuddle != nil {
					addPatientsBasedOnRollOvers(huddle, expiredHuddle)
				}
			}
			// TODO: pick up here
			// Add members to the huddle who had a recent encounter that triggers huddle discussion
			if s.config.Event != nil && len(s.config.Event.EncounterConfigs) > 0 {
				// Loop through the event configs, looking for patients with matching encounters
				for _, e := range s.config.Event.EncounterConfigs {
					earliestDate := e.EarliestLookBackDate(*huddle.Date)
					latestDate := e.LatestLookBackDate(*huddle.Date)
					codeVals := make([]string, len(e.TypeCodes))
					if len(e.TypeCodes) > 0 {
						for i, code := range e.TypeCodes {
							codeVals[i] = code.System + "|" + code.Code
						}
					}
					encounters, err := s.service.FindEncounters(codeVals, earliestDate, latestDate)
					if err != nil {
						// TODO: improve this error message, possibly add exception for if "not found"
						log.Println("error trying to find encounters")
						continue
					}
					addPatientsBasedOnRecentEncounters(huddle, hh, encounters, e.TypeCodes, earliestDate, latestDate)
				}
			}
			checkRollOversAndEvents = false
		}

		// Finally fill out the rest with members based on their risk scores (which determine huddle frequency)
		if s.config.Risk != nil && len(s.config.Risk.Frequencies) > 0 {
			addPatientsBasedOnRiskScores(huddle, huddleCount, targetHuddleSize, s.itineraries)
		}

		// Now go through all of the assigned huddle members and update those itineraries
		s.itineraries.updatePatientsLastHuddle(huddle, s.config, huddleCount)

		hh = append(hh, huddle)
		huddleCount++
		currentDate = currentDate.AddDate(0, 0, 1)
	}
	return hh, nil
}

func inProgress(huddle *app.Huddle) bool {
	if huddle.Date != nil && isToday(*huddle.Date) {
		for _, patient := range huddle.Patients {
			if patient.Reviewed != nil && *patient.Reviewed {
				return true
			}
		}
	}
	return false
}

// TODO: will have to let in what today is for testing
func isToday(date time.Time) bool {
	today := time.Now().In(time.Local)
	return (date.Day() == today.Day() && date.Month() == today.Month() && date.Year() == today.Year())
}

func addPatient(huddle *app.Huddle, patientID, reason, reasonType string) {
	// First look to see if the patient is already in the huddle and act accordingly.
	existing := findPatient(huddle, patientID)
	if existing != nil {
		if existing.Reason != nil && *existing.Reason == "ROLLOVER" {
			// We allow overwrites of rollovers, so remove the existing entry and continue
			removePatient(huddle, patientID)
		} else {
			// We don't allow overwrites of other reasons, so just ignore this request
			return
		}
	}

	// The patient is not yet in the group, so add him/her
	huddle.Patients = append(huddle.Patients, &app.HuddlePatient{ID: &patientID, Reason: &reason, ReasonType: &reasonType})
}

func findPatient(huddle *app.Huddle, patientID string) *app.HuddlePatient {
	for i := range huddle.Patients {
		if *huddle.Patients[i].ID == patientID {
			return huddle.Patients[i]
		}
	}
	return nil
}

func removePatient(huddle *app.Huddle, patientID string) {
	for i := range huddle.Patients {
		if *huddle.Patients[i].ID == patientID {
			if i == (len(huddle.Patients)-1) {
				huddle.Patients = huddle.Patients[:i]
				break
			}
			huddle.Patients = append(huddle.Patients[:i], huddle.Patients[i+1:]...)
			break
		}
	}
}

func printInfo(config Config, huddles []*app.Huddle) {
	log.Printf("Scheduled %d huddles with name %s\n", len(huddles), config.Name)
	for _, h := range huddles {
		log.Printf("\t%s: %d patients\n", getStringDate(h), len(h.Patients))
	}
}

var _nowValueForTestingOnly *time.Time

func now() time.Time {
	if _nowValueForTestingOnly != nil {
		return *_nowValueForTestingOnly
	}
	return time.Now()
}
func today() time.Time {
	now := now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func getStringDate(huddle *app.Huddle) string {
	if huddle.Date != nil {
		return huddle.Date.Format("01/02/2006")
	}
	return ""
}
