package main

import (
	"log"
	"time"

	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/app"
	"github.com/intervention-engine/ie/storage"
)

// CareTeamController implements the care_team resource.
type CareTeamController struct {
	*goa.Controller
}

// NewCareTeamController creates a care_team controller.
func NewCareTeamController(service *goa.Service) *CareTeamController {
	return &CareTeamController{Controller: service.NewController("CareTeamController")}
}

// Create runs the create action.
func (c *CareTeamController) Create(ctx *app.CreateCareTeamContext) error {
	s := GetServiceFactory(ctx.Context)
	cs := s.NewCareTeamService()
	var ct app.CareTeam
	ct.Leader = &ctx.Payload.Leader
	ct.Name = &ctx.Payload.Name
	err := cs.CreateCareTeam(&ct)
	if err != nil {
		return ctx.InternalServerError(goa.ErrInternal("error trying to create care team", "error", err))
	}
	if ct.ID == nil {
		return ctx.InternalServerError(goa.ErrInternal("service failed to make id for care team"))
	}
	ctx.ResponseData.Header().Set("Location", app.CareTeamHref(*ct.ID))
	return ctx.Created()
}

// Delete runs the delete action.
func (c *CareTeamController) Delete(ctx *app.DeleteCareTeamContext) error {
	s := GetServiceFactory(ctx.Context)
	cs := s.NewCareTeamService()
	err := cs.DeleteCareTeam(ctx.ID)
	if err != nil {
		if err.Error() == "bad id" {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		if err.Error() == "not found" {
			return ctx.NotFound(err)
		}
		return ctx.InternalServerError(goa.ErrInternal("error trying to fulfill request", "error", err))
	}
	return ctx.NoContent()
}

// List runs the list action.
func (c *CareTeamController) List(ctx *app.ListCareTeamContext) error {
	s := GetServiceFactory(ctx.Context)
	cs := s.NewCareTeamService()
	cc, err := cs.CareTeams()
	if err != nil {
		return ctx.InternalServerError(goa.ErrInternal("error trying to list care teams", "error", err))
	}
	return ctx.OK(cc)
}

// Show runs the show action.
func (c *CareTeamController) Show(ctx *app.ShowCareTeamContext) error {
	s := GetServiceFactory(ctx.Context)
	cs := s.NewCareTeamService()
	log.Println("ID received is: " + ctx.ID)
	ct, err := cs.CareTeam(ctx.ID)
	if err != nil {
		if err.Error() == "bad id" {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		if err.Error() == "not found" {
			return ctx.NotFound(err)
		}
		return ctx.InternalServerError(goa.ErrInternal("error trying to fulfill request", "error", err))
	}
	return ctx.OK(ct)
}

// Update runs the update action.
func (c *CareTeamController) Update(ctx *app.UpdateCareTeamContext) error {
	s := GetServiceFactory(ctx.Context)
	cs := s.NewCareTeamService()

	var validupdate bool
	var ct app.CareTeam
	ct.ID = &ctx.ID
	if ctx.Payload.Name != nil {
		validupdate = true
		ct.Name = ctx.Payload.Name
	}
	if ctx.Payload.Leader != nil {
		validupdate = true
		ct.Leader = ctx.Payload.Leader
	}
	if !validupdate {
		return ctx.BadRequest(goa.ErrBadRequest("content for update is not valid"))
	}
	err := cs.UpdateCareTeam(&ct)
	if err != nil {
		if err.Error() == "bad id" {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		if err.Error() == "not found" {
			return ctx.NotFound(err)
		}
		return ctx.InternalServerError(goa.ErrInternal("error trying to fulfill request", "error", err))
	}
	return ctx.NoContent()
}

func (c *CareTeamController) AddPatient(ctx *app.AddPatientCareTeamContext) error {
	s := GetServiceFactory(ctx.Context)
	cs := s.NewCareTeamService()
	err := cs.AddPatient(ctx.ID, ctx.PatientID)
	if err != nil {
		if (err.Error() == "bad patient id") || (err.Error() == "bad care team id") || (err.Error() == "patient already belongs to care team") {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		if (err.Error() == "patient not found") || (err.Error() == "care team not found") {
			return ctx.NotFound(goa.ErrNotFound(err))
		}
		log.Println("error after trying to add patient", err)
		return ctx.InternalServerError(goa.ErrInternal("error adding patient", "error", err))
	}
	return ctx.NoContent()
}

// RemovePatient Removes a patient from a care team and removes them from all future huddles
// with that care team, starting with the day after they are removed from the care team.
func (c *CareTeamController) RemovePatient(ctx *app.RemovePatientCareTeamContext) error {
	s := GetServiceFactory(ctx.Context)
	cs := s.NewCareTeamService()
	err := cs.RemovePatient(ctx.ID, ctx.PatientID)
	if err != nil {
		if (err.Error() == "bad patient id") || (err.Error() == "bad care team id") {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		if (err.Error() == "patient not found") || (err.Error() == "care team not found") || (err.Error() == "membership does not exist") {
			return ctx.NotFound(err)
		}
		return ctx.InternalServerError(goa.ErrInternal("error removing patient", "error", err))
	}
	return ctx.NoContent()
}

// Huddles runs the huddles action.
func (c *CareTeamController) Huddles(ctx *app.HuddlesCareTeamContext) error {
	s := GetServiceFactory(ctx.Context)
	hs := s.NewHuddleService()
	var query storage.HuddleFilterQuery
	if ctx.PatientID != nil {
		query.PatientID = *ctx.PatientID
	}
	if ctx.Date != nil {
		// validate that date fits our format
		date, err := time.Parse("2006-01-02", *ctx.Date)
		if err != nil {
			return ctx.BadRequest(goa.ErrBadRequest(err.Error()))
		}
		query.Date = date
	}
	var hh []*app.Huddle
	query.CareTeamID = ctx.ID
	hh, err := hs.HuddlesFilterBy(query)
	if err != nil {
		if err.Error() == "bad care team id" {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		if err.Error() == "bad patient id" {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		if err.Error() == "not found" {
			return ctx.NotFound(goa.ErrNotFound("could not find huddles for that care team"))
		}
		return ctx.InternalServerError(goa.ErrInternal("error trying to list huddles for care team", "error", err))
	}
	return ctx.OK(hh)
}

// SchedHuddle runs the sched_huddle action.
func (c *CareTeamController) Schedule(ctx *app.ScheduleCareTeamContext) error {
	s := GetServiceFactory(ctx.Context)
	hs := s.NewHuddleService()
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return ctx.InternalServerError(goa.ErrInternal("could not load local time zone", "error", err))
	}
	dateTime, err := time.ParseInLocation("2006-01-02", ctx.Payload.Date, loc)
	if err != nil {
		return ctx.BadRequest(goa.ErrBadRequest("could not parse date, make sure in YYYY-MM-dd format", "error", err))
	}
	now := time.Now()
	if dateTime.Before(now) {
		return ctx.BadRequest(goa.ErrBadRequest("date given is in the past, cannot schedule huddle"))
	}
	h, created, err := hs.ScheduleHuddle(ctx.ID, ctx.Payload.PatientID, dateTime)
	if err != nil {
		if (err.Error() == "bad patient id") || (err.Error() == "bad care team id") {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		if (err.Error() == "patient not found") || (err.Error() == "care team not found") || (err.Error() == "membership does not exist") {
			return ctx.NotFound(err)
		}
		return ctx.InternalServerError(goa.ErrInternal("error scheduling patient", "error", err))
	}
	if created {
		return ctx.Created(h)
	}
	return ctx.OK(h)
}
