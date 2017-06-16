package main

import (
	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/app"
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
