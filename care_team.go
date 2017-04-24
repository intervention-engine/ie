package main

import (
	"github.com/goadesign/goa"
	"github.com/intervention-engine/ie/app"
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
	s := GetStorageService(ctx.Context)
	cs := s.NewCareTeamService()
	var ct app.Careteam
	ct.Leader = &ctx.Payload.Leader
	ct.Name = &ctx.Payload.Name
	err := cs.CreateCareTeam(&ct)
	if err != nil {
		return goa.ErrInternal("internal server error trying to create care team")
	}

	return ctx.OK(&ct)
}

// Delete runs the delete action.
func (c *CareTeamController) Delete(ctx *app.DeleteCareTeamContext) error {
	s := GetStorageService(ctx.Context)
	cs := s.NewCareTeamService()
	err := cs.DeleteCareTeam(ctx.ID)
	if err != nil {
		if err.Error() == "bad id" {
			return goa.ErrBadRequest("id was not a proper bson object id")
		}
		if err.Error() == "not found" {
			return goa.ErrNotFound("could not find patient")
		}
		return goa.ErrInternal("error trying to delete patient")
	}
	return ctx.OK(nil)
}

// List runs the list action.
func (c *CareTeamController) List(ctx *app.ListCareTeamContext) error {
	s := GetStorageService(ctx.Context)
	cs := s.NewCareTeamService()
	cc, err := cs.CareTeams()
	if err != nil {
		return goa.ErrInternal("error trying to list patients")
	}

	return ctx.OK(cc)
}

// Show runs the show action.
func (c *CareTeamController) Show(ctx *app.ShowCareTeamContext) error {
	s := GetStorageService(ctx.Context)
	cs := s.NewCareTeamService()
	ct, err := cs.CareTeam(ctx.ID)
	if err != nil {
		if err.Error() == "bad id" {
			return goa.ErrBadRequest("id was not a proper bson object id")
		}
		if err.Error() == "not found" {
			return goa.ErrNotFound("could not find care team")
		}
		return goa.ErrInternal("internal server error trying to find care team")
	}

	return ctx.OK(ct)
}

// Update runs the update action.
func (c *CareTeamController) Update(ctx *app.UpdateCareTeamContext) error {
	s := GetStorageService(ctx.Context)
	cs := s.NewCareTeamService()

	var ct app.Careteam
	ct.ID = ctx.Payload.ID
	ct.Name = &ctx.Payload.Name
	ct.Leader = &ctx.Payload.Leader
	ct.CreatedAt = ctx.Payload.CreatedAt
	err := cs.UpdateCareTeam(&ct)
	if err != nil {
		if err.Error() == "bad id" {
			return goa.ErrBadRequest("id was not a proper bson object id")
		}
		if err.Error() == "not found" {
			return goa.ErrNotFound("could not find care team")
		}
		return goa.ErrInternal("error trying to update care team")
	}
	return ctx.OK(nil)
}
