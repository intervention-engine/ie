package main

import (
	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/app"
	"github.com/intervention-engine/ie/appt"
)

// HuddleController implements the care_team resource.
type HuddleController struct {
	*goa.Controller
}

// NewHuddleController creates a huddle controller.
func NewHuddleController(service *goa.Service) *HuddleController {
	return &HuddleController{Controller: service.NewController("HuddleController")}
}

func (c *HuddleController) Cancel(ctx *app.CancelHuddleContext) error {
	s := GetServiceFactory(ctx.Context)
	hs := s.NewHuddleService()
	h, err := hs.DeletePatient(ctx.ID, ctx.PatientID)
	if err != nil {
		if (err.Error() == "bad huddle id") || (err.Error() == "bad patient id") {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		if (err.Error() == "huddle not found") || (err.Error() == "patient not found") {
			return ctx.NotFound(goa.ErrNotFound(err))
		}
		return ctx.InternalServerError(goa.ErrInternal("error cancelling patient", "error", err))
	}
	if h == nil {
		return ctx.NoContent()
	}
	return ctx.OK(h)
}

func (c *HuddleController) BatchSchedule(ctx *app.BatchScheduleHuddleContext) error {
	s := GetServiceFactory(ctx.Context)
	files := GetConfigFiles(ctx.Context)
	svc := s.NewSchedService()
	hh, err := appt.ManualSchedule(svc, files)
	if err != nil {
		return ctx.InternalServerError(goa.ErrInternal("error scheduling huddles", "error", err))
	}
	if hh == nil {
		// no huddles were scheduled
		return ctx.NoContent()
	}
	return ctx.OK(hh)
}
