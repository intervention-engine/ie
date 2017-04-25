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
	ctx.ResponseData.Header().Set("Location", careteamHref(*ct.ID))

	return ctx.Created()
}

// Delete runs the delete action.
func (c *CareTeamController) Delete(ctx *app.DeleteCareTeamContext) error {
	s := GetStorageService(ctx.Context)
	cs := s.NewCareTeamService()
	err := cs.DeleteCareTeam(ctx.ID)
	if err != nil {
		if err.Error() == "bad id" {
			// return goa.ErrBadRequest("id was not a proper bson object id")
			return ctx.BadRequest()
		}
		if err.Error() == "not found" {
			// return goa.ErrNotFound("could not find resource")
			return ctx.NotFound()
		}
		// return goa.ErrInternal("internal server error trying to fulfill request")
		return ctx.InternalServerError()
	}
	return ctx.NoContent()
}

// List runs the list action.
func (c *CareTeamController) List(ctx *app.ListCareTeamContext) error {
	s := GetStorageService(ctx.Context)
	cs := s.NewCareTeamService()
	cc, err := cs.CareTeams()
	if err != nil {
		// return goa.ErrInternal("error trying to list patients")
		return ctx.Err()
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
			// return goa.ErrBadRequest("id was not a proper bson object id")
			return ctx.BadRequest()
		}
		if err.Error() == "not found" {
			// return goa.ErrNotFound("could not find resource")
			return ctx.NotFound()
		}
		// return goa.ErrInternal("internal server error trying to fulfill request")
		return ctx.InternalServerError()
	}

	return ctx.OK(ct)
}

// Update runs the update action.
func (c *CareTeamController) Update(ctx *app.UpdateCareTeamContext) error {
	s := GetStorageService(ctx.Context)
	cs := s.NewCareTeamService()

	var ct app.Careteam
	ct.ID = &ctx.Payload.ID
	ct.Name = &ctx.Payload.Name
	ct.Leader = &ctx.Payload.Leader
	ct.CreatedAt = &ctx.Payload.CreatedAt
	err := cs.UpdateCareTeam(&ct)
	if err != nil {
		if err.Error() == "bad id" {
			// return goa.ErrBadRequest("id was not a proper bson object id")
			return ctx.BadRequest()
		}
		if err.Error() == "not found" {
			// return goa.ErrNotFound("could not find resource")
			return ctx.NotFound()
		}
		// return goa.ErrInternal("internal server error trying to fulfill request")
		return ctx.InternalServerError()
	}
	return ctx.NoContent()
}

func careteamHref(ID string) string {
	return "/care_teams/" + ID
}
